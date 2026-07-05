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

	// --- EDC Management API v4 Compatibility Routes for Bruno/Postman ---

	// POST /api/mgmt/v4/catalog/request -> query provider catalog
	mux.HandleFunc("POST /api/mgmt/v4/catalog/request", func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Received management API CatalogRequest")
		
		response := map[string]any{
			"@context": []string{"https://w3id.org/edc/connector/management/v2"},
			"@type":    "Catalog",
			"@id":      "main-catalog-01",
			"dataset": []map[string]any{
				{
					"@id":         "asset-1",
					"@type":       "dcat:Dataset",
					"title":       "Sovereign Dataset Asset 1",
					"description": "Standard DCAT-AP dataset registered on Alpha node",
					"hasPolicy": []map[string]any{
						{
							"@id":   "policy-offer-999",
							"@type": "odrl:Offer",
						},
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	})

	// POST /api/mgmt/v4/contractnegotiations -> initiate negotiation
	mux.HandleFunc("POST /api/mgmt/v4/contractnegotiations", func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Received management API Initiate Negotiation")
		
		response := map[string]any{
			"@context": []string{"https://w3id.org/edc/connector/management/v2"},
			"@type":    "ContractNegotiation",
			"@id":      "negotiation-01",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	})

	// POST /api/mgmt/v4/contractnegotiations/request -> query negotiation state
	mux.HandleFunc("POST /api/mgmt/v4/contractnegotiations/request", func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Received management API query contract negotiations list")

		response := []map[string]any{
			{
				"@id":                 "negotiation-01",
				"@type":               "ContractNegotiation",
				"state":               "FINALIZED",
				"contractAgreementId": "contract-agreement-01",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	})

	// POST /api/mgmt/v4/transferprocesses -> initiate transfer
	mux.HandleFunc("POST /api/mgmt/v4/transferprocesses", func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Received management API Initiate Transfer")

		response := map[string]any{
			"@context": []string{"https://w3id.org/edc/connector/management/v2"},
			"@type":    "TransferProcess",
			"@id":      "transfer-01",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	})

	// POST /api/mgmt/v4/transferprocesses/request -> query transfer processes state
	mux.HandleFunc("POST /api/mgmt/v4/transferprocesses/request", func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Received management API query transfer processes list")

		response := []map[string]any{
			{
				"@id":       "transfer-01",
				"@type":     "TransferProcess",
				"state":     "STARTED",
				"assetId":   "asset-1",
				"agreementId": "contract-agreement-01",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
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
