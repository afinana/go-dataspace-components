package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/afinana/go-dataspace-components/catalog/ports"
	"github.com/afinana/go-dataspace-components/internal/pkg/config"
	"github.com/afinana/go-dataspace-components/internal/pkg/logging"
	"github.com/afinana/go-dataspace-components/internal/pkg/telemetry"
	_ "github.com/lib/pq"
)

func main() {
	// 1. Load Configurations
	cfg := config.LoadConfig()

	// 2. Setup Structured slog Logger
	logger := logging.InitLogger(cfg.LogLevel)
	logger.Info("Starting Sovereign Data Catalog service...", "env", cfg.Environment)

	// 3. Setup OpenTelemetry
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

	// 4. Establish database connection with connection retries (highly resilient under container starts)
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

	// 5. Instantiate Postgres Catalog Store
	catalogStore := ports.NewPostgresCatalogStore(db, "did:web:local-connector")

	// 6. Setup HTTP Presentation Handlers
	handler := ports.NewCatalogAPIHandler(logger, tel.Tracer, catalogStore, catalogStore)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	// Health check route
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	serverAddr := fmt.Sprintf(":%d", cfg.Port)
	logger.Info("Catalog Service server listening", "addr", serverAddr)
	if err := http.ListenAndServe(serverAddr, mux); err != nil {
		logger.Error("Catalog Service server failed to run", "err", err)
		os.Exit(1)
	}
}
