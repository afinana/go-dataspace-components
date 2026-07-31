package ports

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"

	"github.com/afinana/go-dataspace-components/control-plane/domain"
	"github.com/afinana/go-dataspace-components/internal/pkg/jsonld"
)

type ConsumerManagementHandler struct {
	logger              *slog.Logger
	tracer              trace.Tracer
	negotiationStore    *PostgresConsumerNegotiationStore
	transferStore       *PostgresConsumerTransferStore
	dspClient           *DSPClient
	providerDSPURL      string
	consumerCallbackURL string
}

func NewConsumerManagementHandler(logger *slog.Logger, tracer trace.Tracer, negStore *PostgresConsumerNegotiationStore, transStore *PostgresConsumerTransferStore, dspClient *DSPClient, providerDSPURL, consumerCallbackURL string) *ConsumerManagementHandler {
	return &ConsumerManagementHandler{
		logger:              logger,
		tracer:              tracer,
		negotiationStore:    negStore,
		transferStore:       transStore,
		dspClient:           dspClient,
		providerDSPURL:      providerDSPURL,
		consumerCallbackURL: consumerCallbackURL,
	}
}

func (h *ConsumerManagementHandler) writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func (h *ConsumerManagementHandler) writeError(w http.ResponseWriter, status int, msg string) {
	h.writeJSON(w, status, map[string]any{"error": msg})
}

// HandleCatalogRequest handles POST /api/consumer/v4/catalog/request
func (h *ConsumerManagementHandler) HandleCatalogRequest(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "ConsumerManagementHandler.HandleCatalogRequest")
	defer span.End()

	resp, err := h.dspClient.SendCatalogRequest(ctx, h.providerDSPURL, "provider")
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.writeJSON(w, http.StatusOK, resp)
}

// HandleInitiateNegotiation handles POST /api/consumer/v4/contractnegotiations
func (h *ConsumerManagementHandler) HandleInitiateNegotiation(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "ConsumerManagementHandler.HandleInitiateNegotiation")
	defer span.End()

	var req struct {
		CounterPartyAddress string                `json:"counterPartyAddress"`
		Offer               *domain.ContractOffer `json:"offer"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	cn := &domain.ContractNegotiation{
		ID:              uuid.NewString(),
		CounterParty:    req.CounterPartyAddress,
		CallbackAddress: h.consumerCallbackURL,
		Type:            domain.TypeConsumer,
		State:           domain.StateRequested,
		ContractOffer:   req.Offer,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := h.negotiationStore.Save(ctx, cn); err != nil {
		h.writeError(w, http.StatusInternalServerError, "failed to save negotiation")
		return
	}

	msg := &domain.ContractRequestMessage{
		ConsumerPID:     cn.ID,
		CallbackAddress: h.consumerCallbackURL,
		Offer:           req.Offer,
	}

	_, err := h.dspClient.SendContractRequest(ctx, req.CounterPartyAddress, msg)
	if err != nil {
		// Log error but we already saved the initial state
		h.logger.ErrorContext(ctx, "failed to send contract request", slog.Any("error", err))
	}

	resp := map[string]any{
		"@context": jsonld.MgmtContextArray(),
		"@type":    "IdResponse",
		"@id":      cn.ID,
	}
	h.writeJSON(w, http.StatusOK, resp)
}

// HandleGetNegotiation handles GET /api/consumer/v4/contractnegotiations/{id}
func (h *ConsumerManagementHandler) HandleGetNegotiation(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "ConsumerManagementHandler.HandleGetNegotiation")
	defer span.End()

	id := r.PathValue("id")
	cn, err := h.negotiationStore.FindByID(ctx, id)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "negotiation not found")
		return
	}
	h.writeJSON(w, http.StatusOK, cn)
}

// HandleGetNegotiationState handles GET /api/consumer/v4/contractnegotiations/{id}/state
func (h *ConsumerManagementHandler) HandleGetNegotiationState(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "ConsumerManagementHandler.HandleGetNegotiationState")
	defer span.End()

	id := r.PathValue("id")
	cn, err := h.negotiationStore.FindByID(ctx, id)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "negotiation not found")
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]string{"state": cn.State.String()})
}

// HandleGetAgreement handles GET /api/consumer/v4/contractnegotiations/{id}/agreement
func (h *ConsumerManagementHandler) HandleGetAgreement(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "ConsumerManagementHandler.HandleGetAgreement")
	defer span.End()

	id := r.PathValue("id")
	cn, err := h.negotiationStore.FindByID(ctx, id)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "negotiation not found")
		return
	}
	if cn.Agreement == nil {
		h.writeError(w, http.StatusNotFound, "agreement not found")
		return
	}
	h.writeJSON(w, http.StatusOK, cn.Agreement)
}

// HandleQueryNegotiations handles POST /api/consumer/v4/contractnegotiations/request
func (h *ConsumerManagementHandler) HandleQueryNegotiations(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "ConsumerManagementHandler.HandleQueryNegotiations")
	defer span.End()

	cns, err := h.negotiationStore.ListAll(ctx)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "failed to query negotiations")
		return
	}
	h.writeJSON(w, http.StatusOK, cns)
}

// HandleTerminateNegotiation handles POST /api/consumer/v4/contractnegotiations/{id}/terminate
func (h *ConsumerManagementHandler) HandleTerminateNegotiation(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "ConsumerManagementHandler.HandleTerminateNegotiation")
	defer span.End()

	id := r.PathValue("id")
	var req struct {
		Reason string `json:"reason"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	cn, err := h.negotiationStore.FindByID(ctx, id)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "negotiation not found")
		return
	}

	cn.ErrorDetail = req.Reason
	if err := cn.Transition(domain.StateTerminated); err != nil {
		h.writeError(w, http.StatusConflict, err.Error())
		return
	}
	h.negotiationStore.Update(ctx, cn)

	h.writeJSON(w, http.StatusOK, map[string]string{"status": "terminated"})
}

