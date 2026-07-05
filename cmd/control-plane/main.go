package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/afinana/go-dataspace-components/internal/pkg/config"
	"github.com/afinana/go-dataspace-components/internal/pkg/logging"
	"github.com/afinana/go-dataspace-components/internal/pkg/telemetry"
)

func main() {
	cfg := config.LoadConfig()
	logger := logging.InitLogger(cfg.LogLevel)
	logger.Info("Starting Control Plane service...", "env", cfg.Environment)

	_, shutdown, err := telemetry.InitTelemetry(cfg.ServiceName)
	if err != nil {
		logger.Error("Failed to initialize OpenTelemetry", "err", err)
		os.Exit(1)
	}
	_ = shutdown

	mux := http.NewServeMux()

	// DSP protocol endpoints matching W3C spec
	mux.HandleFunc("POST /protocol/negotiation/request", func(w http.ResponseWriter, r *http.Request) {
		// Mock endpoint for receiving ContractRequestMessage
		logger.Info("Received DSP ContractRequestMessage")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "REQUESTED",
			"id":     "negotiation-01",
		})
	})

	mux.HandleFunc("POST /protocol/negotiation/agreement", func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Received DSP ContractAgreementMessage")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "AGREED",
			"id":     "negotiation-01",
		})
	})

	mux.HandleFunc("POST /protocol/transfer/start", func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Received DSP TransferStartMessage")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "STARTED",
			"id":     "transfer-01",
		})
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	mux.HandleFunc("/mock-backend/", func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Received mock backend request", "method", r.Method, "path", r.URL.Path)
		
		headersMap := make(map[string]string)
		for k, v := range r.Header {
			if len(v) > 0 {
				headersMap[k] = v[0]
			}
		}

		response := map[string]any{
			"url":     r.URL.String(),
			"headers": headersMap,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	})

	serverAddr := fmt.Sprintf(":%d", cfg.Port)
	logger.Info("Control Plane server listening", "addr", serverAddr)
	if err := http.ListenAndServe(serverAddr, mux); err != nil {
		logger.Error("Control Plane server failed to run", "err", err)
		os.Exit(1)
	}
}
