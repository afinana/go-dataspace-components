package ports

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"go.opentelemetry.io/otel/trace"

	"github.com/afinana/go-dataspace-components/control-plane/domain"
	"github.com/afinana/go-dataspace-components/internal/pkg/jsonld"
)

type DSPClient struct {
	httpClient *http.Client
	logger     *slog.Logger
	tracer     trace.Tracer
}

func NewDSPClient(logger *slog.Logger, tracer trace.Tracer) *DSPClient {
	return &DSPClient{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		logger:     logger,
		tracer:     tracer,
	}
}

func (c *DSPClient) SendCatalogRequest(ctx context.Context, providerDSPURL, counterPartyID string) (map[string]any, error) {
	ctx, span := c.tracer.Start(ctx, "DSPClient.SendCatalogRequest")
	defer span.End()

	reqMap := map[string]any{
		"@context": jsonld.DSPContextArray(),
		"@type":    jsonld.TypeCatalogRequestMessage,
	}

	return c.postDSPMessage(ctx, providerDSPURL+"/catalog/request", reqMap)
}

func (c *DSPClient) SendContractRequest(ctx context.Context, providerDSPURL string, msg *domain.ContractRequestMessage) (map[string]any, error) {
	ctx, span := c.tracer.Start(ctx, "DSPClient.SendContractRequest")
	defer span.End()

	msg.Context = jsonld.DSPContextArray()
	msg.Type = jsonld.TypeContractRequestMessage

	return c.postDSPMessage(ctx, providerDSPURL+"/negotiations/request", msg)
}

func (c *DSPClient) SendAgreementVerification(ctx context.Context, providerDSPURL, providerPID string, msg *domain.ContractAgreementVerificationMessage) error {
	ctx, span := c.tracer.Start(ctx, "DSPClient.SendAgreementVerification")
	defer span.End()

	msg.Context = jsonld.DSPContextArray()
	msg.Type = jsonld.TypeContractAgreementVerificationMessage

	u, err := url.JoinPath(providerDSPURL, "negotiations", providerPID, "agreement", "verification")
	if err != nil {
		return err
	}

	_, err = c.postDSPMessage(ctx, u, msg)
	return err
}

func (c *DSPClient) SendNegotiationEvent(ctx context.Context, providerDSPURL, providerPID string, msg *domain.ContractNegotiationEventMessage) error {
	ctx, span := c.tracer.Start(ctx, "DSPClient.SendNegotiationEvent")
	defer span.End()

	msg.Context = jsonld.DSPContextArray()
	msg.Type = jsonld.TypeContractNegotiationEventMessage

	u, err := url.JoinPath(providerDSPURL, "negotiations", providerPID, "events")
	if err != nil {
		return err
	}

	_, err = c.postDSPMessage(ctx, u, msg)
	return err
}

func (c *DSPClient) SendTransferRequest(ctx context.Context, providerDSPURL string, msg *domain.TransferRequestMessage) (map[string]any, error) {
	ctx, span := c.tracer.Start(ctx, "DSPClient.SendTransferRequest")
	defer span.End()

	msg.Context = jsonld.DSPContextArray()
	msg.Type = jsonld.TypeTransferRequestMessage

	return c.postDSPMessage(ctx, providerDSPURL+"/transfers/request", msg)
}

func (c *DSPClient) SendTransferCompletion(ctx context.Context, providerDSPURL, providerPID string, msg *domain.TransferCompletionMessage) error {
	ctx, span := c.tracer.Start(ctx, "DSPClient.SendTransferCompletion")
	defer span.End()

	msg.Context = jsonld.DSPContextArray()
	msg.Type = jsonld.TypeTransferCompletionMessage

	u, err := url.JoinPath(providerDSPURL, "transfers", providerPID, "completion")
	if err != nil {
		return err
	}

	_, err = c.postDSPMessage(ctx, u, msg)
	return err
}

func (c *DSPClient) postDSPMessage(ctx context.Context, urlStr string, body any) (map[string]any, error) {
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, urlStr, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	c.logger.InfoContext(ctx, "sending DSP message", slog.String("url", urlStr))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if len(respBody) == 0 {
		return nil, nil
	}

	var result map[string]any
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return result, nil
}
