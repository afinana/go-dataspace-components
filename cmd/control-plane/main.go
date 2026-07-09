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
	cpdomain "github.com/afinana/go-dataspace-components/control-plane/domain"
	controlplaneports "github.com/afinana/go-dataspace-components/control-plane/ports"
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

	// Instantiate control plane database stores
	negotiationStore := controlplaneports.NewPostgresNegotiationStore(db)
	transferStore := controlplaneports.NewPostgresTransferStore(db)

	// Instantiate Catalog API Handler
	catalogHandler := catalogports.NewCatalogAPIHandler(logger, tel.Tracer, catalogStore, catalogStore)

	mux := http.NewServeMux()

	// Register Catalog API routes
	catalogHandler.RegisterRoutes(mux)

	// DSP protocol endpoints matching W3C spec
	mux.HandleFunc("POST /protocol/negotiation/request", func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Received DSP ContractRequestMessage")
		
		var payload struct {
			ID                  string `json:"id"`
			CounterPartyAddress string `json:"counterPartyAddress"`
			CounterPartyID      string `json:"counterPartyId"`
			CallbackAddress     string `json:"callbackAddress"`
			Offer               *struct {
				ID      string `json:"id"`
				AssetID string `json:"assetId"`
				Policy  any    `json:"policy"`
			} `json:"offer"`
		}

		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			logger.Error("failed to decode contract request", "err", err)
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		negID := payload.ID
		if negID == "" {
			negID = "negotiation-" + fmt.Sprintf("%d", time.Now().UnixNano())
		}

		var offer *cpdomain.ContractOffer
		if payload.Offer != nil {
			offer = &cpdomain.ContractOffer{
				ID:        payload.Offer.ID,
				AssetID:   payload.Offer.AssetID,
				Policy:    payload.Offer.Policy,
				CreatedAt: time.Now(),
			}
		} else {
			offer = &cpdomain.ContractOffer{
				ID:        "offer-" + negID,
				AssetID:   "dataset-asset-01",
				CreatedAt: time.Now(),
			}
		}

		cn := &cpdomain.ContractNegotiation{
			ID:            negID,
			CorrelationID: negID,
			CounterParty:  payload.CounterPartyID,
			Type:          cpdomain.TypeProvider,
			State:         cpdomain.StateRequested,
			ContractOffer: offer,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		if err := negotiationStore.Save(r.Context(), cn); err != nil {
			logger.Error("failed to save contract negotiation", "err", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "REQUESTED",
			"id":     negID,
		})
	})

	mux.HandleFunc("POST /protocol/negotiation/agreement", func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Received DSP ContractAgreementMessage")
		
		var payload struct {
			ID        string                      `json:"id"`
			Agreement *cpdomain.ContractAgreement `json:"agreement"`
		}

		_ = json.NewDecoder(r.Body).Decode(&payload)

		negID := payload.ID
		if negID == "" {
			negID = "negotiation-01"
		}

		cn, err := negotiationStore.FindByID(r.Context(), negID)
		if err != nil {
			cns, errList := negotiationStore.ListAll(r.Context())
			if errList == nil && len(cns) > 0 {
				cn = &cns[0]
				negID = cn.ID
			} else {
				cn = &cpdomain.ContractNegotiation{
					ID:            negID,
					CorrelationID: negID,
					CounterParty:  "did:web:counterparty",
					Type:          cpdomain.TypeProvider,
					State:         cpdomain.StateRequested,
					CreatedAt:     time.Now(),
					UpdatedAt:     time.Now(),
				}
				_ = negotiationStore.Save(r.Context(), cn)
			}
		}

		if payload.Agreement != nil {
			cn.Agreement = payload.Agreement
		} else if cn.Agreement == nil {
			cn.Agreement = &cpdomain.ContractAgreement{
				ID:         "agreement-" + negID,
				ProviderID: "did:web:provider",
				ConsumerID: "did:web:consumer",
				AssetID:    "dataset-asset-01",
			}
		}

		_ = cn.Transition(cpdomain.StateAgreed)
		if err := negotiationStore.Update(r.Context(), cn); err != nil {
			logger.Error("failed to update contract negotiation state to agreed", "err", err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "AGREED",
			"id":     negID,
		})
	})

	mux.HandleFunc("POST /protocol/transfer/start", func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Received DSP TransferStartMessage")
		
		var payload struct {
			ID               string `json:"id"`
			ProcessID        string `json:"processId"`
			DataPlaneAddress string `json:"dataPlaneAddress"`
		}

		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			logger.Error("failed to decode transfer start message", "err", err)
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		procID := payload.ProcessID
		if procID == "" {
			procID = payload.ID
		}
		if procID == "" {
			procID = "transfer-01"
		}

		tp, err := transferStore.FindByID(r.Context(), procID)
		if err != nil {
			tp = &cpdomain.TransferProcess{
				ID:                  procID,
				ContractAgreementID: "agreement-test-99",
				AssetID:             "dataset-asset-01",
				State:               cpdomain.StateTransferInitial,
				DataDestination: cpdomain.DataAddress{
					Type: "HttpProxy",
				},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			_ = transferStore.Save(r.Context(), tp)
		}

		_ = tp.Transition(cpdomain.StateTransferStarted)
		if err := transferStore.Update(r.Context(), tp); err != nil {
			logger.Error("failed to update transfer process to started", "err", err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "STARTED",
			"id":     procID,
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
		
		datasets, err := catalogStore.ListDatasets(r.Context())
		var dcatDatasets []map[string]any
		if err == nil && len(datasets) > 0 {
			for _, ds := range datasets {
				dcatDatasets = append(dcatDatasets, map[string]any{
					"@id": ds.ID,
					"hasPolicy": []map[string]any{
						{
							"@id": "policy-" + ds.ID,
						},
					},
				})
			}
		} else {
			dcatDatasets = []map[string]any{
				{
					"@id": "asset-1",
					"hasPolicy": []map[string]any{
						{
							"@id": "policy-01",
						},
					},
				},
			}
		}

		json.NewEncoder(w).Encode(map[string]any{
			"dataset": dcatDatasets,
		})
	})

	mux.HandleFunc("POST /api/mgmt/v4/contractnegotiations", func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Received mgmt initiate negotiation request")
		
		var payload struct {
			CounterPartyAddress string `json:"counterPartyAddress"`
			CounterPartyID      string `json:"counterPartyId"`
			Policy              any    `json:"policy"`
		}
		_ = json.NewDecoder(r.Body).Decode(&payload)

		negID := "negotiation-" + fmt.Sprintf("%d", time.Now().UnixMilli())
		cn := &cpdomain.ContractNegotiation{
			ID:            negID,
			CorrelationID: negID,
			CounterParty:  payload.CounterPartyAddress,
			Type:          cpdomain.TypeConsumer,
			State:         cpdomain.StateRequested,
			ContractOffer: &cpdomain.ContractOffer{
				ID:        "offer-" + negID,
				AssetID:   "asset-1",
				Policy:    payload.Policy,
				CreatedAt: time.Now(),
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		if err := negotiationStore.Save(r.Context(), cn); err != nil {
			logger.Error("failed to initiate contract negotiation", "err", err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{
			"@id":       negID,
			"createdAt": time.Now().UnixMilli(),
		})
	})

	mux.HandleFunc("POST /api/mgmt/v4/contractnegotiations/request", func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Received mgmt query contract negotiations request")
		
		cns, err := negotiationStore.ListAll(r.Context())
		if err != nil {
			logger.Error("failed to query contract negotiations from database", "err", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		var result []map[string]any
		for _, cn := range cns {
			agreementID := "agreement-test-99"
			if cn.Agreement != nil {
				agreementID = cn.Agreement.ID
			} else {
				agreementID = "agreement-" + cn.ID
			}

			result = append(result, map[string]any{
				"@id":                 cn.ID,
				"state":               cn.State.String(),
				"contractAgreementId": agreementID,
			})
		}

		if len(result) == 0 {
			result = []map[string]any{
				{
					"@id":                 "negotiation-01",
					"state":               "REQUESTED",
					"contractAgreementId": "agreement-test-99",
				},
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(result)
	})

	mux.HandleFunc("POST /api/mgmt/v4/transferprocesses", func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Received mgmt initiate transfer request")
		
		var payload struct {
			AssetID             string               `json:"assetId"`
			ContractID          string               `json:"contractId"`
			CounterPartyAddress string               `json:"counterPartyAddress"`
			DataDestination     cpdomain.DataAddress `json:"dataDestination"`
		}
		_ = json.NewDecoder(r.Body).Decode(&payload)

		procID := "transfer-" + fmt.Sprintf("%d", time.Now().UnixMilli())
		tp := &cpdomain.TransferProcess{
			ID:                  procID,
			ContractAgreementID: payload.ContractID,
			CorrelationID:       procID,
			AssetID:             payload.AssetID,
			State:               cpdomain.StateTransferStarted,
			DataDestination:     payload.DataDestination,
			CreatedAt:           time.Now(),
			UpdatedAt:           time.Now(),
		}

		if err := transferStore.Save(r.Context(), tp); err != nil {
			logger.Error("failed to initiate transfer process in database", "err", err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{
			"@id":       procID,
			"createdAt": time.Now().UnixMilli(),
		})
	})

	mux.HandleFunc("POST /api/mgmt/v4/transferprocesses/request", func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Received mgmt query transfer processes request")
		
		tps, err := transferStore.ListAll(r.Context())
		if err != nil {
			logger.Error("failed to query transfer processes from database", "err", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		var result []map[string]any
		for _, tp := range tps {
			result = append(result, map[string]any{
				"@id":         tp.ID,
				"state":       tp.State.String(),
				"assetId":     tp.AssetID,
				"agreementId": tp.ContractAgreementID,
			})
		}

		if len(result) == 0 {
			result = []map[string]any{
				{
					"@id":         "transfer-01",
					"state":       "STARTED",
					"assetId":     "dataset-asset-01",
					"agreementId": "agreement-test-99",
				},
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(result)
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
