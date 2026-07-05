package ports

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"time"

	dp "github.com/afinana/go-dataspace-components/data-plane/domain"
)

// APIProxyController implements dp.DataFlowController to reverse-proxy HTTP REST APIs.
// It manages active proxy mappings securely by using temporary authorization tokens.
type APIProxyController struct {
	mu           sync.RWMutex
	logger       *slog.Logger
	// activeFlows maps token -> DataFlowRequest containing backend source endpoint and auth credentials
	activeFlows  map[string]*dp.DataFlowRequest
	reverseProxy *httputil.ReverseProxy
}

// NewAPIProxyController initializes the API reverse proxy.
func NewAPIProxyController(logger *slog.Logger) *APIProxyController {
	controller := &APIProxyController{
		logger:      logger,
		activeFlows: make(map[string]*dp.DataFlowRequest),
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

	// In a real EDC connector, this would be a secure signed token (e.g. JWT) containing flow state.
	// For simplicity in this scaffold, we generate a secure handle.
	token := req.Properties["auth_token"]
	if token == "" {
		token = fmt.Sprintf("edr-%s-%d", req.ID, time.Now().UnixNano())
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
	if authHeader == "" {
		c.logger.Warn("Unauthorized access attempt: missing Authorization header")
		http.Error(w, "Missing Authorization Header", http.StatusUnauthorized)
		return
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	
	// 2. Validate token against active registered data flows
	c.mu.RLock()
	flowRequest, exists := c.activeFlows[token]
	c.mu.RUnlock()

	if !exists {
		c.logger.Warn("Unauthorized access attempt: invalid or expired proxy token", "token", token)
		http.Error(w, "Invalid or Expired Authorization Token", http.StatusForbidden)
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

	// Strip the routing prefix "/public" if present
	relPath := req.URL.Path
	if strings.HasPrefix(relPath, "/public") {
		relPath = strings.TrimPrefix(relPath, "/public")
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
