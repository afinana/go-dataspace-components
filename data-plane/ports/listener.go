package ports

import (
	"encoding/json"
	"log/slog"
	"net/http"

	dp "github.com/afinana/go-dataspace-components/data-plane/domain"
)

// SignalingListener handles HTTP control signals from the Control Plane to the Data Plane.
// Conforms to the standard EDC Data Plane Signaling Protocol.
type SignalingListener struct {
	logger      *slog.Logger
	controllers []dp.DataFlowController
}

// NewSignalingListener initializes the signaling HTTP endpoint listener.
func NewSignalingListener(logger *slog.Logger, controllers []dp.DataFlowController) *SignalingListener {
	return &SignalingListener{
		logger:      logger,
		controllers: controllers,
	}
}

// RegisterRoutes registers signaling routes onto a standard http.ServeMux or chi router.
func (l *SignalingListener) RegisterRoutes(mux *http.ServeMux) {
	// Standard EDC Data Plane Signaling Protocol endpoints
	mux.HandleFunc("POST /v1/dataflows/start", l.handleStart)
	mux.HandleFunc("POST /v1/dataflows/{id}/terminate", l.handleTerminate)

	// Backward compatibility endpoints
	mux.HandleFunc("POST /signaling/start", l.handleStart)
	mux.HandleFunc("POST /signaling/terminate", l.handleTerminate)
}

func (l *SignalingListener) handleStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var request dp.DataFlowRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		l.logger.Error("Failed to decode data flow start request", "err", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	l.logger.Info("Received signaling START command from Control Plane", "transferId", request.ID)

	// Delegate to the controller that can handle the request
	var handled bool
	for _, ctrl := range l.controllers {
		if ctrl.CanHandle(&request) {
			response, err := ctrl.Initiate(r.Context(), &request)
			if err != nil {
				l.logger.Error("Failed to initiate data flow", "transferId", request.ID, "err", err)
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(dp.DataFlowResponse{Success: false, ErrorDetail: err.Error()})
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
			handled = true
			break
		}
	}

	if !handled {
		l.logger.Warn("No suitable controller found for data flow request", "transferId", request.ID)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(dp.DataFlowResponse{
			Success:     false,
			ErrorDetail: "No suitable DataFlowController found for source/destination combo",
		})
	}
}

func (l *SignalingListener) handleTerminate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		ID     string `json:"id"`
		Reason string `json:"reason"`
	}

	// Read and decode the body if present (errors ignored for path-based parameter fallback)
	_ = json.NewDecoder(r.Body).Decode(&request)

	id := r.PathValue("id")
	if id == "" {
		id = request.ID
	}

	if id == "" {
		l.logger.Error("Failed to extract data flow ID from request path or body")
		http.Error(w, "Bad Request: missing flow ID", http.StatusBadRequest)
		return
	}

	l.logger.Info("Received signaling TERMINATE command from Control Plane", "transferId", id, "reason", request.Reason)

	// In a real data plane, this signals active workers/pipes to close contexts.
	// For this scaffold, we acknowledge the termination signal successfully.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"message": "Data flow terminated successfully",
	})
}
