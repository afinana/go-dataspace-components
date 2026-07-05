package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/afinana/go-dataspace-components/identity-hub/ports"
	"github.com/afinana/go-dataspace-components/internal/pkg/config"
	"github.com/afinana/go-dataspace-components/internal/pkg/logging"
	"github.com/afinana/go-dataspace-components/internal/pkg/telemetry"
)

func main() {
	// 1. Load Configurations
	cfg := config.LoadConfig()

	// 2. Setup Structured slog Logger
	logger := logging.InitLogger(cfg.LogLevel)
	logger.Info("Starting Sovereign Identity Hub...", "env", cfg.Environment)

	// 3. Setup OpenTelemetry
	_, shutdown, err := telemetry.InitTelemetry(cfg.ServiceName)
	if err != nil {
		logger.Error("Failed to initialize OpenTelemetry", "err", err)
		os.Exit(1)
	}
	// Simulated shutdown hook on program exit
	_ = shutdown

	// 4. Setup HTTP Presentation Handlers
	handler := ports.NewPresentationAPIHandler(logger)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	// Health check route
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	serverAddr := fmt.Sprintf(":%d", cfg.Port)
	logger.Info("Identity Hub server listening", "addr", serverAddr)
	if err := http.ListenAndServe(serverAddr, mux); err != nil {
		logger.Error("Identity Hub server failed to run", "err", err)
		os.Exit(1)
	}
}
