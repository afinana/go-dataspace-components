package ports_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"go.opentelemetry.io/otel/trace/noop"

	"github.com/afinana/go-dataspace-components/consumer/ports"
)

func TestConsumerManagementHandler_InvalidJSON(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	tracer := noop.NewTracerProvider().Tracer("test")

	mgmtHandler := ports.NewConsumerManagementHandler(logger, tracer, nil, nil, nil, "http://provider", "http://callback")

	req := httptest.NewRequest(http.MethodPost, "/api/consumer/v4/contractnegotiations", bytes.NewBufferString("{invalid-json"))
	rec := httptest.NewRecorder()

	mgmtHandler.HandleInitiateNegotiation(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request, got %d", rec.Code)
	}

	var errResp map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}
	if errResp["error"] != "invalid request" {
		t.Errorf("expected error 'invalid request', got '%s'", errResp["error"])
	}
}

func TestConsumerManagementHandler_HandleCatalogRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{
			"@type": "dcat:Catalog",
			"id":    "catalog-test",
		})
	}))
	defer server.Close()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	tracer := noop.NewTracerProvider().Tracer("test")
	dspClient := ports.NewDSPClient(logger, tracer)

	mgmtHandler := ports.NewConsumerManagementHandler(logger, tracer, nil, nil, dspClient, server.URL, "http://callback")

	req := httptest.NewRequest(http.MethodPost, "/api/consumer/v4/catalog/request", nil)
	rec := httptest.NewRecorder()

	mgmtHandler.HandleCatalogRequest(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", rec.Code)
	}
}
