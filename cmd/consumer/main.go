package main

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"time"

	consumerports "github.com/afinana/go-dataspace-components/consumer/ports"
	"github.com/afinana/go-dataspace-components/internal/pkg/config"
	"github.com/afinana/go-dataspace-components/internal/pkg/logging"
	"github.com/afinana/go-dataspace-components/internal/pkg/telemetry"
	_ "github.com/lib/pq"
)

func main() {
	cfg := config.LoadConfig()
	logger := logging.InitLogger(cfg.LogLevel)
	logger.Info("Starting Consumer service...", "env", cfg.Environment)

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

	// Ensure tables are created
	schemaBytes, err := os.ReadFile("consumer/ports/schema.sql")
	if err == nil {
		_, err = db.Exec(string(schemaBytes))
		if err != nil {
			logger.Error("Failed to execute schema.sql", "err", err)
		}
	} else {
		logger.Warn("Could not read schema.sql, assuming tables are already created or schema will be applied elsewhere", "err", err)
	}

	negotiationStore := consumerports.NewPostgresConsumerNegotiationStore(db)
	transferStore := consumerports.NewPostgresConsumerTransferStore(db)

	dspClient := consumerports.NewDSPClient(logger, tel.Tracer)

	providerDSPURL := os.Getenv("PROVIDER_DSP_URL")
	if providerDSPURL == "" {
		providerDSPURL = "http://localhost:8081/api/dsp/2025-1" // fallback for local dev
	}
	consumerCallbackURL := os.Getenv("CONSUMER_CALLBACK_URL")
	if consumerCallbackURL == "" {
		consumerCallbackURL = "http://localhost:8091/consumer" // fallback
	}

	callbackHandler := consumerports.NewConsumerCallbackHandler(logger, tel.Tracer, negotiationStore, transferStore)
	mgmtHandler := consumerports.NewConsumerManagementHandler(logger, tel.Tracer, negotiationStore, transferStore, dspClient, providerDSPURL, consumerCallbackURL)

	mux := http.NewServeMux()

	consumerports.RegisterRoutes(mux, callbackHandler, mgmtHandler)

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	serverAddr := ":8091"
	logger.Info("Consumer server listening", "addr", serverAddr)
	if err := http.ListenAndServe(serverAddr, mux); err != nil {
		logger.Error("Consumer server failed to run", "err", err)
		os.Exit(1)
	}
}
