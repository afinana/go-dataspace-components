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

func TestConsumerCallbackHandler_HandleContractOffer_InvalidJSON(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	tracer := noop.NewTracerProvider().Tracer("test")

	handler := ports.NewConsumerCallbackHandler(logger, tracer, nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/consumer/negotiations/offers", bytes.NewBufferString("{invalid"))
	rec := httptest.NewRecorder()

	handler.HandleContractOffer(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request, got %d", rec.Code)
	}

	var errResp map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}
	if errResp["error"] != "invalid request body" {
		t.Errorf("expected 'invalid request body', got '%s'", errResp["error"])
	}
}

func TestConsumerCallbackHandler_HandleContractAgreement_InvalidJSON(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	tracer := noop.NewTracerProvider().Tracer("test")

	handler := ports.NewConsumerCallbackHandler(logger, tracer, nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/consumer/negotiations/test-pid/agreement", bytes.NewBufferString("{invalid"))
	rec := httptest.NewRecorder()

	handler.HandleContractAgreement(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request, got %d", rec.Code)
	}
}

func TestConsumerCallbackHandler_HandleNegotiationEvent_InvalidJSON(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	tracer := noop.NewTracerProvider().Tracer("test")

	handler := ports.NewConsumerCallbackHandler(logger, tracer, nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/consumer/negotiations/test-pid/events", bytes.NewBufferString("{invalid"))
	rec := httptest.NewRecorder()

	handler.HandleNegotiationEvent(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request, got %d", rec.Code)
	}
}

func TestConsumerCallbackHandler_HandleTransferStart_InvalidJSON(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	tracer := noop.NewTracerProvider().Tracer("test")

	handler := ports.NewConsumerCallbackHandler(logger, tracer, nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/consumer/transfers/test-pid/start", bytes.NewBufferString("{invalid"))
	rec := httptest.NewRecorder()

	handler.HandleTransferStart(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request, got %d", rec.Code)
	}
}
