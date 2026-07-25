package core

import (
	"context"
	"testing"
)

func TestEdcClient_MockFallbacks(t *testing.T) {
	cfg := &EdcConfig{
		ID:              "test-node",
		Name:            "Test Connector Node",
		ControlPlaneURL: "http://127.0.0.1:9999",
		CatalogURL:      "http://127.0.0.1:9999",
		DataPlaneURL:    "http://127.0.0.1:9999",
		IdentityHubURL:  "http://127.0.0.1:9999",
		AuthKey:         "test-auth-key",
	}

	client := NewEdcClient(cfg)
	ctx := context.Background()

	// Test Catalog Fallback
	cat, err := client.GetCatalog(ctx)
	if err != nil {
		t.Fatalf("GetCatalog error: %v", err)
	}
	if cat == nil || len(cat.Datasets) == 0 {
		t.Errorf("expected mock catalog datasets fallback")
	}

	// Test Assets List Fallback
	assets, err := client.ListAssets(ctx)
	if err != nil {
		t.Fatalf("ListAssets error: %v", err)
	}
	if len(assets) == 0 {
		t.Errorf("expected mock assets fallback")
	}

	// Test Policies Fallback
	policies, err := client.ListPolicies(ctx)
	if err != nil {
		t.Fatalf("ListPolicies error: %v", err)
	}
	if len(policies) == 0 {
		t.Errorf("expected mock policies fallback")
	}

	// Test Contract Definitions Fallback
	contractDefs, err := client.ListContractDefinitions(ctx)
	if err != nil {
		t.Fatalf("ListContractDefinitions error: %v", err)
	}
	if len(contractDefs) == 0 {
		t.Errorf("expected mock contract definitions fallback")
	}

	// Test Negotiations Fallback
	negotiations, err := client.GetNegotiations(ctx)
	if err != nil {
		t.Fatalf("GetNegotiations error: %v", err)
	}
	if len(negotiations) == 0 {
		t.Errorf("expected mock negotiations fallback")
	}

	// Test Transfers Fallback
	transfers, err := client.GetTransfers(ctx)
	if err != nil {
		t.Fatalf("GetTransfers error: %v", err)
	}
	if len(transfers) == 0 {
		t.Errorf("expected mock transfers fallback")
	}

	// Test Initiate Negotiation
	negID, err := client.InitiateNegotiation(ctx, "http://provider:8282/api/v1/dsp", "asset-01")
	if err != nil || negID == "" {
		t.Errorf("InitiateNegotiation failed: %v", err)
	}

	// Test Initiate Transfer
	transferID, err := client.InitiateTransfer(ctx, "contract-01", "asset-01", "http://provider:8282/api/v1/dsp", DataAddress{Type: "HttpProxy"})
	if err != nil || transferID == "" {
		t.Errorf("InitiateTransfer failed: %v", err)
	}

	// Test Asset CRUD
	asset := &AssetEntry{ID: "test-asset", Title: "Test Asset"}
	_ = client.CreateAsset(ctx, asset)
	_ = client.UpdateAsset(ctx, asset)
	_ = client.DeleteAsset(ctx, "test-asset")

	// Test Policy CRUD
	policy := &PolicyDefinition{ID: "test-policy"}
	_ = client.CreatePolicy(ctx, policy)
	_ = client.UpdatePolicy(ctx, policy)
	_ = client.DeletePolicy(ctx, "test-policy")

	// Test Contract Definition CRUD
	cd := &ContractDefinition{ID: "test-cd"}
	_ = client.CreateContractDefinition(ctx, cd)
	_ = client.UpdateContractDefinition(ctx, cd)
	_ = client.DeleteContractDefinition(ctx, "test-cd")

	// Test Credentials
	creds, _ := client.GetCredentials(ctx)
	if len(creds) == 0 {
		t.Errorf("expected mock credentials")
	}
}
