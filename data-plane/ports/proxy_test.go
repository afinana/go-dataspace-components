package ports_test

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	cp "github.com/afinana/go-dataspace-components/control-plane/domain"
	dp "github.com/afinana/go-dataspace-components/data-plane/domain"
	"github.com/afinana/go-dataspace-components/data-plane/ports"
)

func TestAPIProxyController_InitiateAndProxy(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))

	// Mock backend server
	backendServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("X-API-KEY")
		if apiKey != "backend-secret-xyz" {
			http.Error(w, "Unauthorized backend access", http.StatusForbidden)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Backend payload response"))
	}))
	defer backendServer.Close()

	controller := ports.NewAPIProxyController(logger, nil)

	req := &dp.DataFlowRequest{
		ID: "flow-proxy-test",
		SourceDataAddress: cp.DataAddress{
			Type: "HttpData",
			Properties: map[string]string{
				"endpoint":      backendServer.URL,
				"authType":      "custom",
				"authHeaderKey": "X-API-KEY",
				"authSecret":    "backend-secret-xyz",
			},
		},
		DestinationDataAddress: cp.DataAddress{
			Type: "HttpProxy",
		},
		Properties: map[string]string{
			"auth_token": "valid-proxy-token-123",
		},
	}

	if !controller.CanHandle(req) {
		t.Fatal("expected controller.CanHandle to return true for HttpData -> HttpProxy")
	}

	resp, err := controller.Initiate(context.Background(), req)
	if err != nil {
		t.Fatalf("failed to initiate proxy flow: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected successful response, got error: %s", resp.ErrorDetail)
	}

	// Test proxying request with authorization token
	proxyReq := httptest.NewRequest(http.MethodGet, "/public/data/items", nil)
	proxyReq.Header.Set("Authorization", "Bearer valid-proxy-token-123")
	rec := httptest.NewRecorder()

	controller.ServeHTTP(rec, proxyReq)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status OK 200 from proxied request, got %d. Body: %s", rec.Code, rec.Body.String())
	}
	if rec.Body.String() != "Backend payload response" {
		t.Errorf("unexpected body from backend proxy: %s", rec.Body.String())
	}
}

func TestAPIProxyController_UnauthorizedToken(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	controller := ports.NewAPIProxyController(logger, nil)

	proxyReq := httptest.NewRequest(http.MethodGet, "/public/data/items", nil)
	proxyReq.Header.Set("Authorization", "Bearer invalid-token-999")
	rec := httptest.NewRecorder()

	controller.ServeHTTP(rec, proxyReq)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status Unauthorized 401 for invalid token, got %d", rec.Code)
	}
}

func TestAPIProxyController_HandleFlowsListAndDetail(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	controller := ports.NewAPIProxyController(logger, nil)

	// Test List Flows
	listReq := httptest.NewRequest(http.MethodGet, "/api/proxy/flows", nil)
	listRec := httptest.NewRecorder()
	controller.HandleFlowsList(listRec, listReq)

	if listRec.Code != http.StatusOK {
		t.Errorf("expected status OK 200 for flows list, got %d", listRec.Code)
	}

	// Test Detail Flow (default pre-populated flow)
	detailReq := httptest.NewRequest(http.MethodGet, "/api/proxy/flows/flow-process-test", nil)
	detailReq.SetPathValue("flowId", "flow-process-test")
	detailRec := httptest.NewRecorder()
	controller.HandleFlowsDetail(detailRec, detailReq)

	if detailRec.Code != http.StatusOK {
		t.Errorf("expected status OK 200 for flow detail, got %d", detailRec.Code)
	}

	var detailBody map[string]any
	if err := json.Unmarshal(detailRec.Body.Bytes(), &detailBody); err != nil {
		t.Fatalf("failed to unmarshal detail body: %v", err)
	}
	if _, exists := detailBody["endpointProperties"]; !exists {
		t.Error("expected 'endpointProperties' field in flow detail response")
	}
}
