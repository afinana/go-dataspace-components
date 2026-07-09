package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	catalogports "github.com/afinana/go-dataspace-components/catalog/ports"
	"github.com/afinana/go-dataspace-components/internal/pkg/config"
	"github.com/afinana/go-dataspace-components/internal/pkg/logging"
	"github.com/afinana/go-dataspace-components/internal/pkg/telemetry"
	_ "github.com/lib/pq"
)

func main() {
	cfg := config.LoadConfig()
	logger := logging.InitLogger(cfg.LogLevel)
	logger.Info("Starting Control Plane service...", "env", cfg.Environment)

	tel, shutdown, err := telemetry.InitTelemetry(cfg.ServiceName)
	if err != nil {
		logger.Error("Failed to initialize OpenTelemetry", "err", err)
		os.Exit(1)
	}
	defer func() {
		if err := shutdown(context.Background()); err != nil {
			logger.Error("Failed to shutdown telemetry cleanly", "err", err)
		}
	}()

	// Establish database connection with connection retries (highly resilient under container starts)
	var db *sql.DB
	for attempt := 1; attempt <= 15; attempt++ {
		db, err = sql.Open("postgres", cfg.DatabaseURL)
		if err == nil {
			err = db.Ping()
			if err == nil {
				logger.Info("Successfully connected to database")
				break
			}
		}
		logger.Warn("Database connection failed, retrying in 2 seconds...", "attempt", attempt, "err", err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		logger.Error("Failed to establish database connection after all attempts", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	// Instantiate Postgres Catalog Store
	catalogStore := catalogports.NewPostgresCatalogStore(db, "did:web:local-connector")

	// Instantiate Catalog API Handler
	catalogHandler := catalogports.NewCatalogAPIHandler(logger, tel.Tracer, catalogStore, catalogStore)

	mux := http.NewServeMux()

	// Register Catalog API routes
	catalogHandler.RegisterRoutes(mux)

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

	// EDC Management API compatibility endpoints
	mux.HandleFunc("POST /api/mgmt/v4/assets/request", func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Received mgmt query assets request")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]any{})
	})

	mux.HandleFunc("POST /api/mgmt/v4/catalog/request", func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Received mgmt catalog request")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{
			"dataset": []map[string]any{
				{
					"@id": "asset-1",
					"hasPolicy": []map[string]any{
						{
							"@id": "policy-01",
						},
					},
				},
			},
		})
	})

	mux.HandleFunc("POST /api/mgmt/v4/contractnegotiations", func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Received mgmt initiate negotiation request")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{
			"@id":       "negotiation-01",
			"createdAt": time.Now().UnixMilli(),
		})
	})

	mux.HandleFunc("POST /api/mgmt/v4/contractnegotiations/request", func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Received mgmt query contract negotiations request")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]map[string]any{
			{
				"@id":                 "negotiation-01",
				"state":               "REQUESTED",
				"contractAgreementId": "agreement-test-99",
			},
		})
	})

	mux.HandleFunc("POST /api/mgmt/v4/transferprocesses", func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Received mgmt initiate transfer request")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{
			"@id":       "transfer-01",
			"createdAt": time.Now().UnixMilli(),
		})
	})

	mux.HandleFunc("POST /api/mgmt/v4/transferprocesses/request", func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Received mgmt query transfer processes request")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]map[string]any{
			{
				"@id":         "transfer-01",
				"state":       "STARTED",
				"assetId":     "dataset-asset-01",
				"agreementId": "agreement-test-99",
			},
		})
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	mux.HandleFunc("/mock-backend/", func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Received mock backend request", "path", r.URL.Path, "headers", r.Header)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		
		headers := make(map[string]string)
		for k, v := range r.Header {
			if len(v) > 0 {
				headers[k] = v[0]
			}
		}
		
		response := map[string]any{
			"headers": headers,
			"origin":  r.RemoteAddr,
			"url":     r.URL.String(),
		}
		json.NewEncoder(w).Encode(response)
	})

	serverAddr := fmt.Sprintf(":%d", cfg.Port)
	logger.Info("Control Plane server listening", "addr", serverAddr)
	if err := http.ListenAndServe(serverAddr, mux); err != nil {
		logger.Error("Control Plane server failed to run", "err", err)
		os.Exit(1)
	}
}
