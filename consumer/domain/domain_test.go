package domain_test

import (
	"testing"

	"github.com/afinana/go-dataspace-components/consumer/domain"
)

func TestDomainInterfaces(t *testing.T) {
	// Verify package compiles and domain definitions exist
	var _ domain.ConsumerNegotiationService = nil
	var _ domain.ConsumerTransferService = nil
}
