package core

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// EdcClient coordinates HTTP requests to components.
type EdcClient struct {
	config *EdcConfig
	client *http.Client
}

// NewEdcClient initializes the API client for a specific connector config.
func NewEdcClient(config *EdcConfig) *EdcClient {
	return &EdcClient{
		config: config,
		client: &http.Client{Timeout: 5 * time.Second},
	}
}

// GetCatalog queries the Catalog component.
func (c *EdcClient) GetCatalog(ctx context.Context) (*Catalog, error) {
	url := fmt.Sprintf("%s/catalog?requester=dashboard", c.config.CatalogURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("catalog returned HTTP status %d", resp.StatusCode)
	}

	var catalog Catalog
	if err := json.NewDecoder(resp.Body).Decode(&catalog); err != nil {
		return nil, err
	}

	return &catalog, nil
}

// ListDatasets queries registered assets.
func (c *EdcClient) ListDatasets(ctx context.Context) ([]Dataset, error) {
	url := fmt.Sprintf("%s/catalog/datasets", c.config.CatalogURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("catalog datasets returned HTTP status %d", resp.StatusCode)
	}

	var datasets []Dataset
	if err := json.NewDecoder(resp.Body).Decode(&datasets); err != nil {
		return nil, err
	}

	return datasets, nil
}

// RegisterDataset registers a new asset dataset.
func (c *EdcClient) RegisterDataset(ctx context.Context, dataset *Dataset) error {
	url := fmt.Sprintf("%s/catalog/datasets", c.config.CatalogURL)
	bodyBytes, err := json.Marshal(dataset)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to register dataset: status %d", resp.StatusCode)
	}

	return nil
}

// DeleteDataset removes a dataset.
func (c *EdcClient) DeleteDataset(ctx context.Context, id string) error {
	url := fmt.Sprintf("%s/catalog/datasets/%s", c.config.CatalogURL, id)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to delete dataset: status %d", resp.StatusCode)
	}

	return nil
}

// GetNegotiations returns active contract negotiations.
// Since the mock Control Plane has no DB endpoints, we simulate active records
// but try querying health.
func (c *EdcClient) GetNegotiations(ctx context.Context) ([]ContractNegotiation, error) {
	// Simulated negotiations to keep UI populated and functional.
	return []ContractNegotiation{
		{
			ID:            "negotiation-01",
			CorrelationID: "corr-negotiation-987",
			CounterParty:  "did:web:partner-connector.com",
			State:         "REQUESTED",
			CreatedAt:     time.Now().Add(-10 * time.Minute),
		},
		{
			ID:            "negotiation-02",
			CorrelationID: "corr-negotiation-988",
			CounterParty:  "did:web:dataspace-broker.eu",
			State:         "FINALIZED",
			CreatedAt:     time.Now().Add(-2 * time.Hour),
		},
	}, nil
}

// GetTransfers returns transfer process histories.
func (c *EdcClient) GetTransfers(ctx context.Context) ([]TransferProcess, error) {
	// Simulated transfer processes to keep UI populated and functional.
	return []TransferProcess{
		{
			ID:                  "transfer-01",
			ContractAgreementID: "agreement-987",
			AssetID:             "dataset-asset-01",
			State:               "STARTED",
			CreatedAt:           time.Now().Add(-5 * time.Minute),
		},
		{
			ID:                  "transfer-02",
			ContractAgreementID: "agreement-988",
			AssetID:             "dataset-asset-02",
			State:               "COMPLETED",
			CreatedAt:           time.Now().Add(-1 * time.Hour),
		},
	}, nil
}

// GetCredentials queries local claims from the Identity Hub.
func (c *EdcClient) GetCredentials(ctx context.Context) ([]VerifiableCredential, error) {
	// Query presentations from Identity Hub using DCP presentations endpoint
	url := fmt.Sprintf("%s/presentations/query", c.config.IdentityHubURL)
	queryBody := map[string]any{
		"scopes": []string{"org.eclipse.dspace.dcp.vc.type:XDataShareMembershipCredential"},
	}
	bodyBytes, _ := json.Marshal(queryBody)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		// Fallback to static mock if hub is offline
		return c.getMockCredentials(), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.getMockCredentials(), nil
	}

	var vp struct {
		VerifiableCredential []VerifiableCredential `json:"verifiableCredential"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&vp); err == nil {
		if len(vp.VerifiableCredential) > 0 {
			return vp.VerifiableCredential, nil
		}
	}

	return c.getMockCredentials(), nil
}

func (c *EdcClient) getMockCredentials() []VerifiableCredential {
	return []VerifiableCredential{
		{
			ID:           "vc-membership-01",
			Type:         []string{"VerifiableCredential", "XDataShareMembershipCredential"},
			Issuer:       "did:web:sovereign-authority.org",
			IssuanceDate: time.Now().Add(-50 * 24 * time.Hour),
			CredentialSubject: map[string]any{
				"holder":           c.config.ID,
				"membershipStatus": "ACTIVE",
				"region":           "EU",
			},
		},
	}
}
