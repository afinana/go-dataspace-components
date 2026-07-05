package domain

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// TransferState defines the current state in the Transfer Process state machine.
type TransferState int

const (
	StateTransferInitial TransferState = iota
	StateTransferRequested
	StateTransferStarting
	StateTransferStarted
	StateTransferCompleted
	StateTransferFailed
)

func (s TransferState) String() string {
	switch s {
	case StateTransferInitial:
		return "INITIAL"
	case StateTransferRequested:
		return "REQUESTED"
	case StateTransferStarting:
		return "STARTING"
	case StateTransferStarted:
		return "STARTED"
	case StateTransferCompleted:
		return "COMPLETED"
	case StateTransferFailed:
		return "FAILED"
	default:
		return "UNKNOWN"
	}
}

// DataAddress models the connection parameters for data sources or destinations.
type DataAddress struct {
	Type       string            `json:"type"` // e.g. "HttpData", "AmazonS3", "GoogleCloudStorage"
	Properties map[string]string `json:"properties"`
}

// GetProperty retrieves a connection property helper.
func (da *DataAddress) GetProperty(key string) string {
	if da.Properties == nil {
		return ""
	}
	return da.Properties[key]
}

// TransferProcess manages the state of a single transfer flow between two connectors.
type TransferProcess struct {
	ID                 string            `json:"id"`
	ContractAgreementID string            `json:"contractAgreementId"`
	CorrelationID      string            `json:"correlationId,omitempty"` // ID of the transfer on the peer side
	AssetID            string            `json:"assetId"`
	State              TransferState     `json:"state"`
	DataDestination    DataAddress       `json:"dataDestination"`
	DataSource         DataAddress       `json:"dataSource,omitempty"` // Available only on the provider side
	ErrorDetail        string            `json:"errorDetail,omitempty"`
	CreatedAt          time.Time         `json:"createdAt"`
	UpdatedAt          time.Time         `json:"updatedAt"`
}

// DSP Protocol Messages mapping to domain structures

// TransferStartMessage signals the start of the data transmission.
type TransferStartMessage struct {
	ID                 string      `json:"id"`
	ProcessID          string      `json:"processId"`
	DataPlaneAddress   string      `json:"dataPlaneAddress,omitempty"`
}

// State Machine transition rules.
var (
	ErrInvalidTransferTransition = errors.New("invalid transfer process transition")
)

// Transition shifts the transfer process to a new state if compliant with standard transfer lifecycle rules.
func (tp *TransferProcess) Transition(to TransferState) error {
	valid := false
	switch tp.State {
	case StateTransferInitial:
		valid = (to == StateTransferRequested || to == StateTransferStarting || to == StateTransferFailed)
	case StateTransferRequested:
		valid = (to == StateTransferStarting || to == StateTransferFailed)
	case StateTransferStarting:
		valid = (to == StateTransferStarted || to == StateTransferFailed)
	case StateTransferStarted:
		valid = (to == StateTransferCompleted || to == StateTransferFailed)
	case StateTransferCompleted:
		valid = false // Terminal state
	case StateTransferFailed:
		valid = false // Terminal state
	}

	if !valid {
		return fmt.Errorf("%w: cannot transition from %s to %s", ErrInvalidTransferTransition, tp.State, to)
	}

	tp.State = to
	tp.UpdatedAt = time.Now()
	return nil
}

// Inbound Ports (Domain Services)

// TransferProcessService coordinates initiating, starting, tracking, and ending transfer runs.
type TransferProcessService interface {
	// InitiateTransfer starts the transfer process request on the consumer side.
	InitiateTransfer(ctx context.Context, agreementID string, destination DataAddress) (*TransferProcess, error)

	// ProcessStart processes an incoming TransferStartMessage on the provider side.
	ProcessStart(ctx context.Context, startMsg *TransferStartMessage) error

	// CompleteTransfer marks the transfer as successfully finished.
	CompleteTransfer(ctx context.Context, processID string) error

	// FailTransfer aborts the transfer due to errors or timeouts.
	FailTransfer(ctx context.Context, processID string, reason string) error
}

// Outbound Ports (Infrastructure)

// TransferProcessStore manages state storage for Transfer Processes.
type TransferProcessStore interface {
	Save(ctx context.Context, tp *TransferProcess) error
	FindByID(ctx context.Context, id string) (*TransferProcess, error)
	FindByCorrelationID(ctx context.Context, correlationID string) (*TransferProcess, error)
	Update(ctx context.Context, tp *TransferProcess) error
}

// DataPlaneSignaler is the port to send START/TERMINATE instructions to the Data Plane.
type DataPlaneSignaler interface {
	// SignalStart commands the data plane to initiate data transmission.
	SignalStart(ctx context.Context, process *TransferProcess) error

	// SignalTerminate commands the data plane to terminate/abort active data transmission.
	SignalTerminate(ctx context.Context, process *TransferProcess) error
}
