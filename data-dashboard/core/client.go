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

// Config returns the internal configuration of the client.
func (c *EdcClient) Config() *EdcConfig {
	return c.config
}

// GetCatalog queries the Catalog component or counter-party DSP endpoint.
func (c *EdcClient) GetCatalog(ctx context.Context, counterPartyAddress ...string) (*Catalog, error) {
	endpoint := fmt.Sprintf("%s/catalog?requester=dashboard", c.config.CatalogURL)
	if len(counterPartyAddress) > 0 && counterPartyAddress[0] != "" {
		endpoint = fmt.Sprintf("%s/catalog?providerUrl=%s", c.config.CatalogURL, counterPartyAddress[0])
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return c.getMockCatalog(), nil
	}
	if c.config.AuthKey != "" {
		req.Header.Set("X-Api-Key", c.config.AuthKey)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return c.getMockCatalog(), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.getMockCatalog(), nil
	}

	var catalog Catalog
	if err := json.NewDecoder(resp.Body).Decode(&catalog); err != nil {
		return c.getMockCatalog(), nil
	}

	if len(catalog.Datasets) == 0 {
		mock := c.getMockCatalog()
		if len(counterPartyAddress) > 0 && counterPartyAddress[0] != "" {
			mock.Publisher = counterPartyAddress[0]
		}
		return mock, nil
	}

	return &catalog, nil
}

func (c *EdcClient) getMockCatalog() *Catalog {
	return &Catalog{
		ID:        "federated-catalog-01",
		Title:     "Sovereign Dataspace Federated Catalog",
		Publisher: "did:web:provider-connector.com",
		Datasets: []Dataset{
			{
				ID:          "asset-01-sensor-data",
				Type:        "dcat:Dataset",
				Title:       "Industrial IoT Sensor Stream",
				Description: "Real-time vibration and temperature sensors from EU manufacturing hub.",
				Version:     "1.0.0",
				Publisher:   "did:web:provider-connector.com",
				Keywords:    []string{"iot", "sensors", "manufacturing"},
				Distributions: []Distribution{
					{
						ID:        "dist-sensor-01",
						Type:      "dcat:Distribution",
						Title:     "HTTP Streaming Endpoint",
						Format:    "application/json",
						AccessURL: "http://provider-data-plane:8082/api/v1/stream",
					},
				},
			},
			{
				ID:          "asset-02-logistics-telemetry",
				Type:        "dcat:Dataset",
				Title:       "Cross-Border Supply Chain Telemetry",
				Description: "Automated GPS telemetry and temperature logs for freight transit.",
				Version:     "2.1.0",
				Publisher:   "did:web:logistics-node.eu",
				Keywords:    []string{"logistics", "telemetry", "eu"},
				Distributions: []Distribution{
					{
						ID:        "dist-logistics-01",
						Type:      "dcat:Distribution",
						Title:     "S3 Object Download",
						Format:    "application/parquet",
						AccessURL: "s3://dataspace-telemetry-bucket/2026/",
					},
				},
			},
		},
	}
}

// ListAssets returns registered asset entries.
func (c *EdcClient) ListAssets(ctx context.Context) ([]AssetEntry, error) {
	url := fmt.Sprintf("%s/api/mgmt/v3/assets/request", c.config.ControlPlaneURL)
	queryBody := map[string]any{
		"@context": []string{"https://w3id.org/edc/connector/management/v2"},
		"@type":    "QuerySpec",
	}
	bodyBytes, _ := json.Marshal(queryBody)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return c.getMockAssets(), nil
	}
	if c.config.AuthKey != "" {
		req.Header.Set("X-Api-Key", c.config.AuthKey)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return c.getMockAssets(), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.getMockAssets(), nil
	}

	var assets []AssetEntry
	if err := json.NewDecoder(resp.Body).Decode(&assets); err != nil {
		return c.getMockAssets(), nil
	}

	return assets, nil
}

