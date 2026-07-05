package ports

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/afinana/go-dataspace-components/identity-hub/domain"
)

// PresentationAPIHandler handles the DCP (Decentralized Claims Protocol) public Credentials API.
type PresentationAPIHandler struct {
	mu          sync.RWMutex
	logger      *slog.Logger
	credentials map[string]domain.VerifiableCredential
}

// NewPresentationAPIHandler initializes the presentation and credential HTTP API handler.
func NewPresentationAPIHandler(logger *slog.Logger) *PresentationAPIHandler {
	return &PresentationAPIHandler{
		logger:      logger,
		credentials: make(map[string]domain.VerifiableCredential),
	}
}

// RegisterRoutes sets up routes for the DCP Credentials API.
func (h *PresentationAPIHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/presentations/query", h.handleQuery)
	mux.HandleFunc("/credentials", h.handleCredentials)
}

// QueryRequest models the request structure for POST /presentations/query.
type QueryRequest struct {
	Scopes []string `json:"scopes"` // e.g. ["org.eclipse.dspace.dcp.vc.type:XDataShareMembershipCredential"]
}

// handleQuery processes POST /presentations/query to filter and return Verifiable Presentations.
func (h *PresentationAPIHandler) handleQuery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var req QueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to parse query request payload", "err", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	h.logger.Info("Received Verifiable Presentation query request", "scopes", req.Scopes)

	// Filter VCs matching the requested scopes
	var matchedVCs []domain.VerifiableCredential
	h.mu.RLock()
	for _, vc := range h.credentials {
		for _, scope := range req.Scopes {
			// A scope can look like "org.eclipse.dspace.dcp.vc.type:XDataShareMembershipCredential"
			// Check if any of the VC types match the suffix of the scope
			for _, vcType := range vc.Type {
				if vcType == scope || (scope != "" && (scope == vcType || (len(scope) > len(vcType) && scope[len(scope)-len(vcType):] == vcType))) {
					matchedVCs = append(matchedVCs, vc)
					break
				}
			}
		}
	}
	h.mu.RUnlock()

	// Build the Verifiable Presentation envelope
	vp := domain.VerifiablePresentation{
		Context: []string{
			"https://www.w3.org/2018/credentials/v1",
			"https://w3id.org/security/suites/jws-2020/v1",
		},
		Type:                 []string{"VerifiablePresentation"},
		VerifiableCredential: matchedVCs,
		Proof: &domain.Proof{
			Type:               "JsonWebSignature2020",
			Created:            time.Now(),
			VerificationMethod: "did:web:local-connector#key-1",
			ProofPurpose:       "assertionMethod",
			ProofValue:         "mock-jws-signature-envelope-value",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(vp)
}

// handleCredentials processes POST /credentials to store new credentials.
func (h *PresentationAPIHandler) handleCredentials(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var vc domain.VerifiableCredential
	if err := json.NewDecoder(r.Body).Decode(&vc); err != nil {
		h.logger.Error("Failed to parse credential body", "err", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	if vc.ID == "" {
		vc.ID = "vc-gen-" + time.Now().Format("20060102150405")
	}

	h.logger.Info("Received Verifiable Credential update/store request", "vcId", vc.ID, "issuer", vc.Issuer)

	h.mu.Lock()
	h.credentials[vc.ID] = vc
	h.mu.Unlock()

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{
		"success":      true,
		"credentialId": vc.ID,
		"message":      "Verifiable Credential saved successfully",
	})
}
