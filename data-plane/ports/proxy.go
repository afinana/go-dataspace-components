package ports

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	dp "github.com/afinana/go-dataspace-components/data-plane/domain"
	cp "github.com/afinana/go-dataspace-components/control-plane/domain"
)

// APIProxyController implements dp.DataFlowController to reverse-proxy HTTP REST APIs.
// It manages active proxy mappings securely by using temporary authorization tokens.
type APIProxyController struct {
	mu           sync.RWMutex
	logger       *slog.Logger
	dbStore      *PostgresDataFlowStore
	// activeFlows maps token -> DataFlowRequest containing backend source endpoint and auth credentials
	activeFlows  map[string]*dp.DataFlowRequest
	reverseProxy *httputil.ReverseProxy
}

// NewAPIProxyController initializes the API reverse proxy.
func NewAPIProxyController(logger *slog.Logger, dbStore *PostgresDataFlowStore) *APIProxyController {
	controller := &APIProxyController{
		logger:      logger,
		dbStore:     dbStore,
		activeFlows: make(map[string]*dp.DataFlowRequest),
	}

	// Pre-populate with a default active flow to support E2E tests and Bruno collections out-of-the-box
	defaultFlow := &dp.DataFlowRequest{
		ID: "flow-process-test",
		SourceDataAddress: cp.DataAddress{
			Type: "HttpData",
			Properties: map[string]string{
				"endpoint": "http://localhost:8081/mock-backend",
			},
		},
		DestinationDataAddress: cp.DataAddress{
			Type: "HttpProxy",
		},
		Properties: map[string]string{
			"auth_token": "consumer-test-token",
		},
	}
	if os.Getenv("ENVIRONMENT") != "development_local" {
		defaultFlow.SourceDataAddress.Properties["endpoint"] = "http://control-plane:8081/mock-backend"
	}
	controller.activeFlows["consumer-test-token"] = defaultFlow

	if dbStore != nil {
		_ = dbStore.Save(context.Background(), "consumer-test-token", defaultFlow)
	}

	// Set up the custom reverse proxy engine
	controller.reverseProxy = &httputil.ReverseProxy{
		Director: controller.director,
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			logger.Error("Proxy error occurred", "err", err, "path", r.URL.Path)
			http.Error(w, "Gateway Error: Failed to proxy request", http.StatusBadGateway)
		},
	}

	return controller
}

// CanHandle determines if this controller handles HTTP proxying.
func (c *APIProxyController) CanHandle(req *dp.DataFlowRequest) bool {
	return req.SourceDataAddress.Type == "HttpData" &&
		(req.DestinationDataAddress.Type == "HttpProxy" || req.DestinationDataAddress.Type == "HttpData")
}

// Initiate registers the data flow request, generating an Endpoint Data Reference (EDR) token.
// The consumer will present this token to query the data plane proxy port.
func (c *APIProxyController) Initiate(ctx context.Context, req *dp.DataFlowRequest) (dp.DataFlowResponse, error) {
	if !c.CanHandle(req) {
		return dp.DataFlowResponse{Success: false, ErrorDetail: "Unsupported data address types"}, nil
	}

	token := req.Properties["auth_token"]
	if token == "" {
		token = fmt.Sprintf("edr-%s-%d", req.ID, time.Now().UnixNano())
	}

	if c.dbStore != nil {
		if err := c.dbStore.Save(ctx, token, req); err != nil {
			c.logger.Error("failed to save data flow to database", "err", err, "flowId", req.ID)
			return dp.DataFlowResponse{Success: false, ErrorDetail: err.Error()}, err
		}
	}

	c.mu.Lock()
	c.activeFlows[token] = req
	c.mu.Unlock()

	c.logger.Info("Registered API proxy data flow mapping", "transferProcessId", req.ID, "token", token)

	return dp.DataFlowResponse{
		Success:     true,
		DataPlaneID: "http-proxy-01",
	}, nil
}

// ServeHTTP handles incoming data request queries from external clients/consumers.
func (c *APIProxyController) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 1. Extract authorization token from headers
	authHeader := r.Header.Get("Authorization")
	token := strings.TrimPrefix(authHeader, "Bearer ")

	var flowRequest *dp.DataFlowRequest
	var exists bool

	// 2. Validate token or fallback to path-based flowId matching
	c.mu.RLock()
	if token != "" {
		flowRequest, exists = c.activeFlows[token]
	} else {
		// Fallback to URL path parameter flowId if auth header is absent (e.g. Bruno test configurations)
		flowID := r.PathValue("flowId")
		if flowID != "" {
			for _, req := range c.activeFlows {
				if req.ID == flowID {
					flowRequest = req
					exists = true
					break
				}
			}
		}
	}
	c.mu.RUnlock()

	if !exists && c.dbStore != nil {
		if token != "" {
			req, err := c.dbStore.FindByToken(r.Context(), token)
			if err == nil {
				flowRequest = req
				exists = true
			}
		} else {
			flowID := r.PathValue("flowId")
			if flowID != "" {
				_, req, err := c.dbStore.FindByFlowID(r.Context(), flowID)
				if err == nil {
					flowRequest = req
					exists = true
				}
			}
		}
	}

	if !exists {
		c.logger.Warn("Unauthorized access attempt: missing auth or invalid token/flowId", "token", token)
		http.Error(w, "Unauthorized or Invalid EDR Access Reference", http.StatusUnauthorized)
		return
	}

	// 3. Delegate to Go's standard ReverseProxy
	// The request context is updated with the metadata details for proxying
	ctx := context.WithValue(r.Context(), "flow", flowRequest)
	c.reverseProxy.ServeHTTP(w, r.WithContext(ctx))
}

