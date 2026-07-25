package ports

import (
	"encoding/json"
	"html/template"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"
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
	mux.HandleFunc("POST /assets/edit", s.handleCreateAsset)
	mux.HandleFunc("POST /assets/delete", s.handleDeleteAsset)

	mux.HandleFunc("GET /policies", s.handlePolicies)
	mux.HandleFunc("POST /policies", s.handleCreatePolicy)
	mux.HandleFunc("POST /policies/edit", s.handleCreatePolicy)
	mux.HandleFunc("POST /policies/delete", s.handleDeletePolicy)

	mux.HandleFunc("GET /contract-definitions", s.handleContractDefinitions)
	mux.HandleFunc("POST /contract-definitions", s.handleCreateContractDefinition)
	mux.HandleFunc("POST /contract-definitions/edit", s.handleCreateContractDefinition)
	mux.HandleFunc("POST /contract-definitions/delete", s.handleDeleteContractDefinition)

	mux.HandleFunc("GET /catalog", s.handleCatalog)
	mux.HandleFunc("GET /transfers", s.handleTransfer)

	// API endpoints for dynamic GUI operations
	mux.HandleFunc("GET /api/connector/health", s.handleConnectorHealth)
	mux.HandleFunc("POST /api/catalog/query", s.handleQueryCatalog)
	mux.HandleFunc("GET /api/catalog/query", s.handleQueryCatalog)
	mux.HandleFunc("POST /api/negotiate/start", s.handleInitiateNegotiation)
	mux.HandleFunc("POST /api/transfer/start", s.handleInitiateTransfer)

	// Serve static files (styles.css and public configs)
	mux.HandleFunc("GET /styles.css", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css")
		http.ServeFile(w, r, filepath.Join(s.templatesDir, "core", "style.css"))
	})

	mux.HandleFunc("GET /public/config/app-config.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(s.config.App)
	})

	mux.HandleFunc("GET /public/config/edc-connector-config.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(s.config.Connectors)
	})
}

// ViewParams aggregates parameters passed into the root Go html/template parser.
type ViewParams struct {
	Title            string
	HealthInterval   int
	EnableUserConfig bool
	MenuItems        []core.MenuItem
	Connectors       []core.EdcConfig
	ActiveConnID     string
	ActiveTab        string
	Data             any
}

// renderView parses and renders modular view templates wrapped in the base core layout.
func (s *DashboardServer) renderView(w http.ResponseWriter, r *http.Request, activeTab string, viewTemplate string, data any) {
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
	if s.client != nil && s.client.Config() != nil {
		activeConnectorID = s.client.Config().ID
	} else if len(s.config.Connectors) > 0 {
		activeConnectorID = s.config.Connectors[0].ID
	}

	params := ViewParams{
		Title:            s.config.App.AppTitle,
		HealthInterval:   s.config.App.HealthCheckIntervalSeconds,
		EnableUserConfig: s.config.App.EnableUserConfig,
		MenuItems:        s.config.App.MenuItems,
		Connectors:       s.config.Connectors,
		ActiveConnID:     activeConnectorID,
		ActiveTab:        activeTab,
		Data:             data,
	}

	layoutPath := filepath.Join(s.templatesDir, "core", "layout.html")
	viewPath := filepath.Join(s.templatesDir, viewTemplate)

	tmpl, err := template.ParseFiles(layoutPath, viewPath)
	if err != nil {
		s.logger.Error("Failed to parse templates", "err", err, "view", viewTemplate)
		http.Error(w, "Internal Template Error", http.StatusInternalServerError)
		return
	}

	if err := tmpl.ExecuteTemplate(w, "layout", params); err != nil {
		s.logger.Error("Template execution failed", "err", err)
		http.Error(w, "Execution Failure", http.StatusInternalServerError)
	}
}

