package ports

import (
	"encoding/json"
	"html/template"
	"log/slog"
	"net/http"
	"path/filepath"
	"time"

	"github.com/afinana/go-dataspace-components/data-dashboard/core"
)

// DashboardServer coordinates rendering view pages.
type DashboardServer struct {
	logger       *slog.Logger
	config       *core.DashboardConfig
	templatesDir string
	client       *core.EdcClient
}

// NewDashboardServer initializes the dashboard server handlers.
func NewDashboardServer(logger *slog.Logger, cfg *core.DashboardConfig, templatesDir string) *DashboardServer {
	// Set default client to the first pre-configured connector
	var client *core.EdcClient
	if len(cfg.Connectors) > 0 {
		client = core.NewEdcClient(&cfg.Connectors[0])
	}

	return &DashboardServer{
		logger:       logger,
		config:       cfg,
		templatesDir: templatesDir,
		client:       client,
	}
}

// RegisterRoutes binds HTTP views to handler routes.
func (s *DashboardServer) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /", s.handleHome)
	mux.HandleFunc("GET /assets", s.handleAssets)
	mux.HandleFunc("POST /assets", s.handleCreateAsset)
	mux.HandleFunc("POST /assets/delete", s.handleDeleteAsset)
	mux.HandleFunc("GET /catalog", s.handleCatalog)
	mux.HandleFunc("GET /policies", s.handlePolicies)
	mux.HandleFunc("GET /transfer", s.handleTransfer)

	// API endpoints for dynamic GUI operations
	mux.HandleFunc("GET /api/connector/health", s.handleConnectorHealth)
	mux.HandleFunc("POST /api/negotiate/start", s.handleInitiateNegotiation)
	mux.HandleFunc("POST /api/transfer/start", s.handleInitiateTransfer)

	// Serve CSS styling statically
	mux.HandleFunc("GET /styles.css", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css")
		http.ServeFile(w, r, filepath.Join(s.templatesDir, "core", "style.css"))
	})
}

// ViewParams aggregates parameters passed into the root Go html/template parser.
type ViewParams struct {
	Title        string
	MenuItems    []core.MenuItem
	Connectors   []core.EdcConfig
	ActiveConnID string
	ActiveTab    string
	Data         any
}

// renderView parses and renders modular view templates wrapped in the base core layout.
func (s *DashboardServer) renderView(w http.ResponseWriter, r *http.Request, activeTab string, viewTemplate string, data any) {
	// Handle dynamic connector switching if query param is set
	connID := r.URL.Query().Get("connector")
	if connID != "" {
		for _, conn := range s.config.Connectors {
			if conn.ID == connID {
				s.client = core.NewEdcClient(&conn)
				s.logger.Info("Switched active connector context", "connectorId", connID)
				break
			}
		}
	}

	activeConnectorID := ""
	if s.config.Connectors != nil && len(s.config.Connectors) > 0 {
		activeConnectorID = s.config.Connectors[0].ID
	}

	params := ViewParams{
		Title:        s.config.App.AppTitle,
		MenuItems:    s.config.App.MenuItems,
		Connectors:   s.config.Connectors,
		ActiveConnID: activeConnectorID,
		ActiveTab:    activeTab,
		Data:         data,
	}

	layoutPath := filepath.Join(s.templatesDir, "core", "layout.html")
	viewPath := filepath.Join(s.templatesDir, viewTemplate)

	tmpl, err := template.ParseFiles(layoutPath, viewPath)
	if err != nil {
		s.logger.Error("Failed to parse templates", "err", err, "view", viewTemplate)
		http.Error(w, "Internal Template Error", http.StatusInternalServerError)
		return
	}

	// Execution merges variables into the core layout shell
	if err := tmpl.ExecuteTemplate(w, "layout", params); err != nil {
		s.logger.Error("Template execution failed", "err", err)
		http.Error(w, "Execution Failure", http.StatusInternalServerError)
	}
}

func (s *DashboardServer) handleHome(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	creds, _ := s.client.GetCredentials(ctx)

	// Combine simple stats for overview index
	stats := map[string]any{
		"CredentialsCount": len(creds),
		"ConnectorsCount":  len(s.config.Connectors),
	}

	s.renderView(w, r, "Overview", "home/index.html", stats)
}