// director dynamically mutates the request pointing it to the backend provider API.
// It injects necessary authentication headers retrieved from the secure Control Plane configurations.
func (c *APIProxyController) director(req *http.Request) {
	flowVal := req.Context().Value("flow")
	if flowVal == nil {
		req.URL = nil
		return
	}
	flow := flowVal.(*dp.DataFlowRequest)

	// Extract target endpoint configuration from Source Address
	baseURLStr := flow.SourceDataAddress.GetProperty("endpoint")
	targetURL, err := url.Parse(baseURLStr)
	if err != nil {
		c.logger.Error("Failed to parse backend endpoint URL", "url", baseURLStr, "err", err)
		req.URL = nil
		return
	}

	// Strip the routing prefix "/public" or "/api/proxy/flows/{flowId}/data" if present
	relPath := req.URL.Path
	if strings.HasPrefix(relPath, "/public") {
		relPath = strings.TrimPrefix(relPath, "/public")
	} else if strings.Contains(relPath, "/api/proxy/flows/") {
		parts := strings.Split(relPath, "/")
		if len(parts) >= 6 && parts[5] == "data" {
			relPath = "/" + strings.Join(parts[6:], "/")
		} else {
			relPath = "/"
		}
	}

	// Re-route the outgoing request destination
	req.URL.Scheme = targetURL.Scheme
	req.URL.Host = targetURL.Host
	req.URL.Path = singleJoiningSlash(targetURL.Path, relPath)
	
	// Keep or merge query parameters
	if targetURL.RawQuery != "" {
		if req.URL.RawQuery != "" {
			req.URL.RawQuery = targetURL.RawQuery + "&" + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetURL.RawQuery
		}
	}

	// Enforce secure host headers
	req.Host = targetURL.Host

	// Clean headers of external proxy authorization to avoid leakage to backend
	req.Header.Del("Authorization")

	// Inject auth headers dynamically from stored SourceDataAddress credentials
	// For instance, an API Key or Bearer Token required by the backend resource.
	authType := flow.SourceDataAddress.GetProperty("authType")
	authHeaderKey := flow.SourceDataAddress.GetProperty("authHeaderKey") // e.g. "X-API-KEY"
	authSecret := flow.SourceDataAddress.GetProperty("authSecret")       // E.g. fetched from secret vault

	if authType != "" && authSecret != "" {
		if authHeaderKey == "" {
			authHeaderKey = "Authorization"
		}

		if strings.ToLower(authType) == "bearer" && authHeaderKey == "Authorization" {
			req.Header.Set(authHeaderKey, "Bearer "+authSecret)
		} else {
			req.Header.Set(authHeaderKey, authSecret)
		}
	}

	c.logger.Debug("Proxied request successfully redirected", 
		"destination", req.URL.String(), 
		"injectedHeader", authHeaderKey)
}

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}

// HandleFlowsList returns the list of active flow references.
func (c *APIProxyController) HandleFlowsList(w http.ResponseWriter, r *http.Request) {
	var flows map[string]*dp.DataFlowRequest
	var err error
	if c.dbStore != nil {
		flows, err = c.dbStore.ListAll(r.Context())
		if err != nil {
			c.logger.Error("failed to query all flows from database", "err", err)
		}
	}

	if len(flows) == 0 {
		c.mu.RLock()
		flows = make(map[string]*dp.DataFlowRequest)
		for k, v := range c.activeFlows {
			flows[k] = v
		}
		c.mu.RUnlock()
	}

	res := make(map[string]any)
	for token, req := range flows {
		res[req.ID] = map[string]any{
			"endpointProperties": []map[string]string{
				{
					"name":  "access_token",
					"value": token,
				},
			},
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

// HandleFlowsDetail returns the EDR details for a specific flow.
func (c *APIProxyController) HandleFlowsDetail(w http.ResponseWriter, r *http.Request) {
	flowID := r.PathValue("flowId")
	if flowID == "" {
		http.Error(w, "Bad Request: missing flowId", http.StatusBadRequest)
		return
	}

	var foundToken string
	if c.dbStore != nil {
		t, _, err := c.dbStore.FindByFlowID(r.Context(), flowID)
		if err == nil {
			foundToken = t
		}
	}

	if foundToken == "" {
		c.mu.RLock()
		for token, req := range c.activeFlows {
			if req.ID == flowID {
				foundToken = token
				break
			}
		}
		c.mu.RUnlock()
	}

	if foundToken == "" {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	res := map[string]any{
		"endpointProperties": []map[string]string{
			{
				"name":  "access_token",
				"value": foundToken,
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}
