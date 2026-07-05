package ports

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/afinana/go-dataspace-components/catalog/domain"
	"github.com/afinana/go-dataspace-components/internal/pkg/logging"
	"github.com/afinana/go-dataspace-components/internal/pkg/telemetry"
	"go.opentelemetry.io/otel/trace"
)

// CatalogAPIHandler handles the DCAT HTTP public catalog API.
type CatalogAPIHandler struct {
	logger        *slog.Logger
	tracer        trace.Tracer
	assetRegistry domain.AssetRegistry
	queryService  domain.CatalogQueryService
}

// NewCatalogAPIHandler initializes a new CatalogAPIHandler.
func NewCatalogAPIHandler(logger *slog.Logger, tracer trace.Tracer, assetRegistry domain.AssetRegistry, queryService domain.CatalogQueryService) *CatalogAPIHandler {
	return &CatalogAPIHandler{
		logger:        logger,
		tracer:        tracer,
		assetRegistry: assetRegistry,
		queryService:  queryService,
	}
}

// RegisterRoutes registers the W3C DCAT API paths.
func (h *CatalogAPIHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /catalog", h.handleGetCatalog)
	mux.HandleFunc("POST /catalog/datasets", h.handleRegisterDataset)
	mux.HandleFunc("GET /catalog/datasets", h.handleListDatasets)
	mux.HandleFunc("GET /catalog/datasets/{id}", h.handleGetDataset)
	mux.HandleFunc("DELETE /catalog/datasets/{id}", h.handleDeleteDataset)
}

func (h *CatalogAPIHandler) handleGetCatalog(w http.ResponseWriter, r *http.Request) {
	ctx, span := telemetry.StartSpan(r.Context(), h.tracer, "CatalogAPIHandler.handleGetCatalog")
	defer span.End()

	logger := logging.WithContext(ctx, h.logger)
	logger.Info("Fetching full catalog")

	requester := r.URL.Query().Get("requester")
	catalog, err := h.queryService.GetCatalog(ctx, requester)
	if err != nil {
		logger.Error("Failed to fetch catalog", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(catalog)
}

func (h *CatalogAPIHandler) handleRegisterDataset(w http.ResponseWriter, r *http.Request) {
	ctx, span := telemetry.StartSpan(r.Context(), h.tracer, "CatalogAPIHandler.handleRegisterDataset")
	defer span.End()

	logger := logging.WithContext(ctx, h.logger)

	var dataset domain.Dataset
	if err := json.NewDecoder(r.Body).Decode(&dataset); err != nil {
		logger.Error("Failed to decode dataset payload", "err", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	logger.Info("Registering dataset", "id", dataset.ID, "title", dataset.Title)

	if err := h.assetRegistry.RegisterDataset(ctx, &dataset); err != nil {
		logger.Error("Failed to register dataset", "id", dataset.ID, "err", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{"success": true, "id": dataset.ID})
}

func (h *CatalogAPIHandler) handleListDatasets(w http.ResponseWriter, r *http.Request) {
	ctx, span := telemetry.StartSpan(r.Context(), h.tracer, "CatalogAPIHandler.handleListDatasets")
	defer span.End()

	logger := logging.WithContext(ctx, h.logger)
	logger.Info("Listing datasets")

	datasets, err := h.assetRegistry.ListDatasets(ctx)
	if err != nil {
		logger.Error("Failed to list datasets", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(datasets)
}

func (h *CatalogAPIHandler) handleGetDataset(w http.ResponseWriter, r *http.Request) {
	ctx, span := telemetry.StartSpan(r.Context(), h.tracer, "CatalogAPIHandler.handleGetDataset")
	defer span.End()

	logger := logging.WithContext(ctx, h.logger)
	id := r.PathValue("id")
	logger.Info("Retrieving dataset", "id", id)

	if id == "" {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	dataset, err := h.assetRegistry.GetDataset(ctx, id)
	if err != nil {
		logger.Warn("Dataset not found", "id", id, "err", err)
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(dataset)
}

func (h *CatalogAPIHandler) handleDeleteDataset(w http.ResponseWriter, r *http.Request) {
	ctx, span := telemetry.StartSpan(r.Context(), h.tracer, "CatalogAPIHandler.handleDeleteDataset")
	defer span.End()

	logger := logging.WithContext(ctx, h.logger)
	id := r.PathValue("id")
	logger.Info("Deleting dataset", "id", id)

	if id == "" {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	if err := h.assetRegistry.DeleteDataset(ctx, id); err != nil {
		logger.Warn("Failed to delete dataset (not found)", "id", id, "err", err)
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{"success": true})
}