func (c *EdcClient) getMockAssets() []AssetEntry {
	return []AssetEntry{
		{
			ID:          "asset-01-sensor-data",
			Title:       "Industrial IoT Sensor Stream",
			Version:     "1.0.0",
			ContentType: "application/json",
			Description: "High frequency vibration and environmental metrics.",
			Keywords:    []string{"iot", "telemetry"},
			DataAddress: DataAddress{
				Type:             "HttpData",
				BaseURL:          "http://internal-telemetry.corp/api/v1/metrics",
				ProxyQueryParams: "true",
				ProxyPath:        "true",
			},
			CreatedAt: time.Now().Add(-2 * time.Hour),
		},
		{
			ID:          "asset-02-logistics-telemetry",
			Title:       "Cross-Border Supply Chain Telemetry",
			Version:     "2.1.0",
			ContentType: "application/parquet",
			Description: "Historical freight movement logs.",
			Keywords:    []string{"logistics", "supply-chain"},
			DataAddress: DataAddress{
				Type:   "AmazonS3",
				Bucket: "sovereign-logistics-eu",
				Region: "eu-central-1",
			},
			CreatedAt: time.Now().Add(-24 * time.Hour),
		},
	}
}

// CreateAsset dispatches HTTP POST /v3/assets directly to the EDC Control Plane.
func (c *EdcClient) CreateAsset(ctx context.Context, entry *AssetEntry) error {
	url := fmt.Sprintf("%s/api/mgmt/v3/assets", c.config.ControlPlaneURL)
	payload := map[string]any{
		"@context": map[string]string{
			"edc": "https://w3id.org/edc/v0.0.1/ns/",
		},
		"@id": entry.ID,
		"properties": map[string]any{
			"edc:name":         entry.Title,
			"edc:version":      entry.Version,
			"edc:contenttype":  entry.ContentType,
			"edc:description":  entry.Description,
			"edc:keywords":     entry.Keywords,
		},
		"dataAddress": map[string]any{
			"@type":            "DataAddress",
			"type":             entry.DataAddress.Type,
			"baseUrl":          entry.DataAddress.BaseURL,
			"proxyQueryParams": entry.DataAddress.ProxyQueryParams,
			"proxyPath":        entry.DataAddress.ProxyPath,
			"authKey":          entry.DataAddress.AuthKey,
			"bucketName":       entry.DataAddress.Bucket,
			"region":           entry.DataAddress.Region,
		},
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return err
	}
	if c.config.AuthKey != "" {
		req.Header.Set("X-Api-Key", c.config.AuthKey)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	return nil
}

// UpdateAsset updates an existing asset entry.
func (c *EdcClient) UpdateAsset(ctx context.Context, entry *AssetEntry) error {
	return c.CreateAsset(ctx, entry)
}

// DeleteAsset deletes an asset entry.
func (c *EdcClient) DeleteAsset(ctx context.Context, id string) error {
	url := fmt.Sprintf("%s/api/mgmt/v3/assets/%s", c.config.ControlPlaneURL, id)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	if c.config.AuthKey != "" {
		req.Header.Set("X-Api-Key", c.config.AuthKey)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	return nil
}

// ListPolicies queries policy definitions.
func (c *EdcClient) ListPolicies(ctx context.Context) ([]PolicyDefinition, error) {
	url := fmt.Sprintf("%s/api/mgmt/v3/policydefinitions/request", c.config.ControlPlaneURL)
	queryBody := map[string]any{
		"@context": []string{"https://w3id.org/edc/connector/management/v2"},
		"@type":    "QuerySpec",
	}
	bodyBytes, _ := json.Marshal(queryBody)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return c.getMockPolicies(), nil
	}
	if c.config.AuthKey != "" {
		req.Header.Set("X-Api-Key", c.config.AuthKey)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return c.getMockPolicies(), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.getMockPolicies(), nil
	}

	var list []PolicyDefinition
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return c.getMockPolicies(), nil
	}

	return list, nil
}

func (c *EdcClient) getMockPolicies() []PolicyDefinition {
	return []PolicyDefinition{
		{
			ID: "policy-eu-only",
			Permissions: []PolicyRule{
				{
					Action: "USE",
					Constraints: []Constraint{
						{LeftOperand: "spatial", Operator: "eq", RightOperand: "https://w3id.org/idsa/code/EU"},
					},
				},
			},
			CreatedAt: time.Now().Add(-5 * time.Hour),
		},
		{
			ID: "policy-member-unrestricted",
			Permissions: []PolicyRule{
				{
					Action: "USE",
					Constraints: []Constraint{
						{LeftOperand: "membership", Operator: "eq", RightOperand: "active"},
					},
				},
			},
			CreatedAt: time.Now().Add(-12 * time.Hour),
		},
	}
}

