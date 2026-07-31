package ports

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"go.opentelemetry.io/otel/trace"

	"github.com/afinana/go-dataspace-components/control-plane/domain"
)

type ConsumerCallbackHandler struct {
	logger           *slog.Logger
	tracer           trace.Tracer
	negotiationStore *PostgresConsumerNegotiationStore
	transferStore    *PostgresConsumerTransferStore
}

func NewConsumerCallbackHandler(logger *slog.Logger, tracer trace.Tracer, negStore *PostgresConsumerNegotiationStore, transStore *PostgresConsumerTransferStore) *ConsumerCallbackHandler {
	return &ConsumerCallbackHandler{
		logger:           logger,
		tracer:           tracer,
		negotiationStore: negStore,
		transferStore:    transStore,
	}
}

func (h *ConsumerCallbackHandler) writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func (h *ConsumerCallbackHandler) writeError(w http.ResponseWriter, status int, msg string) {
	h.writeJSON(w, status, map[string]any{"error": msg})
}

// HandleContractOffer handles POST /consumer/negotiations/offers
func (h *ConsumerCallbackHandler) HandleContractOffer(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "ConsumerCallbackHandler.HandleContractOffer")
	defer span.End()

	var msg domain.ContractOfferMessage
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	h.logger.InfoContext(ctx, "received contract offer", slog.String("consumerPid", msg.ConsumerPID))

	// In a full implementation, we'd find the negotiation and update it, but it might be new if provider initiated
	// Here we assume it already exists and was requested by consumer
	cn, err := h.negotiationStore.FindByID(ctx, msg.ConsumerPID)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "negotiation not found")
		return
	}

	if err := cn.Transition(domain.StateOffered); err != nil {
		h.writeError(w, http.StatusConflict, err.Error())
		return
	}

	cn.ContractOffer = msg.Offer
	cn.CorrelationID = msg.ProviderPID
	if msg.CallbackAddress != "" {
		cn.CallbackAddress = msg.CallbackAddress
	}

	if err := h.negotiationStore.Update(ctx, cn); err != nil {
		h.writeError(w, http.StatusInternalServerError, "failed to update negotiation")
		return
	}

	w.WriteHeader(http.StatusOK)
}

// HandleGetNegotiation handles GET /consumer/negotiations/{consumerPid}
func (h *ConsumerCallbackHandler) HandleGetNegotiation(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "ConsumerCallbackHandler.HandleGetNegotiation")
	defer span.End()

	pid := r.PathValue("consumerPid")
	cn, err := h.negotiationStore.FindByID(ctx, pid)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "negotiation not found")
		return
	}

	h.writeJSON(w, http.StatusOK, cn)
}

// HandleContractAgreement handles POST /consumer/negotiations/{consumerPid}/agreement
func (h *ConsumerCallbackHandler) HandleContractAgreement(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "ConsumerCallbackHandler.HandleContractAgreement")
	defer span.End()

	pid := r.PathValue("consumerPid")
	var msg domain.ContractAgreementMessage
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cn, err := h.negotiationStore.FindByID(ctx, pid)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "negotiation not found")
		return
	}

	if err := cn.Transition(domain.StateAgreed); err != nil {
		h.writeError(w, http.StatusConflict, err.Error())
		return
	}

	cn.Agreement = msg.Agreement
	if err := h.negotiationStore.Update(ctx, cn); err != nil {
		h.writeError(w, http.StatusInternalServerError, "failed to update negotiation")
		return
	}

	w.WriteHeader(http.StatusOK)
}

// HandleNegotiationEvent handles POST /consumer/negotiations/{consumerPid}/events
func (h *ConsumerCallbackHandler) HandleNegotiationEvent(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "ConsumerCallbackHandler.HandleNegotiationEvent")
	defer span.End()

	pid := r.PathValue("consumerPid")
	var msg domain.ContractNegotiationEventMessage
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cn, err := h.negotiationStore.FindByID(ctx, pid)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "negotiation not found")
		return
	}

	var targetState domain.NegotiationState
	switch msg.EventType {
	case "ACCEPTED":
		targetState = domain.StateAccepted
	case "FINALIZED":
		targetState = domain.StateFinalized
	default:
		h.writeError(w, http.StatusBadRequest, "unknown event type")
		return
	}

	if err := cn.Transition(targetState); err != nil {
		h.writeError(w, http.StatusConflict, err.Error())
		return
	}

	if err := h.negotiationStore.Update(ctx, cn); err != nil {
		h.writeError(w, http.StatusInternalServerError, "failed to update negotiation")
		return
	}

	w.WriteHeader(http.StatusOK)
}

