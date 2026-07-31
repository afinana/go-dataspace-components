package ports_test

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"go.opentelemetry.io/otel/trace/noop"

	"github.com/afinana/go-dataspace-components/consumer/ports"
	"github.com/afinana/go-dataspace-components/control-plane/domain"
)

func TestDSPClient_SendCatalogRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/catalog/request" {
			t.Errorf("expected path /catalog/request, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{
			"@type": "dcat:Catalog",
			"id":    "catalog-1",
		})
	}))
	defer server.Close()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	tracer := noop.NewTracerProvider().Tracer("test")

	client := ports.NewDSPClient(logger, tracer)
	resp, err := client.SendCatalogRequest(context.Background(), server.URL, "provider-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp["@type"] != "dcat:Catalog" {
		t.Errorf("expected catalog type dcat:Catalog, got %v", resp["@type"])
	}
}

func TestDSPClient_SendContractRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/negotiations/request" {
			t.Errorf("expected path /negotiations/request, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{
			"@type":       "dspace:ContractNegotiation",
			"providerPid": "provider-pid-123",
		})
	}))
	defer server.Close()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	tracer := noop.NewTracerProvider().Tracer("test")

	client := ports.NewDSPClient(logger, tracer)
	msg := &domain.ContractRequestMessage{
		ConsumerPID:     "consumer-pid-123",
		CallbackAddress: "http://callback",
	}
	resp, err := client.SendContractRequest(context.Background(), server.URL, msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp["providerPid"] != "provider-pid-123" {
		t.Errorf("expected providerPid provider-pid-123, got %v", resp["providerPid"])
	}
}

func TestDSPClient_SendTransferRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/transfers/request" {
			t.Errorf("expected path /transfers/request, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{
			"@type":       "dspace:TransferProcess",
			"providerPid": "provider-transfer-pid-123",
		})
	}))
	defer server.Close()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	tracer := noop.NewTracerProvider().Tracer("test")

	client := ports.NewDSPClient(logger, tracer)
	msg := &domain.TransferRequestMessage{
		ConsumerPID:     "consumer-pid-123",
		CallbackAddress: "http://callback",
		AgreementID:     "agreement-1",
	}
	resp, err := client.SendTransferRequest(context.Background(), server.URL, msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp["providerPid"] != "provider-transfer-pid-123" {
		t.Errorf("expected providerPid provider-transfer-pid-123, got %v", resp["providerPid"])
	}
}
