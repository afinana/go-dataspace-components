package ports

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/afinana/go-dataspace-components/data-dashboard/core"
	"github.com/afinana/go-dataspace-components/internal/pkg/logging"
)

func setupTestServer(t *testing.T) (*DashboardServer, *http.ServeMux, string) {
	logger := logging.InitLogger("error")
	
	tmpDir, err := os.MkdirTemp("", "dashboard-templates-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	dirs := []string{
		filepath.Join(tmpDir, "core"),
		filepath.Join(tmpDir, "home"),
		filepath.Join(tmpDir, "assets"),
		filepath.Join(tmpDir, "policies"),
		filepath.Join(tmpDir, "contract-definitions"),
		filepath.Join(tmpDir, "catalog"),
		filepath.Join(tmpDir, "transfer"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			t.Fatalf("failed to create dir %s: %v", d, err)
		}
	}

	layoutContent := `{{define "layout"}}{{.Title}} - {{template "content" .Data}}{{end}}`
	indexContent := `{{define "content"}}OK{{end}}`

	os.WriteFile(filepath.Join(tmpDir, "core", "layout.html"), []byte(layoutContent), 0644)
	os.WriteFile(filepath.Join(tmpDir, "core", "style.css"), []byte("/* css */"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "home", "index.html"), []byte(indexContent), 0644)
	os.WriteFile(filepath.Join(tmpDir, "assets", "index.html"), []byte(indexContent), 0644)
	os.WriteFile(filepath.Join(tmpDir, "policies", "index.html"), []byte(indexContent), 0644)
	os.WriteFile(filepath.Join(tmpDir, "contract-definitions", "index.html"), []byte(indexContent), 0644)
	os.WriteFile(filepath.Join(tmpDir, "catalog", "index.html"), []byte(indexContent), 0644)
	os.WriteFile(filepath.Join(tmpDir, "transfer", "index.html"), []byte(indexContent), 0644)

	cfg := &core.DashboardConfig{
		App: core.AppConfig{
			AppTitle:                   "Test EDC Dashboard",
			HealthCheckIntervalSeconds: 30,
			EnableUserConfig:           true,
			MenuItems: []core.MenuItem{
				{Text: "Home", Route: "/"},
				{Text: "Assets", Route: "/assets"},
			},
		},
		Connectors: []core.EdcConfig{
			{
				ID:              "provider-edc",
				Name:            "Provider Node",
				ControlPlaneURL: "http://control-plane:8081",
				CatalogURL:      "http://catalog:8081",
				DataPlaneURL:    "http://data-plane:8082",
				IdentityHubURL:  "http://identity-hub:8080",
			},
		},
	}

	server := NewDashboardServer(logger, cfg, tmpDir)
	mux := http.NewServeMux()
	server.RegisterRoutes(mux)

	return server, mux, tmpDir
}

func TestDashboardServer_Routes(t *testing.T) {
	_, mux, tmpDir := setupTestServer(t)
	defer os.RemoveAll(tmpDir)

	endpoints := []string{
		"/",
		"/assets",
		"/policies",
		"/contract-definitions",
		"/catalog",
		"/transfers",
		"/public/config/app-config.json",
		"/public/config/edc-connector-config.json",
		"/styles.css",
		"/api/connector/health",
	}

	for _, ep := range endpoints {
		req := httptest.NewRequest(http.MethodGet, ep, nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected GET %s to return 200 OK, got %d", ep, rec.Code)
		}
	}
}

func TestDashboardServer_FormsAndAPIs(t *testing.T) {
	_, mux, tmpDir := setupTestServer(t)
	defer os.RemoveAll(tmpDir)

	// Test POST /assets (Form submit)
	form := url.Values{}
	form.Set("id", "asset-test-01")
	form.Set("title", "Test Asset")
	form.Set("version", "1.0.0")
	form.Set("contentType", "application/json")
	form.Set("description", "Unit test asset")
	form.Set("dataType", "HttpData")
	form.Set("baseUrl", "http://test-service/api")

	reqAsset := httptest.NewRequest(http.MethodPost, "/assets", strings.NewReader(form.Encode()))
	reqAsset.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	recAsset := httptest.NewRecorder()
	mux.ServeHTTP(recAsset, reqAsset)

	if recAsset.Code != http.StatusSeeOther {
		t.Errorf("expected POST /assets redirect 303, got %d", recAsset.Code)
	}

	// Test POST /policies (Form submit)
	formPol := url.Values{}
	formPol.Set("id", "policy-test-01")
	formPol.Set("leftOperand", "spatial")
	formPol.Set("operator", "eq")
	formPol.Set("rightOperand", "https://w3id.org/idsa/code/EU")

	reqPol := httptest.NewRequest(http.MethodPost, "/policies", strings.NewReader(formPol.Encode()))
	reqPol.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	recPol := httptest.NewRecorder()
	mux.ServeHTTP(recPol, reqPol)

	if recPol.Code != http.StatusSeeOther {
		t.Errorf("expected POST /policies redirect 303, got %d", recPol.Code)
	}

	// Test POST /contract-definitions (Form submit)
	formCd := url.Values{}
	formCd.Set("id", "cd-test-01")
	formCd.Set("accessPolicyId", "policy-test-01")
	formCd.Set("contractPolicyId", "policy-test-01")
	formCd.Set("assetId", "asset-test-01")

	reqCd := httptest.NewRequest(http.MethodPost, "/contract-definitions", strings.NewReader(formCd.Encode()))
	reqCd.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	recCd := httptest.NewRecorder()
	mux.ServeHTTP(recCd, reqCd)

	if recCd.Code != http.StatusSeeOther {
		t.Errorf("expected POST /contract-definitions redirect 303, got %d", recCd.Code)
	}

	// Test API POST /api/negotiate/start
	negPayload := map[string]string{
		"counterPartyAddress": "http://provider:8282/api/v1/dsp",
		"assetId":             "asset-test-01",
	}
	bodyNeg, _ := json.Marshal(negPayload)
	reqNeg := httptest.NewRequest(http.MethodPost, "/api/negotiate/start", bytes.NewReader(bodyNeg))
	recNeg := httptest.NewRecorder()
	mux.ServeHTTP(recNeg, reqNeg)

	if recNeg.Code != http.StatusOK {
		t.Errorf("expected POST /api/negotiate/start to return 200 OK, got %d", recNeg.Code)
	}

	// Test API POST /api/transfer/start
	transferPayload := map[string]any{
		"contractId":          "contract-01",
		"assetId":             "asset-test-01",
		"counterPartyAddress": "http://provider:8282/api/v1/dsp",
		"dataDestination": map[string]string{
			"type": "HttpProxy",
		},
	}
	bodyTrf, _ := json.Marshal(transferPayload)
	reqTrf := httptest.NewRequest(http.MethodPost, "/api/transfer/start", bytes.NewReader(bodyTrf))
	recTrf := httptest.NewRecorder()
	mux.ServeHTTP(recTrf, reqTrf)

	if recTrf.Code != http.StatusOK {
		t.Errorf("expected POST /api/transfer/start to return 200 OK, got %d", recTrf.Code)
	}
}
