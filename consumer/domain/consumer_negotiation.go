// Package domain defines consumer-specific domain service interfaces.
// It reuses core domain types from control-plane/domain but defines
// consumer-side workflows and port contracts.
package domain

import (
	"context"

	coredomain "github.com/afinana/go-dataspace-components/control-plane/domain"
)

// ConsumerNegotiationService defines consumer-side negotiation operations.
type ConsumerNegotiationService interface {
	// InitiateNegotiation sends a ContractRequestMessage to the provider.
	InitiateNegotiation(ctx context.Context, providerURL string, offer *coredomain.ContractOffer) (*coredomain.ContractNegotiation, error)

	// ProcessOffer handles an incoming ContractOfferMessage from provider.
	ProcessOffer(ctx context.Context, msg *coredomain.ContractOfferMessage) (*coredomain.ContractNegotiation, error)

	// ProcessAgreement handles an incoming ContractAgreementMessage from provider.
	ProcessAgreement(ctx context.Context, msg *coredomain.ContractAgreementMessage) (*coredomain.ContractNegotiation, error)

	// AcceptOffer transitions negotiation to ACCEPTED state.
	AcceptOffer(ctx context.Context, negotiationID string) error

	// VerifyAgreement sends verification to provider.
	VerifyAgreement(ctx context.Context, negotiationID string) error

	// TerminateNegotiation terminates a negotiation with reason.
	TerminateNegotiation(ctx context.Context, negotiationID string, reason string) error

	// GetNegotiation retrieves the current state of a negotiation.
	GetNegotiation(ctx context.Context, negotiationID string) (*coredomain.ContractNegotiation, error)

	// ListNegotiations returns all consumer negotiations.
	ListNegotiations(ctx context.Context) ([]coredomain.ContractNegotiation, error)
}

// ConsumerNegotiationStore provides the persistence store for consumer negotiation records.
// This mirrors the core ContractNegotiationStore but is specific to consumer-side storage.
type ConsumerNegotiationStore interface {
	Save(ctx context.Context, cn *coredomain.ContractNegotiation) error
	FindByID(ctx context.Context, id string) (*coredomain.ContractNegotiation, error)
	FindByCorrelationID(ctx context.Context, correlationID string) (*coredomain.ContractNegotiation, error)
	Update(ctx context.Context, cn *coredomain.ContractNegotiation) error
	ListAll(ctx context.Context) ([]coredomain.ContractNegotiation, error)
}
