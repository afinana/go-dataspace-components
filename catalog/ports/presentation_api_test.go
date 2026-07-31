package ports_test

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"go.opentelemetry.io/otel/trace/noop"

	"github.com/afinana/go-dataspace-components/catalog/ports"
)

func TestCatalogAPIHandler_RegisterRoutes(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	tracer := noop.NewTracerProvider().Tracer("test")

	handler := ports.NewCatalogAPIHandler(logger, tracer, nil, nil)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodPost, "/catalog/datasets", bytes.NewBufferString("{invalid-json"))
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request, got %d", rec.Code)
	}
}
