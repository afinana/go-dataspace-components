package ports

import (
	"net/http"
)

// RegisterRoutes registers all the consumer module HTTP routes on the provided mux.
func RegisterRoutes(mux *http.ServeMux, callbackHandler *ConsumerCallbackHandler, mgmtHandler *ConsumerManagementHandler) {
	// Callback Routes
	mux.HandleFunc("POST /consumer/negotiations/offers", callbackHandler.HandleContractOffer)
	mux.HandleFunc("GET /consumer/negotiations/{consumerPid}", callbackHandler.HandleGetNegotiation)
	mux.HandleFunc("POST /consumer/negotiations/{consumerPid}/agreement", callbackHandler.HandleContractAgreement)
	mux.HandleFunc("POST /consumer/negotiations/{consumerPid}/events", callbackHandler.HandleNegotiationEvent)
	mux.HandleFunc("POST /consumer/negotiations/{consumerPid}/termination", callbackHandler.HandleNegotiationTermination)

	mux.HandleFunc("GET /consumer/transfers/{consumerPid}", callbackHandler.HandleGetTransfer)
	mux.HandleFunc("POST /consumer/transfers/{consumerPid}/start", callbackHandler.HandleTransferStart)
	mux.HandleFunc("POST /consumer/transfers/{consumerPid}/completion", callbackHandler.HandleTransferCompletion)
	mux.HandleFunc("POST /consumer/transfers/{consumerPid}/suspension", callbackHandler.HandleTransferSuspension)
	mux.HandleFunc("POST /consumer/transfers/{consumerPid}/termination", callbackHandler.HandleTransferTermination)

	// Management API Routes
	mux.HandleFunc("POST /api/consumer/v4/catalog/request", mgmtHandler.HandleCatalogRequest)
	mux.HandleFunc("POST /api/consumer/v4/contractnegotiations", mgmtHandler.HandleInitiateNegotiation)
	mux.HandleFunc("GET /api/consumer/v4/contractnegotiations/{id}", mgmtHandler.HandleGetNegotiation)
	mux.HandleFunc("GET /api/consumer/v4/contractnegotiations/{id}/state", mgmtHandler.HandleGetNegotiationState)
	mux.HandleFunc("GET /api/consumer/v4/contractnegotiations/{id}/agreement", mgmtHandler.HandleGetAgreement)
	mux.HandleFunc("POST /api/consumer/v4/contractnegotiations/request", mgmtHandler.HandleQueryNegotiations)
	mux.HandleFunc("POST /api/consumer/v4/contractnegotiations/{id}/terminate", mgmtHandler.HandleTerminateNegotiation)

	mux.HandleFunc("POST /api/consumer/v4/transferprocesses", mgmtHandler.HandleInitiateTransfer)
	mux.HandleFunc("GET /api/consumer/v4/transferprocesses/{id}", mgmtHandler.HandleGetTransfer)
	mux.HandleFunc("GET /api/consumer/v4/transferprocesses/{id}/state", mgmtHandler.HandleGetTransferState)
	mux.HandleFunc("POST /api/consumer/v4/transferprocesses/request", mgmtHandler.HandleQueryTransfers)
	mux.HandleFunc("POST /api/consumer/v4/transferprocesses/{id}/terminate", mgmtHandler.HandleTerminateTransfer)
	mux.HandleFunc("POST /api/consumer/v4/transferprocesses/{id}/suspend", mgmtHandler.HandleSuspendTransfer)
}
