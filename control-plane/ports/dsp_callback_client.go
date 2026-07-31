package ports

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/afinana/go-dataspace-components/control-plane/domain"
	"github.com/afinana/go-dataspace-components/internal/pkg/logging"
	"github.com/afinana/go-dataspace-components/internal/pkg/telemetry"
	"go.opentelemetry.io/otel/trace"
)

type ProviderCallbackClient struct {
	httpClient *http.Client
	logger     *slog.Logger
	tracer     trace.Tracer
}

func NewProviderCallbackClient(logger *slog.Logger, tracer trace.Tracer) *ProviderCallbackClient {
	return &ProviderCallbackClient{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		logger:     logger,
		tracer:     tracer,
	}
}

func (c *ProviderCallbackClient) sendRequest(ctx context.Context, url string, msg any) error {
	ctx, span := telemetry.StartSpan(ctx, c.tracer, "ProviderCallbackClient.sendRequest")
	defer span.End()

	logger := logging.WithContext(ctx, c.logger)

	payload, err := json.Marshal(msg)
	if err != nil {
		logger.Error("Failed to marshal message", "err", err)
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(payload))
	if err != nil {
		logger.Error("Failed to create request", "err", err)
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		logger.Error("Failed to send request", "err", err)
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		logger.Error("Received non-success status code", "status", resp.StatusCode)
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func (c *ProviderCallbackClient) SendContractOffer(ctx context.Context, consumerCallbackURL string, msg *domain.ContractOfferMessage) error {
	return c.sendRequest(ctx, consumerCallbackURL, msg)
}

func (c *ProviderCallbackClient) SendContractAgreement(ctx context.Context, consumerCallbackURL, consumerPid string, msg *domain.ContractAgreementMessage) error {
	url := fmt.Sprintf("%s/contractnegotiations/%s/agreement", consumerCallbackURL, consumerPid)
	return c.sendRequest(ctx, url, msg)
}

func (c *ProviderCallbackClient) SendNegotiationEvent(ctx context.Context, consumerCallbackURL, consumerPid string, msg *domain.ContractNegotiationEventMessage) error {
	url := fmt.Sprintf("%s/contractnegotiations/%s/events", consumerCallbackURL, consumerPid)
	return c.sendRequest(ctx, url, msg)
}

func (c *ProviderCallbackClient) SendTransferStart(ctx context.Context, consumerCallbackURL, consumerPid string, msg *domain.TransferStartMessage) error {
	url := fmt.Sprintf("%s/transferprocesses/%s/start", consumerCallbackURL, consumerPid)
	return c.sendRequest(ctx, url, msg)
}

func (c *ProviderCallbackClient) SendTransferCompletion(ctx context.Context, consumerCallbackURL, consumerPid string, msg *domain.TransferCompletionMessage) error {
	url := fmt.Sprintf("%s/transferprocesses/%s/completion", consumerCallbackURL, consumerPid)
	return c.sendRequest(ctx, url, msg)
}