func (s *DashboardServer) handleHome(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	assets, _ := s.client.ListAssets(ctx)
	policies, _ := s.client.ListPolicies(ctx)
	contractDefs, _ := s.client.ListContractDefinitions(ctx)
	negotiations, _ := s.client.GetNegotiations(ctx)
	transfers, _ := s.client.GetTransfers(ctx)
	creds, _ := s.client.GetCredentials(ctx)

	stats := map[string]any{
		"AssetsCount":       len(assets),
		"PoliciesCount":     len(policies),
		"ContractsCount":    len(contractDefs),
		"NegotiationsCount": len(negotiations),
		"TransfersCount":    len(transfers),
		"CredentialsCount":  len(creds),
		"ActiveConnector":   s.client.Config(),
		"Credentials":       creds,
	}

	s.renderView(w, r, "Home", "home/index.html", stats)
}

func (s *DashboardServer) handleAssets(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	assets, err := s.client.ListAssets(ctx)
	if err != nil {
		s.logger.Error("Failed to fetch assets for dashboard", "err", err)
	}

	s.renderView(w, r, "Assets", "assets/index.html", assets)
}

func (s *DashboardServer) handleCreateAsset(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.FormValue("id")
	title := r.FormValue("title")
	version := r.FormValue("version")
	contentType := r.FormValue("contentType")
	desc := r.FormValue("description")
	keywordsRaw := r.FormValue("keywords")

	dataType := r.FormValue("dataType")
	baseURL := r.FormValue("baseUrl")
	proxyQueryParams := r.FormValue("proxyQueryParams")
	proxyPath := r.FormValue("proxyPath")
	authKey := r.FormValue("authKey")
	bucket := r.FormValue("bucket")
	region := r.FormValue("region")

	var keywords []string
	if keywordsRaw != "" {
		for _, kw := range strings.Split(keywordsRaw, ",") {
			keywords = append(keywords, strings.TrimSpace(kw))
		}
	}

	entry := core.AssetEntry{
		ID:          id,
		Title:       title,
		Version:     version,
		ContentType: contentType,
		Description: desc,
		Keywords:    keywords,
		DataAddress: core.DataAddress{
			Type:             dataType,
			BaseURL:          baseURL,
			ProxyQueryParams: proxyQueryParams,
			ProxyPath:        proxyPath,
			AuthKey:          authKey,
			Bucket:           bucket,
			Region:           region,
		},
		CreatedAt: time.Now(),
	}

	if err := s.client.CreateAsset(ctx, &entry); err != nil {
		s.logger.Error("Failed to create asset via dashboard", "err", err)
	}

	http.Redirect(w, r, "/assets", http.StatusSeeOther)
}

func (s *DashboardServer) handleDeleteAsset(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.FormValue("id")

	if err := s.client.DeleteAsset(ctx, id); err != nil {
		s.logger.Error("Failed to delete asset via dashboard", "id", id, "err", err)
	}

	http.Redirect(w, r, "/assets", http.StatusSeeOther)
}

func (s *DashboardServer) handlePolicies(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	policies, err := s.client.ListPolicies(ctx)
	if err != nil {
		s.logger.Error("Failed to fetch policy definitions", "err", err)
	}

	s.renderView(w, r, "Policies", "policies/index.html", policies)
}

func (s *DashboardServer) handleCreatePolicy(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.FormValue("id")
	action := r.FormValue("action")
	if action == "" {
		action = "USE"
	}
	leftOp := r.FormValue("leftOperand")
	operator := r.FormValue("operator")
	rightOp := r.FormValue("rightOperand")

	pol := core.PolicyDefinition{
		ID: id,
		Permissions: []core.PolicyRule{
			{
				Action: action,
				Constraints: []core.Constraint{
					{
						LeftOperand:  leftOp,
						Operator:     operator,
						RightOperand: rightOp,
					},
				},
			},
		},
		CreatedAt: time.Now(),
	}

	if err := s.client.CreatePolicy(ctx, &pol); err != nil {
		s.logger.Error("Failed to create policy via dashboard", "err", err)
	}

	http.Redirect(w, r, "/policies", http.StatusSeeOther)
}

func (s *DashboardServer) handleDeletePolicy(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.FormValue("id")

	if err := s.client.DeletePolicy(ctx, id); err != nil {
		s.logger.Error("Failed to delete policy via dashboard", "id", id, "err", err)
	}

	http.Redirect(w, r, "/policies", http.StatusSeeOther)
}

