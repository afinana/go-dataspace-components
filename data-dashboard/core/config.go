package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// AppConfig maps application properties from app-config.json.
type AppConfig struct {
	AppTitle                   string     `json:"appTitle"`
	HealthCheckIntervalSeconds int        `json:"healthCheckIntervalSeconds"`
	InitialTheme               string     `json:"initialTheme"`
	EnableUserConfig           bool       `json:"enableUserConfig"`
	MenuItems                  []MenuItem `json:"menuItems"`
}

// MenuItem represents a dashboard navbar entry.
type MenuItem struct {
	Text        string `json:"text"`
	Icon        string `json:"icon"`
	Route       string `json:"route"`
	Description string `json:"description"`
}

// EdcConfig defines connection credentials for a target dataspace connector.
type EdcConfig struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	ControlPlaneURL string `json:"controlPlaneUrl"`
	CatalogURL      string `json:"catalogUrl"`
	DataPlaneURL    string `json:"dataPlaneUrl"`
	IdentityHubURL  string `json:"identityHubUrl"`
	AuthKey         string `json:"authKey"`
}

// DashboardConfig aggregates both configurations loaded on startup.
type DashboardConfig struct {
	App        AppConfig
	Connectors []EdcConfig
}

// LoadDashboardConfigs loads config structures from files under path prefix.
func LoadDashboardConfigs(configDir string) (*DashboardConfig, error) {
	appPath := filepath.Join(configDir, "app-config.json")
	connPath := filepath.Join(configDir, "edc-connector-config.json")

	appBytes, err := os.ReadFile(appPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read app configuration from %s: %w", appPath, err)
	}

	var appConfig AppConfig
	if err := json.Unmarshal(appBytes, &appConfig); err != nil {
		return nil, fmt.Errorf("failed to parse app configuration: %w", err)
	}

	connBytes, err := os.ReadFile(connPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read connector configuration from %s: %w", connPath, err)
	}

	var connectors []EdcConfig
	if err := json.Unmarshal(connBytes, &connectors); err != nil {
		return nil, fmt.Errorf("failed to parse connector configurations: %w", err)
	}

	return &DashboardConfig{
		App:        appConfig,
		Connectors: connectors,
	}, nil
}