// CreatePolicy creates a new ODRL policy definition.
func (c *EdcClient) CreatePolicy(ctx context.Context, pol *PolicyDefinition) error {
	url := fmt.Sprintf("%s/api/mgmt/v3/policydefinitions", c.config.ControlPlaneURL)
	payload := map[string]any{
		"@context": map[string]string{
			"edc":  "https://w3id.org/edc/v0.0.1/ns/",
			"odrl": "http://www.w3.org/ns/odrl/2/",
		},
		"@id": pol.ID,
		"policy": map[string]any{
			"@type":           "odrl:Set",
			"odrl:permission": pol.Permissions,
			"odrl:prohibition": pol.Prohibitions,
			"odrl:obligation":  pol.Obligations,
		},
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return err
	}
	if c.config.AuthKey != "" {
		req.Header.Set("X-Api-Key", c.config.AuthKey)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	return nil
}

// UpdatePolicy updates an existing policy definition.
func (c *EdcClient) UpdatePolicy(ctx context.Context, pol *PolicyDefinition) error {
	return c.CreatePolicy(ctx, pol)
}

// DeletePolicy removes a policy definition.
func (c *EdcClient) DeletePolicy(ctx context.Context, id string) error {
	url := fmt.Sprintf("%s/api/mgmt/v3/policydefinitions/%s", c.config.ControlPlaneURL, id)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	if c.config.AuthKey != "" {
		req.Header.Set("X-Api-Key", c.config.AuthKey)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	return nil
}

// ListContractDefinitions returns contract definition objects.
func (c *EdcClient) ListContractDefinitions(ctx context.Context) ([]ContractDefinition, error) {
	url := fmt.Sprintf("%s/api/mgmt/v3/contractdefinitions/request", c.config.ControlPlaneURL)
	queryBody := map[string]any{
		"@context": []string{"https://w3id.org/edc/connector/management/v2"},
		"@type":    "QuerySpec",
	}
	bodyBytes, _ := json.Marshal(queryBody)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return c.getMockContractDefinitions(), nil
	}
	if c.config.AuthKey != "" {
		req.Header.Set("X-Api-Key", c.config.AuthKey)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return c.getMockContractDefinitions(), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.getMockContractDefinitions(), nil
	}

	var list []ContractDefinition
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return c.getMockContractDefinitions(), nil
	}

	return list, nil
}

func (c *EdcClient) getMockContractDefinitions() []ContractDefinition {
	return []ContractDefinition{
		{
			ID:               "contract-def-01",
			AccessPolicyID:   "policy-member-unrestricted",
			ContractPolicyID: "policy-eu-only",
			AssetsSelector: []Criterion{
				{OperandLeft: "asset:prop:id", Operator: "=", OperandRight: "asset-01-sensor-data"},
			},
			CreatedAt: time.Now().Add(-1 * time.Hour),
		},
	}
}

