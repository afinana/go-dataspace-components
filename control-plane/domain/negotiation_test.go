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
		// REQUESTED transitions
		{
			name:      "Requested to Agreed - Valid",
			initial:   StateRequested,
			target:    StateAgreed,
			expectErr: false,
		},
		{
			name:      "Requested to Offered - Valid",
			initial:   StateRequested,
			target:    StateOffered,
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
		// OFFERED transitions
		{
			name:      "Offered to Requested - Valid (counter-offer)",
			initial:   StateOffered,
			target:    StateRequested,
			expectErr: false,
		},
		{
			name:      "Offered to Accepted - Valid",
			initial:   StateOffered,
			target:    StateAccepted,
			expectErr: false,
		},
		// ACCEPTED transitions
		{
			name:      "Accepted to Agreed - Valid",
			initial:   StateAccepted,
			target:    StateAgreed,
			expectErr: false,
		},
		// AGREED transitions
		{
			name:      "Agreed to Verified - Valid",
			initial:   StateAgreed,
			target:    StateVerified,
			expectErr: false,
		},
		{
			name:      "Agreed to Finalized - Invalid (must go through Verified)",
			initial:   StateAgreed,
			target:    StateFinalized,
			expectErr: true,
		},
		// VERIFIED transitions
		{
			name:      "Verified to Finalized - Valid",
			initial:   StateVerified,
			target:    StateFinalized,
			expectErr: false,
		},
		// Terminal state checks
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
