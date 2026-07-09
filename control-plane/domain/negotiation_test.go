package domain

import (
	"testing"
)

func TestContractNegotiation_Transition(t *testing.T) {
	tests := []struct {
		name        string
		initial     NegotiationState
		target      NegotiationState
		errorDetail string
		expectErr   bool
	}{
		{
			name:      "Requested to Agreed - Valid",
			initial:   StateRequested,
			target:    StateAgreed,
			expectErr: false,
		},
		{
			name:        "Requested to Terminated - Valid with ErrorDetail",
			initial:     StateRequested,
			target:      StateTerminated,
			errorDetail: "some error message",
			expectErr:   false,
		},
		{
			name:        "Requested to Terminated - Invalid without ErrorDetail",
			initial:     StateRequested,
			target:      StateTerminated,
			errorDetail: "",
			expectErr:   true,
		},
		{
			name:      "Agreed to Verified - Valid",
			initial:   StateAgreed,
			target:    StateVerified,
			expectErr: false,
		},
		{
			name:      "Agreed to Finalized - Valid",
			initial:   StateAgreed,
			target:    StateFinalized,
			expectErr: false,
		},
		{
			name:      "Verified to Finalized - Valid",
			initial:   StateVerified,
			target:    StateFinalized,
			expectErr: false,
		},
		{
			name:      "Finalized to Agreed - Invalid",
			initial:   StateFinalized,
			target:    StateAgreed,
			expectErr: true,
		},
		{
			name:      "Terminated to Finalized - Invalid",
			initial:   StateTerminated,
			target:    StateFinalized,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cn := &ContractNegotiation{
				State:       tt.initial,
				ErrorDetail: tt.errorDetail,
			}
			err := cn.Transition(tt.target)
			if (err != nil) != tt.expectErr {
				t.Errorf("ContractNegotiation.Transition() error = %v, expectErr %v", err, tt.expectErr)
			}
			if err == nil && cn.State != tt.target {
				t.Errorf("ContractNegotiation.Transition() state not updated: got %v, expected %v", cn.State, tt.target)
			}
		})
	}
}
