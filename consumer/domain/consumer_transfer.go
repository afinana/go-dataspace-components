package domain

import (
	"context"

	coredomain "github.com/afinana/go-dataspace-components/control-plane/domain"
)

// ConsumerTransferService defines consumer-side transfer operations.
type ConsumerTransferService interface {
	// InitiateTransfer sends a TransferRequestMessage to the provider.
	InitiateTransfer(ctx context.Context, providerURL, agreementID string, dest coredomain.DataAddress) (*coredomain.TransferProcess, error)

	// ProcessStart handles incoming TransferStartMessage with DataAddress/EDR.
	ProcessStart(ctx context.Context, msg *coredomain.TransferStartMessage) error

	// ProcessCompletion handles incoming TransferCompletionMessage.
	ProcessCompletion(ctx context.Context, msg *coredomain.TransferCompletionMessage) error

	// ProcessSuspension handles incoming TransferSuspensionMessage.
	ProcessSuspension(ctx context.Context, msg *coredomain.TransferSuspensionMessage) error

	// TerminateTransfer terminates a transfer with reason.
	TerminateTransfer(ctx context.Context, transferID string, reason string) error

	// SuspendTransfer suspends a transfer.
	SuspendTransfer(ctx context.Context, transferID string) error

	// GetTransfer retrieves the current state of a transfer.
	GetTransfer(ctx context.Context, transferID string) (*coredomain.TransferProcess, error)

	// ListTransfers returns all consumer transfers.
	ListTransfers(ctx context.Context) ([]coredomain.TransferProcess, error)
}

// ConsumerTransferStore provides the persistence store for consumer transfer records.
type ConsumerTransferStore interface {
	Save(ctx context.Context, tp *coredomain.TransferProcess) error
	FindByID(ctx context.Context, id string) (*coredomain.TransferProcess, error)
	FindByCorrelationID(ctx context.Context, correlationID string) (*coredomain.TransferProcess, error)
	Update(ctx context.Context, tp *coredomain.TransferProcess) error
	ListAll(ctx context.Context) ([]coredomain.TransferProcess, error)
}

// DSPOutboundClient defines the outbound DSP protocol client for consumer-side calls.
type DSPOutboundClient interface {
	// SendCatalogRequest forwards a catalog query to the provider.
	SendCatalogRequest(ctx context.Context, providerDSPURL, counterPartyID string) (map[string]any, error)

	// SendContractRequest sends a contract request to initiate negotiation.
	SendContractRequest(ctx context.Context, providerDSPURL string, msg *coredomain.ContractRequestMessage) (map[string]any, error)

	// SendAgreementVerification sends agreement verification to the provider.
	SendAgreementVerification(ctx context.Context, providerDSPURL, providerPID string, msg *coredomain.ContractAgreementVerificationMessage) error

	// SendTransferRequest sends a transfer request to the provider.
	SendTransferRequest(ctx context.Context, providerDSPURL string, msg *coredomain.TransferRequestMessage) (map[string]any, error)
}
