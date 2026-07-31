// Package jsonld provides JSON-LD context constants and helpers for DSP 2025-1.
package jsonld

// DSP 2025-1 namespace contexts.
const (
	DSP2025Context = "https://w3id.org/dspace/2025/1/context/"
	DCATContext    = "http://www.w3.org/ns/dcat#"
	ODRLContext    = "http://www.w3.org/ns/odrl/2/"
	EDCMgmtContext = "https://w3id.org/edc/connector/management/v2"
)

// DSP 2025-1 message types.
const (
	TypeCatalogRequestMessage                 = "dspace:CatalogRequestMessage"
	TypeContractRequestMessage                = "dspace:ContractRequestMessage"
	TypeContractOfferMessage                  = "dspace:ContractOfferMessage"
	TypeContractAgreementMessage              = "dspace:ContractAgreementMessage"
	TypeContractAgreementVerificationMessage  = "dspace:ContractAgreementVerificationMessage"
	TypeContractNegotiationEventMessage       = "dspace:ContractNegotiationEventMessage"
	TypeContractNegotiationTerminationMessage = "dspace:ContractNegotiationTerminationMessage"
	TypeTransferRequestMessage                = "dspace:TransferRequestMessage"
	TypeTransferStartMessage                  = "dspace:TransferStartMessage"
	TypeTransferCompletionMessage             = "dspace:TransferCompletionMessage"
	TypeTransferSuspensionMessage             = "dspace:TransferSuspensionMessage"
	TypeTransferTerminationMessage            = "dspace:TransferTerminationMessage"
)

// DSPContextArray returns the standard DSP 2025-1 context array for protocol messages.
func DSPContextArray() []string {
	return []string{DSP2025Context, DCATContext, ODRLContext}
}

// MgmtContextArray returns the context array for management API responses.
func MgmtContextArray() []string {
	return []string{EDCMgmtContext}
}
