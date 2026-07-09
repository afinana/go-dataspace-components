package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"time"

	dp "github.com/afinana/go-dataspace-components/data-plane/domain"
	"github.com/afinana/go-dataspace-components/data-plane/ports"
	"github.com/afinana/go-dataspace-components/internal/pkg/config"
	"github.com/afinana/go-dataspace-components/internal/pkg/logging"
	"github.com/afinana/go-dataspace-components/internal/pkg/telemetry"
	_ "github.com/lib/pq"
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

	// Establish database connection with connection retries
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

	// Initialize proxy and streaming controllers
	dbStore := ports.NewPostgresDataFlowStore(db)
	proxyController := ports.NewAPIProxyController(logger, dbStore)
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

	// Expose additional REST proxy management APIs (compat with Bruno collections)
	mux.HandleFunc("GET /api/proxy/flows", proxyController.HandleFlowsList)
	mux.HandleFunc("GET /api/proxy/flows/{flowId}", proxyController.HandleFlowsDetail)
	mux.Handle("/api/proxy/flows/{flowId}/data", proxyController)
	mux.Handle("/api/proxy/flows/{flowId}/data/", proxyController)

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