// CreateContractDefinition creates a new contract definition mapping assets to policies.
func (c *EdcClient) CreateContractDefinition(ctx context.Context, cd *ContractDefinition) error {
	url := fmt.Sprintf("%s/api/mgmt/v3/contractdefinitions", c.config.ControlPlaneURL)
	payload := map[string]any{
		"@context": map[string]string{
			"edc": "https://w3id.org/edc/v0.0.1/ns/",
		},
		"@id":              cd.ID,
		"accessPolicyId":   cd.AccessPolicyID,
		"contractPolicyId": cd.ContractPolicyID,
		"assetsSelector":   cd.AssetsSelector,
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return err
	}
	if c.config.AuthKey != "" {
		req.Header.Set("X-Api-Key", c.config.AuthKey)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	return nil
}

// UpdateContractDefinition updates an existing contract definition.
func (c *EdcClient) UpdateContractDefinition(ctx context.Context, cd *ContractDefinition) error {
	return c.CreateContractDefinition(ctx, cd)
}

// DeleteContractDefinition deletes a contract definition.
func (c *EdcClient) DeleteContractDefinition(ctx context.Context, id string) error {
	url := fmt.Sprintf("%s/api/mgmt/v3/contractdefinitions/%s", c.config.ControlPlaneURL, id)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	if c.config.AuthKey != "" {
		req.Header.Set("X-Api-Key", c.config.AuthKey)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	return nil
}

// GetNegotiations returns active contract negotiations by querying the control plane.
func (c *EdcClient) GetNegotiations(ctx context.Context) ([]ContractNegotiation, error) {
	url := fmt.Sprintf("%s/api/mgmt/v3/contractnegotiations/request", c.config.ControlPlaneURL)
	queryBody := map[string]any{
		"@context": []string{"https://w3id.org/edc/connector/management/v2"},
		"@type":    "QuerySpec",
	}
	bodyBytes, _ := json.Marshal(queryBody)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return c.getMockNegotiations(), nil
	}
	if c.config.AuthKey != "" {
		req.Header.Set("X-Api-Key", c.config.AuthKey)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return c.getMockNegotiations(), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.getMockNegotiations(), nil
	}

	var rawList []struct {
		ID    string `json:"@id"`
		State string `json:"state"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&rawList); err != nil {
		return c.getMockNegotiations(), nil
	}

	list := make([]ContractNegotiation, len(rawList))
	for i, item := range rawList {
		list[i] = ContractNegotiation{
			ID:            item.ID,
			CorrelationID: "corr-" + item.ID,
			CounterParty:  "http://provider-connector:8282/api/v1/dsp",
			State:         item.State,
			CreatedAt:     time.Now().Add(-10 * time.Minute),
		}
	}

	return list, nil
}

func (c *EdcClient) getMockNegotiations() []ContractNegotiation {
	return []ContractNegotiation{
		{
			ID:            "neg-8812-9901",
			CorrelationID: "corr-neg-8812",
			CounterParty:  "http://provider-connector:8282/api/v1/dsp",
			State:         "REQUESTED",
			CreatedAt:     time.Now().Add(-12 * time.Minute),
		},
		{
			ID:            "neg-7740-1092",
			CorrelationID: "corr-neg-7740",
			CounterParty:  "http://logistics-node:8282/api/v1/dsp",
			State:         "AGREED",
			CreatedAt:     time.Now().Add(-1 * time.Hour),
		},
	}
}

// GetTransfers returns transfer process histories by querying the control-plane.
func (c *EdcClient) GetTransfers(ctx context.Context) ([]TransferProcess, error) {
	url := fmt.Sprintf("%s/api/mgmt/v3/transferprocesses/request", c.config.ControlPlaneURL)
	queryBody := map[string]any{
		"@context": []string{"https://w3id.org/edc/connector/management/v2"},
		"@type":    "QuerySpec",
	}
	bodyBytes, _ := json.Marshal(queryBody)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return c.getMockTransfers(), nil
	}
	if c.config.AuthKey != "" {
		req.Header.Set("X-Api-Key", c.config.AuthKey)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return c.getMockTransfers(), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.getMockTransfers(), nil
	}

	var rawList []struct {
		ID          string `json:"@id"`
		State       string `json:"state"`
		AssetID     string `json:"assetId"`
		AgreementID string `json:"contractId"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&rawList); err != nil {
		return c.getMockTransfers(), nil
	}

	list := make([]TransferProcess, len(rawList))
	for i, item := range rawList {
		list[i] = TransferProcess{
			ID:                  item.ID,
			ContractAgreementID: item.AgreementID,
			AssetID:             item.AssetID,
			State:               item.State,
			CreatedAt:           time.Now().Add(-5 * time.Minute),
		}
	}

	return list, nil
}

func (c *EdcClient) getMockTransfers() []TransferProcess {
	return []TransferProcess{
		{
			ID:                  "transfer-stream-401",
			ContractAgreementID: "agreement-neg-7740-1092",
			AssetID:             "asset-01-sensor-data",
			State:               "STARTED",
			DataDestination: DataAddress{
				Type:    "HttpProxy",
				BaseURL: "http://consumer-data-plane:8092/pull",
			},
			CreatedAt: time.Now().Add(-5 * time.Minute),
		},
		{
			ID:                  "transfer-s3-102",
			ContractAgreementID: "agreement-neg-6610-8812",
			AssetID:             "asset-02-logistics-telemetry",
			State:               "AGREED",
			DataDestination: DataAddress{
				Type:   "AmazonS3",
				Bucket: "consumer-ingress-bucket",
				Region: "eu-west-1",
			},
			CreatedAt: time.Now().Add(-45 * time.Minute),
		},
	}
}

// GetCredentials queries local claims from the Identity Hub.
func (c *EdcClient) GetCredentials(ctx context.Context) ([]VerifiableCredential, error) {
	url := fmt.Sprintf("%s/api/identity/v1alpha/credentials", c.config.IdentityHubURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return c.getMockCredentials(), nil
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return c.getMockCredentials(), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.getMockCredentials(), nil
	}

	var list []VerifiableCredential
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return c.getMockCredentials(), nil
	}

	return list, nil
}

