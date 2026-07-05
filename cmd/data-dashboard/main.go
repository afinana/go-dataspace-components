package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/afinana/go-dataspace-components/data-dashboard/core"
	"github.com/afinana/go-dataspace-components/data-dashboard/ports"
	"github.com/afinana/go-dataspace-components/internal/pkg/config"
	"github.com/afinana/go-dataspace-components/internal/pkg/logging"
	"github.com/afinana/go-dataspace-components/internal/pkg/telemetry"
)

func main() {
	cfg := config.LoadConfig()
	logger := logging.InitLogger(cfg.LogLevel)
	logger.Info("Starting EDC Sovereign Data Dashboard...", "env", cfg.Environment)

	_, shutdown, err := telemetry.InitTelemetry(cfg.ServiceName)
	if err != nil {
		logger.Error("Failed to initialize OpenTelemetry", "err", err)
		os.Exit(1)
	}
	_ = shutdown

	// Retrieve configurations and templates directories from environment or defaults
	configDir := getEnv("CONFIG_DIR", "data-dashboard/config")
	templatesDir := getEnv("TEMPLATES_DIR", "data-dashboard/templates")

	// Load JSON configuration descriptors
	dashboardCfg, err := core.LoadDashboardConfigs(configDir)
	if err != nil {
		logger.Error("Failed to load dashboard configurations", "err", err)
		os.Exit(1)
	}

	// Initialize dashboard HTTP server
	server := ports.NewDashboardServer(logger, dashboardCfg, templatesDir)
	mux := http.NewServeMux()
	server.RegisterRoutes(mux)

	// Health check endpoint
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Override port from environment if specifically set
	port := cfg.Port
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = cfg.Port // Keep from config or read env
	}

	serverAddr := fmt.Sprintf(":%d", port)
	logger.Info("Data Dashboard UI listening", "addr", serverAddr)
	if err := http.ListenAndServe(serverAddr, mux); err != nil {
		logger.Error("Data Dashboard server run failure", "err", err)
		os.Exit(1)
	}
}

func getEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}