// HandleInitiateTransfer handles POST /api/consumer/v4/transferprocesses
func (h *ConsumerManagementHandler) HandleInitiateTransfer(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "ConsumerManagementHandler.HandleInitiateTransfer")
	defer span.End()

	var req struct {
		CounterPartyAddress string              `json:"counterPartyAddress"`
		ContractId          string              `json:"contractId"`
		DataDestination     domain.DataAddress  `json:"dataDestination"`
		AssetId             string              `json:"assetId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	tp := &domain.TransferProcess{
		ID:                  uuid.NewString(),
		ContractAgreementID: req.ContractId,
		CallbackAddress:     h.consumerCallbackURL,
		AssetID:             req.AssetId,
		State:               domain.StateTransferRequested,
		DataDestination:     req.DataDestination,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	if err := h.transferStore.Save(ctx, tp); err != nil {
		h.writeError(w, http.StatusInternalServerError, "failed to save transfer process")
		return
	}

	msg := &domain.TransferRequestMessage{
		ConsumerPID:     tp.ID,
		AgreementID:     tp.ContractAgreementID,
		Format:          "HttpData-PULL",
		DataAddress:     &req.DataDestination,
		CallbackAddress: h.consumerCallbackURL,
	}

	_, err := h.dspClient.SendTransferRequest(ctx, req.CounterPartyAddress, msg)
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to send transfer request", slog.Any("error", err))
	}

	resp := map[string]any{
		"@context": jsonld.MgmtContextArray(),
		"@type":    "IdResponse",
		"@id":      tp.ID,
	}
	h.writeJSON(w, http.StatusOK, resp)
}

// HandleGetTransfer handles GET /api/consumer/v4/transferprocesses/{id}
func (h *ConsumerManagementHandler) HandleGetTransfer(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "ConsumerManagementHandler.HandleGetTransfer")
	defer span.End()

	id := r.PathValue("id")
	tp, err := h.transferStore.FindByID(ctx, id)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "transfer process not found")
		return
	}
	h.writeJSON(w, http.StatusOK, tp)
}

// HandleGetTransferState handles GET /api/consumer/v4/transferprocesses/{id}/state
func (h *ConsumerManagementHandler) HandleGetTransferState(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "ConsumerManagementHandler.HandleGetTransferState")
	defer span.End()

	id := r.PathValue("id")
	tp, err := h.transferStore.FindByID(ctx, id)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "transfer process not found")
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]string{"state": tp.State.String()})
}

// HandleQueryTransfers handles POST /api/consumer/v4/transferprocesses/request
func (h *ConsumerManagementHandler) HandleQueryTransfers(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "ConsumerManagementHandler.HandleQueryTransfers")
	defer span.End()

	tps, err := h.transferStore.ListAll(ctx)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "failed to query transfers")
		return
	}
	h.writeJSON(w, http.StatusOK, tps)
}

// HandleTerminateTransfer handles POST /api/consumer/v4/transferprocesses/{id}/terminate
func (h *ConsumerManagementHandler) HandleTerminateTransfer(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "ConsumerManagementHandler.HandleTerminateTransfer")
	defer span.End()

	id := r.PathValue("id")
	var req struct {
		Reason string `json:"reason"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	tp, err := h.transferStore.FindByID(ctx, id)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "transfer process not found")
		return
	}

	tp.ErrorDetail = req.Reason
	if err := tp.Transition(domain.StateTransferTerminated); err != nil {
		h.writeError(w, http.StatusConflict, err.Error())
		return
	}
	h.transferStore.Update(ctx, tp)

	h.writeJSON(w, http.StatusOK, map[string]string{"status": "terminated"})
}

// HandleSuspendTransfer handles POST /api/consumer/v4/transferprocesses/{id}/suspend
func (h *ConsumerManagementHandler) HandleSuspendTransfer(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "ConsumerManagementHandler.HandleSuspendTransfer")
	defer span.End()

	id := r.PathValue("id")
	var req struct {
		Reason string `json:"reason"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	tp, err := h.transferStore.FindByID(ctx, id)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "transfer process not found")
		return
	}

	// We assume a suspension is requested, usually by provider, but consumer can too.
	if err := tp.Transition(domain.StateTransferSuspended); err != nil {
		h.writeError(w, http.StatusConflict, err.Error())
		return
	}
	h.transferStore.Update(ctx, tp)

	h.writeJSON(w, http.StatusOK, map[string]string{"status": "suspended"})
}
