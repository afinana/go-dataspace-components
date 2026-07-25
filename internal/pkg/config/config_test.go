package config_test

import (
	"os"
	"testing"

	"github.com/afinana/go-dataspace-components/internal/pkg/config"
)

func TestLoadConfig_Defaults(t *testing.T) {
	// Ensure env vars are cleared for test
	os.Unsetenv("SERVICE_NAME")
	os.Unsetenv("ENVIRONMENT")
	os.Unsetenv("LOG_LEVEL")
	os.Unsetenv("CONTROL_PLANE_URL")
	os.Unsetenv("DATA_PLANE_URL")
	os.Unsetenv("PORT")
	os.Unsetenv("DATABASE_URL")

	cfg := config.LoadConfig()

	if cfg.ServiceName != "dataspace-connector" {
		t.Errorf("expected default ServiceName 'dataspace-connector', got '%s'", cfg.ServiceName)
	}
	if cfg.Environment != "production" {
		t.Errorf("expected default Environment 'production', got '%s'", cfg.Environment)
	}
	if cfg.LogLevel != "INFO" {
		t.Errorf("expected default LogLevel 'INFO', got '%s'", cfg.LogLevel)
	}
	if cfg.ControlPlaneURL != "http://localhost:8081" {
		t.Errorf("expected default ControlPlaneURL 'http://localhost:8081', got '%s'", cfg.ControlPlaneURL)
	}
	if cfg.DataPlaneURL != "http://localhost:8082" {
		t.Errorf("expected default DataPlaneURL 'http://localhost:8082', got '%s'", cfg.DataPlaneURL)
	}
	if cfg.Port != 8080 {
		t.Errorf("expected default Port 8080, got %d", cfg.Port)
	}
	if cfg.DatabaseURL == "" {
		t.Error("expected non-empty default DatabaseURL")
	}
}

func TestLoadConfig_EnvOverrides(t *testing.T) {
	t.Setenv("SERVICE_NAME", "custom-service")
	t.Setenv("ENVIRONMENT", "development")
	t.Setenv("LOG_LEVEL", "DEBUG")
	t.Setenv("CONTROL_PLANE_URL", "http://custom-cp:8081")
	t.Setenv("DATA_PLANE_URL", "http://custom-dp:8082")
	t.Setenv("PORT", "9090")
	t.Setenv("DATABASE_URL", "postgres://user:pass@db:5432/dbname")

	cfg := config.LoadConfig()

	if cfg.ServiceName != "custom-service" {
		t.Errorf("expected ServiceName 'custom-service', got '%s'", cfg.ServiceName)
	}
	if cfg.Environment != "development" {
		t.Errorf("expected Environment 'development', got '%s'", cfg.Environment)
	}
	if cfg.LogLevel != "DEBUG" {
		t.Errorf("expected LogLevel 'DEBUG', got '%s'", cfg.LogLevel)
	}
	if cfg.ControlPlaneURL != "http://custom-cp:8081" {
		t.Errorf("expected ControlPlaneURL 'http://custom-cp:8081', got '%s'", cfg.ControlPlaneURL)
	}
	if cfg.DataPlaneURL != "http://custom-dp:8082" {
		t.Errorf("expected DataPlaneURL 'http://custom-dp:8082', got '%s'", cfg.DataPlaneURL)
	}
	if cfg.Port != 9090 {
		t.Errorf("expected Port 9090, got %d", cfg.Port)
	}
	if cfg.DatabaseURL != "postgres://user:pass@db:5432/dbname" {
		t.Errorf("expected DatabaseURL 'postgres://user:pass@db:5432/dbname', got '%s'", cfg.DatabaseURL)
	}
}

func TestLoadConfig_InvalidPortFallback(t *testing.T) {
	t.Setenv("PORT", "not-a-number")

	cfg := config.LoadConfig()

	if cfg.Port != 8080 {
		t.Errorf("expected fallback Port 8080 for invalid int string, got %d", cfg.Port)
	}
}