func (s *DashboardServer) handleContractDefinitions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	contractDefs, err := s.client.ListContractDefinitions(ctx)
	if err != nil {
		s.logger.Error("Failed to fetch contract definitions", "err", err)
	}
	policies, _ := s.client.ListPolicies(ctx)
	assets, _ := s.client.ListAssets(ctx)

	data := map[string]any{
		"ContractDefinitions": contractDefs,
		"Policies":            policies,
		"Assets":              assets,
	}

	s.renderView(w, r, "Contract Definitions", "contract-definitions/index.html", data)
}

func (s *DashboardServer) handleCreateContractDefinition(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.FormValue("id")
	accessPolicyID := r.FormValue("accessPolicyId")
	contractPolicyID := r.FormValue("contractPolicyId")
	assetID := r.FormValue("assetId")

	cd := core.ContractDefinition{
		ID:               id,
		AccessPolicyID:   accessPolicyID,
		ContractPolicyID: contractPolicyID,
		AssetsSelector: []core.Criterion{
			{
				OperandLeft:  "asset:prop:id",
				Operator:     "=",
				OperandRight: assetID,
			},
		},
		CreatedAt: time.Now(),
	}

	if err := s.client.CreateContractDefinition(ctx, &cd); err != nil {
		s.logger.Error("Failed to create contract definition", "err", err)
	}

	http.Redirect(w, r, "/contract-definitions", http.StatusSeeOther)
}

func (s *DashboardServer) handleDeleteContractDefinition(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.FormValue("id")

	if err := s.client.DeleteContractDefinition(ctx, id); err != nil {
		s.logger.Error("Failed to delete contract definition", "id", id, "err", err)
	}

	http.Redirect(w, r, "/contract-definitions", http.StatusSeeOther)
}

func (s *DashboardServer) handleCatalog(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	providerURL := r.URL.Query().Get("providerUrl")
	catalog, err := s.client.GetCatalog(ctx, providerURL)
	if err != nil {
		s.logger.Error("Failed to query federated catalog", "err", err)
	}

	data := map[string]any{
		"Catalog":     catalog,
		"ProviderUrl": providerURL,
	}

	s.renderView(w, r, "Federated Catalog", "catalog/index.html", data)
}

func (s *DashboardServer) handleTransfer(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	negotiations, _ := s.client.GetNegotiations(ctx)
	transfers, _ := s.client.GetTransfers(ctx)

	data := map[string]any{
		"Negotiations": negotiations,
		"Transfers":    transfers,
	}

	s.renderView(w, r, "Transfers & Agreements", "transfer/index.html", data)
}

func (s *DashboardServer) handleQueryCatalog(w http.ResponseWriter, r *http.Request) {
	providerURL := r.URL.Query().Get("providerUrl")

	if providerURL == "" && r.Body != nil {
		var payload struct {
			ProviderUrl         string `json:"providerUrl"`
			CounterPartyAddress string `json:"counterPartyAddress"`
		}
		_ = json.NewDecoder(r.Body).Decode(&payload)
		if payload.ProviderUrl != "" {
			providerURL = payload.ProviderUrl
		} else if payload.CounterPartyAddress != "" {
			providerURL = payload.CounterPartyAddress
		}
	}

	catalog, err := s.client.GetCatalog(r.Context(), providerURL)
	if err != nil {
		http.Error(w, "Failed to query catalog", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(catalog)
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
	json.NewEncoder(w).Encode(map[string]string{"id": negID, "state": "REQUESTED"})
}

func (s *DashboardServer) handleInitiateTransfer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload struct {
		ContractID          string           `json:"contractId"`
		AssetID             string           `json:"assetId"`
		CounterPartyAddress string           `json:"counterPartyAddress"`
		DataDestination     core.DataAddress `json:"dataDestination"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	if payload.DataDestination.Type == "" {
		payload.DataDestination.Type = "HttpProxy"
	}

	transferID, err := s.client.InitiateTransfer(r.Context(), payload.ContractID, payload.AssetID, payload.CounterPartyAddress, payload.DataDestination)
	if err != nil {
		s.logger.Error("Failed to initiate transfer", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"id": transferID, "state": "STARTED"})
}
