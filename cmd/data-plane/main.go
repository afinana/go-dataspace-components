package main

import (
	"fmt"
	"net/http"
	"os"

	dp "github.com/afinana/go-dataspace-components/data-plane/domain"
	"github.com/afinana/go-dataspace-components/data-plane/ports"
	"github.com/afinana/go-dataspace-components/internal/pkg/config"
	"github.com/afinana/go-dataspace-components/internal/pkg/logging"
	"github.com/afinana/go-dataspace-components/internal/pkg/telemetry"
)

func main() {
	cfg := config.LoadConfig()
	logger := logging.InitLogger(cfg.LogLevel)
	logger.Info("Starting Data Plane service...", "env", cfg.Environment)

	_, shutdown, err := telemetry.InitTelemetry(cfg.ServiceName)
	if err != nil {
		logger.Error("Failed to initialize OpenTelemetry", "err", err)
		os.Exit(1)
	}
	_ = shutdown

	// Initialize proxy and streaming controllers
	proxyController := ports.NewAPIProxyController(logger)
	streamController := ports.NewFileStreamController(logger)

	controllers := []dp.DataFlowController{
		proxyController,
		streamController,
	}

	// Setup standard Signaling Listener
	signalingListener := ports.NewSignalingListener(logger, controllers)
	mux := http.NewServeMux()
	signalingListener.RegisterRoutes(mux)

	// Expose proxy endpoint path for consumers to request proxied data assets
	mux.Handle("/public/", proxyController)

	// Health check route
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	serverAddr := fmt.Sprintf(":%d", cfg.Port)
	logger.Info("Data Plane server listening", "addr", serverAddr)
	if err := http.ListenAndServe(serverAddr, mux); err != nil {
		logger.Error("Data Plane server failed to run", "err", err)
		os.Exit(1)
	}
}
