package ports_test

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"go.opentelemetry.io/otel/trace/noop"

	"github.com/afinana/go-dataspace-components/control-plane/domain"
	"github.com/afinana/go-dataspace-components/control-plane/ports"
)

func TestProviderCallbackClient_SendContractAgreement(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	tracer := noop.NewTracerProvider().Tracer("test")

	client := ports.NewProviderCallbackClient(logger, tracer)
	msg := &domain.ContractAgreementMessage{
		ProviderPID: "provider-pid-1",
		ConsumerPID: "consumer-pid-1",
	}

	err := client.SendContractAgreement(context.Background(), server.URL, "consumer-pid-1", msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestProviderCallbackClient_SendTransferStart(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	tracer := noop.NewTracerProvider().Tracer("test")

	client := ports.NewProviderCallbackClient(logger, tracer)
	msg := &domain.TransferStartMessage{
		ProviderPID: "provider-pid-1",
		ConsumerPID: "consumer-pid-1",
	}

	err := client.SendTransferStart(context.Background(), server.URL, "consumer-pid-1", msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestProviderCallbackClient_SendContractOffer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	tracer := noop.NewTracerProvider().Tracer("test")

	client := ports.NewProviderCallbackClient(logger, tracer)
	msg := &domain.ContractOfferMessage{
		ProviderPID: "provider-pid-1",
		ConsumerPID: "consumer-pid-1",
	}

	err := client.SendContractOffer(context.Background(), server.URL, msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestProviderCallbackClient_SendNegotiationEvent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	tracer := noop.NewTracerProvider().Tracer("test")

	client := ports.NewProviderCallbackClient(logger, tracer)
	msg := &domain.ContractNegotiationEventMessage{
		ProviderPID: "provider-pid-1",
		ConsumerPID: "consumer-pid-1",
		EventType:   "ACCEPTED",
	}

	err := client.SendNegotiationEvent(context.Background(), server.URL, "consumer-pid-1", msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