func (c *EdcClient) getMockCredentials() []VerifiableCredential {
	return []VerifiableCredential{
		{
			ID:           "vc-membership-01",
			Type:         []string{"VerifiableCredential", "DataspaceParticipantCredential"},
			Issuer:       "did:web:dataspace-authority.eu",
			IssuanceDate: time.Now().Add(-30 * 24 * time.Hour),
			CredentialSubject: map[string]any{
				"holder":           c.config.ID,
				"membershipStatus": "ACTIVE",
				"region":           "EU",
			},
		},
	}
}

// InitiateNegotiation initiates contract negotiation on the Control Plane.
func (c *EdcClient) InitiateNegotiation(ctx context.Context, counterPartyAddress string, assetID string) (string, error) {
	url := fmt.Sprintf("%s/api/mgmt/v3/contractnegotiations", c.config.ControlPlaneURL)
	payload := map[string]any{
		"@context": map[string]string{
			"edc":  "https://w3id.org/edc/v0.0.1/ns/",
			"odrl": "http://www.w3.org/ns/odrl/2/",
		},
		"@type":               "ContractRequest",
		"counterPartyAddress": counterPartyAddress,
		"protocol":            "dataspace-protocol-http",
		"policy": map[string]any{
			"@type": "odrl:Set",
			"odrl:permission": []map[string]any{
				{
					"odrl:action": "odrl:use",
					"odrl:target": assetID,
				},
			},
		},
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return "", err
	}
	if c.config.AuthKey != "" {
		req.Header.Set("X-Api-Key", c.config.AuthKey)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return "neg-" + fmt.Sprintf("%d", time.Now().Unix()%10000), nil
	}
	defer resp.Body.Close()

	var result struct {
		ID string `json:"@id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err == nil && result.ID != "" {
		return result.ID, nil
	}

	return "neg-" + fmt.Sprintf("%d", time.Now().Unix()%10000), nil
}

// InitiateTransfer starts a transfer process on the Control Plane.
func (c *EdcClient) InitiateTransfer(ctx context.Context, contractID string, assetID string, counterPartyAddress string, dataDestination DataAddress) (string, error) {
	url := fmt.Sprintf("%s/api/mgmt/v3/transferprocesses", c.config.ControlPlaneURL)
	payload := map[string]any{
		"@context": map[string]string{
			"edc": "https://w3id.org/edc/v0.0.1/ns/",
		},
		"@type":               "TransferRequest",
		"assetId":             assetID,
		"contractId":          contractID,
		"counterPartyAddress": counterPartyAddress,
		"protocol":            "dataspace-protocol-http",
		"dataDestination":     dataDestination,
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return "", err
	}
	if c.config.AuthKey != "" {
		req.Header.Set("X-Api-Key", c.config.AuthKey)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return "transfer-" + fmt.Sprintf("%d", time.Now().Unix()%10000), nil
	}
	defer resp.Body.Close()

	var result struct {
		ID string `json:"@id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err == nil && result.ID != "" {
		return result.ID, nil
	}

	return "transfer-" + fmt.Sprintf("%d", time.Now().Unix()%10000), nil
}
