package ports

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/afinana/go-dataspace-components/identity-hub/domain"
	"github.com/afinana/go-dataspace-components/internal/pkg/logging"
)

func TestPresentationAPIHandler_CompatibilityRoutes(t *testing.T) {
	logger := logging.InitLogger("DEBUG")
	handler := NewPresentationAPIHandler(logger, nil)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	// 1. Test GET /api/identity/v1alpha/dids
	req := httptest.NewRequest(http.MethodGet, "/api/identity/v1alpha/dids", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected GET /api/identity/v1alpha/dids status 200, got %d", rr.Code)
	}

	var dids []map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &dids); err != nil {
		t.Fatalf("failed to parse dids response: %v", err)
	}
	if len(dids) != 1 || dids[0]["did"] != "did:web:local-connector" {
		t.Errorf("unexpected dids payload: %v", dids)
	}

	// 2. Test POST /api/identity/v1alpha/dids
	req = httptest.NewRequest(http.MethodPost, "/api/identity/v1alpha/dids", nil)
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected POST /api/identity/v1alpha/dids status 201, got %d", rr.Code)
	}

	// 3. Test GET /api/identity/v1alpha/credentials
	req = httptest.NewRequest(http.MethodGet, "/api/identity/v1alpha/credentials", nil)
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected GET /api/identity/v1alpha/credentials status 200, got %d", rr.Code)
	}

	var creds []domain.VerifiableCredential
	if err := json.Unmarshal(rr.Body.Bytes(), &creds); err != nil {
		t.Fatalf("failed to parse credentials list: %v", err)
	}
	if len(creds) == 0 {
		t.Error("expected credentials list to contain fallback mock credentials, got empty list")
	}

	// 4. Test GET /api/identity/v1alpha/participants/12345/credentials
	req = httptest.NewRequest(http.MethodGet, "/api/identity/v1alpha/participants/12345/credentials", nil)
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected GET participant credentials status 200, got %d", rr.Code)
	}
}

func TestPresentationAPIHandler_DCPRoutes(t *testing.T) {
	logger := logging.InitLogger("DEBUG")
	handler := NewPresentationAPIHandler(logger, nil)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	// 1. Ingest a credential
	vc := domain.VerifiableCredential{
		ID:           "test-vc-id",
		Type:         []string{"VerifiableCredential", "XDataShareMembershipCredential"},
		Issuer:       "did:web:sovereign-authority.org",
		CredentialSubject: map[string]any{
			"holder": "did:web:local-connector",
		},
	}
	bodyBytes, _ := json.Marshal(vc)
	req := httptest.NewRequest(http.MethodPost, "/credentials", bytes.NewReader(bodyBytes))
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected POST /credentials status 201, got %d", rr.Code)
	}

	// 2. Query the presentation
	query := map[string]any{
		"scopes": []string{"org.eclipse.dspace.dcp.vc.type:XDataShareMembershipCredential"},
	}
	queryBytes, _ := json.Marshal(query)
	req = httptest.NewRequest(http.MethodPost, "/presentations/query", bytes.NewReader(queryBytes))
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected POST /presentations/query status 200, got %d", rr.Code)
	}

	var vp domain.VerifiablePresentation
	if err := json.Unmarshal(rr.Body.Bytes(), &vp); err != nil {
		t.Fatalf("failed to decode verifiable presentation: %v", err)
	}
	if len(vp.VerifiableCredential) != 1 || vp.VerifiableCredential[0].ID != "test-vc-id" {
		t.Errorf("unexpected query result: %+v", vp)
	}
}

func TestPresentationAPIHandler_MethodNotAllowed(t *testing.T) {
	logger := logging.InitLogger("DEBUG")
	handler := NewPresentationAPIHandler(logger, nil)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	// POST to query with GET should return Method Not Allowed
	req := httptest.NewRequest(http.MethodGet, "/presentations/query", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405 for GET query, got %d", rr.Code)
	}

	// GET to credentials ingestion should return Method Not Allowed
	req = httptest.NewRequest(http.MethodGet, "/credentials", nil)
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405 for GET credentials ingestion, got %d", rr.Code)
	}
}

func TestPresentationAPIHandler_InvalidPayloads(t *testing.T) {
	logger := logging.InitLogger("DEBUG")
	handler := NewPresentationAPIHandler(logger, nil)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	// Send broken JSON to /credentials
	req := httptest.NewRequest(http.MethodPost, "/credentials", io.NopCloser(bytes.NewReader([]byte("{invalid-json"))))
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for invalid credential json, got %d", rr.Code)
	}

	// Send broken JSON to /presentations/query
	req = httptest.NewRequest(http.MethodPost, "/presentations/query", io.NopCloser(bytes.NewReader([]byte("{invalid-json"))))
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for invalid query json, got %d", rr.Code)
	}
}
