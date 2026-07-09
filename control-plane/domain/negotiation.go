package domain

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// NegotiationState defines the state of the Contract Negotiation state machine.
type NegotiationState int

const (
	StateRequested NegotiationState = iota
	StateAgreed
	StateVerified
	StateFinalized
	StateTerminated
)

func (s NegotiationState) String() string {
	switch s {
	case StateRequested:
		return "REQUESTED"
	case StateAgreed:
		return "AGREED"
	case StateVerified:
		return "VERIFIED"
	case StateFinalized:
		return "FINALIZED"
	case StateTerminated:
		return "TERMINATED"
	default:
		return "UNKNOWN"
	}
}

// ContractNegotiation tracks the state machine of a contract negotiation between a provider and consumer.
type ContractNegotiation struct {
	ID            string            `json:"id"`
	CorrelationID string            `json:"correlationId"` // External ID used by the peer connector
	CounterParty  string            `json:"counterParty"`  // Peer connector endpoint/DID
	Type          NegotiationType   `json:"type"`          // PROVIDER or CONSUMER role
	State         NegotiationState  `json:"state"`
	ContractOffer *ContractOffer    `json:"contractOffer,omitempty"`
	Agreement     *ContractAgreement `json:"agreement,omitempty"`
	ErrorDetail   string            `json:"errorDetail,omitempty"`
	CreatedAt     time.Time         `json:"createdAt"`
	UpdatedAt     time.Time         `json:"updatedAt"`
}

type NegotiationType string

const (
	TypeConsumer NegotiationType = "CONSUMER"
	TypeProvider NegotiationType = "PROVIDER"
)

// ContractOffer is an offering of a contract linked to an asset.
type ContractOffer struct {
	ID        string    `json:"id"`
	AssetID   string    `json:"assetId"`
	Policy    any       `json:"policy"` // Matches the ODRL Policy structure
	CreatedAt time.Time `json:"createdAt"`
}

// ContractAgreement represents a legally/cryptographically signed contract agreement.
type ContractAgreement struct {
	ID              string    `json:"id"`
	ProviderID      string    `json:"providerId"`
	ConsumerID      string    `json:"consumerId"`
	AssetID         string    `json:"assetId"`
	Policy          any       `json:"policy"` // The agreed ODRL policy
	ContractSigning time.Time `json:"contractSigning"`
	ValidStartDate  time.Time `json:"validStartDate"`
	ValidEndDate    time.Time `json:"validEndDate"`
}

// DSP Protocol Messages mapping to domain structures

// ContractRequestMessage is the message sent to initiate or counter a negotiation.
type ContractRequestMessage struct {
	ID            string         `json:"id"`
	CallbackAddress string       `json:"callbackAddress"`
	Offer          *ContractOffer `json:"offer"`
}

// ContractAgreementMessage is the message confirming the agreed contract terms.
type ContractAgreementMessage struct {
	ID        string             `json:"id"`
	Agreement *ContractAgreement `json:"agreement"`
}

// State Machine transitions rules
var (
	ErrInvalidStateTransition = errors.New("invalid state transition")
)

// Transition transitions the negotiation to a new state if valid under standard EDC rules.
func (cn *ContractNegotiation) Transition(to NegotiationState) error {
	valid := false
	switch cn.State {
	case StateRequested:
		valid = (to == StateAgreed || to == StateTerminated)
	case StateAgreed:
		valid = (to == StateVerified || to == StateFinalized || to == StateTerminated)
	case StateVerified:
		valid = (to == StateFinalized || to == StateTerminated)
	case StateFinalized:
		valid = false // Terminal state
	case StateTerminated:
		valid = false // Terminal state
	}

	if !valid {
		return fmt.Errorf("%w: cannot transition from %s to %s", ErrInvalidStateTransition, cn.State, to)
	}

	if to == StateTerminated && cn.ErrorDetail == "" {
		return fmt.Errorf("%w: cannot transition to TERMINATED state without ErrorDetail being populated", ErrInvalidStateTransition)
	}

	cn.State = to
	cn.UpdatedAt = time.Now()
	return nil
}

// Inbound Ports (Domain Services)

// ContractNegotiationService defines the engine operations coordinating negotiations.
type ContractNegotiationService interface {
	// Initiate starts a new negotiation on the consumer side.
	Initiate(ctx context.Context, peerURL string, offer *ContractOffer) (*ContractNegotiation, error)

	// ProcessRequest processes an incoming request on the provider side.
	ProcessRequest(ctx context.Context, request *ContractRequestMessage) (*ContractNegotiation, error)

	// ProcessAgreement processes an incoming agreement confirmation.
	ProcessAgreement(ctx context.Context, agreementMsg *ContractAgreementMessage) error

	// VerifyAgreement marks a negotiation verified (e.g. key verification or remote signatures checked).
	VerifyAgreement(ctx context.Context, negotiationID string) error

	// GetState retrieves the current state of a negotiation.
	GetState(ctx context.Context, negotiationID string) (NegotiationState, error)
}

// Outbound Ports (Infrastructure)

// ContractNegotiationStore provides the persistence store for negotiation records.
type ContractNegotiationStore interface {
	Save(ctx context.Context, cn *ContractNegotiation) error
	FindByID(ctx context.Context, id string) (*ContractNegotiation, error)
	FindByCorrelationID(ctx context.Context, correlationID string) (*ContractNegotiation, error)
	Update(ctx context.Context, cn *ContractNegotiation) error
}

// PolicyEvaluator validates ODRL policies against participant credentials.
type PolicyEvaluator interface {
	// Evaluate verifies if the client credentials satisfy the constraints in the ODRL policy (e.g. spatial/temporal).
	Evaluate(ctx context.Context, policy any, participantClaims map[string]any) (bool, error)
}
