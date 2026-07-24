package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDashboardConfigs(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "edc-config-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	appConfigJSON := `{
		"appTitle": "Test EDC Dashboard",
		"healthCheckIntervalSeconds": 30,
		"initialTheme": "dark",
		"enableUserConfig": true,
		"menuItems": [
			{"text": "Home", "icon": "home", "route": "/", "description": "Overview"}
		]
	}`
	connectorConfigJSON := `[
		{
			"id": "provider-edc",
			"name": "Provider EDC (Local)",
			"controlPlaneUrl": "http://control-plane:8081",
			"catalogUrl": "http://control-plane:8081",
			"dataPlaneUrl": "http://data-plane:8082",
			"identityHubUrl": "http://identity-hub:8080",
			"authKey": "secret-123"
		}
	]`

	if err := os.WriteFile(filepath.Join(tmpDir, "app-config.json"), []byte(appConfigJSON), 0644); err != nil {
		t.Fatalf("failed to write app-config: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "edc-connector-config.json"), []byte(connectorConfigJSON), 0644); err != nil {
		t.Fatalf("failed to write edc-connector-config: %v", err)
	}

	cfg, err := LoadDashboardConfigs(tmpDir)
	if err != nil {
		t.Fatalf("LoadDashboardConfigs returned unexpected error: %v", err)
	}

	if cfg.App.AppTitle != "Test EDC Dashboard" {
		t.Errorf("expected AppTitle 'Test EDC Dashboard', got '%s'", cfg.App.AppTitle)
	}
	if cfg.App.HealthCheckIntervalSeconds != 30 {
		t.Errorf("expected HealthCheckIntervalSeconds 30, got %d", cfg.App.HealthCheckIntervalSeconds)
	}
	if !cfg.App.EnableUserConfig {
		t.Errorf("expected EnableUserConfig to be true")
	}
	if len(cfg.Connectors) != 1 {
		t.Fatalf("expected 1 connector, got %d", len(cfg.Connectors))
	}
	if cfg.Connectors[0].ID != "provider-edc" {
		t.Errorf("expected connector ID 'provider-edc', got '%s'", cfg.Connectors[0].ID)
	}
}
