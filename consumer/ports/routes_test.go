package ports_test

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"go.opentelemetry.io/otel/trace/noop"

	"github.com/afinana/go-dataspace-components/consumer/ports"
)

func TestRegisterRoutes(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	tracer := noop.NewTracerProvider().Tracer("test")

	dspClient := ports.NewDSPClient(logger, tracer)
	mgmtHandler := ports.NewConsumerManagementHandler(logger, tracer, nil, nil, dspClient, "http://provider", "http://callback")
	callbackHandler := ports.NewConsumerCallbackHandler(logger, tracer, nil, nil)

	mux := http.NewServeMux()
	ports.RegisterRoutes(mux, callbackHandler, mgmtHandler)

	req := httptest.NewRequest(http.MethodPost, "/api/consumer/v4/catalog/request", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	// Route must match (500 or 200, not 404)
	if rec.Code == http.StatusNotFound {
		t.Errorf("expected route to be matched, got 404 Not Found")
	}
}
