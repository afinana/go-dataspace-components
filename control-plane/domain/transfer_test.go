package domain

import (
	"testing"
)

func TestTransferProcess_StateTransitions(t *testing.T) {
	tp := &TransferProcess{
		ID:                  "tp-01",
		ContractAgreementID: "ag-01",
		AssetID:             "asset-01",
		State:               StateTransferInitial,
		DataDestination: DataAddress{
			Type:       "HttpProxy",
			Properties: map[string]string{"endpoint": "http://127.0.0.1:8080"},
		},
	}

	if tp.DataDestination.GetProperty("endpoint") != "http://127.0.0.1:8080" {
		t.Errorf("expected property 'http://127.0.0.1:8080'")
	}
	if tp.DataDestination.GetProperty("nonexistent") != "" {
		t.Errorf("expected empty string for nonexistent property")
	}

	var emptyAddr DataAddress
	if emptyAddr.GetProperty("key") != "" {
		t.Errorf("expected empty string when Properties map is nil")
	}

	states := []TransferState{
		StateTransferInitial,
		StateTransferRequested,
		StateTransferStarting,
		StateTransferStarted,
		StateTransferCompleted,
		StateTransferTerminated,
		TransferState(99),
	}
	expectedStrings := []string{
		"INITIAL",
		"REQUESTED",
		"STARTING",
		"STARTED",
		"COMPLETED",
		"TERMINATED",
		"UNKNOWN",
	}

	for i, s := range states {
		if s.String() != expectedStrings[i] {
			t.Errorf("expected state string %s, got %s", expectedStrings[i], s.String())
		}
	}

	// Test valid transitions: Initial -> Requested -> Starting -> Started -> Completed
	if err := tp.Transition(StateTransferRequested); err != nil {
		t.Errorf("failed transition Initial -> Requested: %v", err)
	}
	if err := tp.Transition(StateTransferStarting); err != nil {
		t.Errorf("failed transition Requested -> Starting: %v", err)
	}
	if err := tp.Transition(StateTransferStarted); err != nil {
		t.Errorf("failed transition Starting -> Started: %v", err)
	}
	if err := tp.Transition(StateTransferCompleted); err != nil {
		t.Errorf("failed transition Started -> Completed: %v", err)
	}

	// Completed is terminal
	if err := tp.Transition(StateTransferTerminated); err == nil {
		t.Errorf("expected error transitioning from terminal state Completed")
	}

	// Test termination without ErrorDetail
	tp2 := &TransferProcess{State: StateTransferRequested}
	if err := tp2.Transition(StateTransferTerminated); err == nil {
		t.Errorf("expected error transitioning to Terminated without ErrorDetail")
	}
	tp2.ErrorDetail = "Timeout occurred"
	if err := tp2.Transition(StateTransferTerminated); err != nil {
		t.Errorf("failed transition to Terminated with ErrorDetail: %v", err)
	}
}