func (s *DashboardServer) handleAssets(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	datasets, err := s.client.ListDatasets(ctx)
	if err != nil {
		s.logger.Error("Failed to fetch assets for dashboard", "err", err)
	}

	s.renderView(w, r, "Asset Catalog", "assets/index.html", datasets)
}

func (s *DashboardServer) handleCreateAsset(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.FormValue("id")
	title := r.FormValue("title")
	desc := r.FormValue("description")
	format := r.FormValue("format")
	accessURL := r.FormValue("accessUrl")

	dataset := core.Dataset{
		ID:          id,
		Type:        "dcat:Dataset",
		Title:       title,
		Description: desc,
		Distributions: []core.Distribution{
			{
				ID:        id + "-dist",
				Type:      "dcat:Distribution",
				Title:     title + " Distribution",
				Format:    format,
				AccessURL: accessURL,
			},
		},
	}

	if err := s.client.RegisterDataset(ctx, &dataset); err != nil {
		s.logger.Error("Failed to register asset via dashboard", "err", err)
	}

	http.Redirect(w, r, "/assets", http.StatusSeeOther)
}

func (s *DashboardServer) handleDeleteAsset(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.FormValue("id")

	if err := s.client.DeleteDataset(ctx, id); err != nil {
		s.logger.Error("Failed to delete asset via dashboard", "id", id, "err", err)
	}

	http.Redirect(w, r, "/assets", http.StatusSeeOther)
}

func (s *DashboardServer) handleCatalog(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	catalog, err := s.client.GetCatalog(ctx)
	if err != nil {
		s.logger.Error("Failed to query federated catalog", "err", err)
	}

	s.renderView(w, r, "Federated Catalog", "catalog/index.html", catalog)
}

func (s *DashboardServer) handlePolicies(w http.ResponseWriter, r *http.Request) {
	// Policies page details
	s.renderView(w, r, "Policy Definitions", "policies/index.html", nil)
}

func (s *DashboardServer) handleTransfer(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	negotiations, _ := s.client.GetNegotiations(ctx)
	transfers, _ := s.client.GetTransfers(ctx)

	data := map[string]any{
		"Negotiations": negotiations,
		"Transfers":    transfers,
	}

	s.renderView(w, r, "Negotiations & Transfers", "transfer/index.html", data)
}

func (s *DashboardServer) handleConnectorHealth(w http.ResponseWriter, r *http.Request) {
	if s.client == nil || s.client.Config() == nil {
		http.Error(w, "No active connector context", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	probe := func(url string) string {
		client := &http.Client{Timeout: 1 * time.Second}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url+"/health", nil)
		if err != nil {
			return "DOWN"
		}
		resp, err := client.Do(req)
		if err != nil {
			return "DOWN"
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			return "UP"
		}
		return "DOWN"
	}

	cfg := s.client.Config()
	status := map[string]string{
		"controlPlane": probe(cfg.ControlPlaneURL),
		"dataPlane":    probe(cfg.DataPlaneURL),
		"identityHub":  probe(cfg.IdentityHubURL),
		"catalog":      probe(cfg.CatalogURL),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func (s *DashboardServer) handleInitiateNegotiation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload struct {
		CounterPartyAddress string `json:"counterPartyAddress"`
		AssetID             string `json:"assetId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	negID, err := s.client.InitiateNegotiation(r.Context(), payload.CounterPartyAddress, payload.AssetID)
	if err != nil {
		s.logger.Error("Failed to initiate negotiation", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"id": negID})
}

func (s *DashboardServer) handleInitiateTransfer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload struct {
		ContractID          string `json:"contractId"`
		AssetID             string `json:"assetId"`
		CounterPartyAddress string `json:"counterPartyAddress"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	transferID, err := s.client.InitiateTransfer(r.Context(), payload.ContractID, payload.AssetID, payload.CounterPartyAddress)
	if err != nil {
		s.logger.Error("Failed to initiate transfer", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"id": transferID})
}