// HandleNegotiationTermination handles POST /consumer/negotiations/{consumerPid}/termination
func (h *ConsumerCallbackHandler) HandleNegotiationTermination(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "ConsumerCallbackHandler.HandleNegotiationTermination")
	defer span.End()

	pid := r.PathValue("consumerPid")
	var msg domain.ContractNegotiationTerminationMessage
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cn, err := h.negotiationStore.FindByID(ctx, pid)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "negotiation not found")
		return
	}

	cn.ErrorDetail = msg.Code
	if err := cn.Transition(domain.StateTerminated); err != nil {
		h.writeError(w, http.StatusConflict, err.Error())
		return
	}

	if err := h.negotiationStore.Update(ctx, cn); err != nil {
		h.writeError(w, http.StatusInternalServerError, "failed to update negotiation")
		return
	}

	w.WriteHeader(http.StatusOK)
}

// HandleGetTransfer handles GET /consumer/transfers/{consumerPid}
func (h *ConsumerCallbackHandler) HandleGetTransfer(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "ConsumerCallbackHandler.HandleGetTransfer")
	defer span.End()

	pid := r.PathValue("consumerPid")
	tp, err := h.transferStore.FindByID(ctx, pid)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "transfer process not found")
		return
	}

	h.writeJSON(w, http.StatusOK, tp)
}

// HandleTransferStart handles POST /consumer/transfers/{consumerPid}/start
func (h *ConsumerCallbackHandler) HandleTransferStart(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "ConsumerCallbackHandler.HandleTransferStart")
	defer span.End()

	pid := r.PathValue("consumerPid")
	var msg domain.TransferStartMessage
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tp, err := h.transferStore.FindByID(ctx, pid)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "transfer process not found")
		return
	}

	if err := tp.Transition(domain.StateTransferStarted); err != nil {
		h.writeError(w, http.StatusConflict, err.Error())
		return
	}

	if err := h.transferStore.Update(ctx, tp); err != nil {
		h.writeError(w, http.StatusInternalServerError, "failed to update transfer process")
		return
	}

	w.WriteHeader(http.StatusOK)
}

// HandleTransferCompletion handles POST /consumer/transfers/{consumerPid}/completion
func (h *ConsumerCallbackHandler) HandleTransferCompletion(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "ConsumerCallbackHandler.HandleTransferCompletion")
	defer span.End()

	pid := r.PathValue("consumerPid")
	var msg domain.TransferCompletionMessage
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tp, err := h.transferStore.FindByID(ctx, pid)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "transfer process not found")
		return
	}

	if err := tp.Transition(domain.StateTransferCompleted); err != nil {
		h.writeError(w, http.StatusConflict, err.Error())
		return
	}

	if err := h.transferStore.Update(ctx, tp); err != nil {
		h.writeError(w, http.StatusInternalServerError, "failed to update transfer process")
		return
	}

	w.WriteHeader(http.StatusOK)
}

// HandleTransferSuspension handles POST /consumer/transfers/{consumerPid}/suspension
func (h *ConsumerCallbackHandler) HandleTransferSuspension(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "ConsumerCallbackHandler.HandleTransferSuspension")
	defer span.End()

	pid := r.PathValue("consumerPid")
	var msg domain.TransferSuspensionMessage
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tp, err := h.transferStore.FindByID(ctx, pid)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "transfer process not found")
		return
	}

	if err := tp.Transition(domain.StateTransferSuspended); err != nil {
		h.writeError(w, http.StatusConflict, err.Error())
		return
	}

	if err := h.transferStore.Update(ctx, tp); err != nil {
		h.writeError(w, http.StatusInternalServerError, "failed to update transfer process")
		return
	}

	w.WriteHeader(http.StatusOK)
}

// HandleTransferTermination handles POST /consumer/transfers/{consumerPid}/termination
func (h *ConsumerCallbackHandler) HandleTransferTermination(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "ConsumerCallbackHandler.HandleTransferTermination")
	defer span.End()

	pid := r.PathValue("consumerPid")
	var msg domain.TransferTerminationMessage
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tp, err := h.transferStore.FindByID(ctx, pid)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "transfer process not found")
		return
	}

	tp.ErrorDetail = msg.Code
	if err := tp.Transition(domain.StateTransferTerminated); err != nil {
		h.writeError(w, http.StatusConflict, err.Error())
		return
	}

	if err := h.transferStore.Update(ctx, tp); err != nil {
		h.writeError(w, http.StatusInternalServerError, "failed to update transfer process")
		return
	}

	w.WriteHeader(http.StatusOK)
}
