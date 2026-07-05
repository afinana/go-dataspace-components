package domain

import (
	"context"
	"time"
)

// DIDDocument represents a Decentralized Identifier Document compliant with W3C standards.
type DIDDocument struct {
	ID                 string               `json:"id"`
	Context            []string             `json:"@context"`
	VerificationMethod []VerificationMethod `json:"verificationMethod,omitempty"`
	Authentication     []string             `json:"authentication,omitempty"`
	Service            []DIDService         `json:"service,omitempty"`
}

// VerificationMethod defines public keys linked to a DID.
type VerificationMethod struct {
	ID           string `json:"id"`
	Type         string `json:"type"`
	Controller   string `json:"controller"`
	PublicKeyJwk any    `json:"publicKeyJwk,omitempty"`
}

// DIDService defines endpoints linked to the DID for interaction (e.g., Credential Issuance, Data Planes).
type DIDService struct {
	ID              string   `json:"id"`
	Type            string   `json:"type"`
	ServiceEndpoint string   `json:"serviceEndpoint"`
	Properties      Metadata `json:"properties,omitempty"`
}

// Metadata represents unstructured property values.
type Metadata map[string]any

// VerifiableCredential represents a signed assertion about a subject.
type VerifiableCredential struct {
	Context           []string       `json:"@context"`
	ID                string         `json:"id,omitempty"`
	Type              []string       `json:"type"`
	Issuer            string         `json:"issuer"`
	IssuanceDate      time.Time      `json:"issuanceDate"`
	ExpirationDate    *time.Time     `json:"expirationDate,omitempty"`
	CredentialSubject Metadata       `json:"credentialSubject"`
	Proof             *Proof         `json:"proof,omitempty"`
}

// Proof contains cryptographic signature details for W3C credentials.
type Proof struct {
	Type               string    `json:"type"`
	Created            time.Time `json:"created"`
	VerificationMethod string    `json:"verificationMethod"`
	ProofPurpose       string    `json:"proofPurpose"`
	ProofValue         string    `json:"proofValue"`
}

// VerifiablePresentation represents a container containing one or more credentials wrapped for exchange.
type VerifiablePresentation struct {
	Context              []string               `json:"@context"`
	ID                   string                 `json:"id,omitempty"`
	Type                 []string               `json:"type"`
	VerifiableCredential []VerifiableCredential `json:"verifiableCredential"`
	Proof                *Proof                 `json:"proof,omitempty"`
}

// PresentationDefinition defines the query schema for credentials.
type PresentationDefinition struct {
	ID    string         `json:"id"`
	Input []InputDescriptor `json:"input_descriptors"`
}

// InputDescriptor specifies criteria for matching credentials.
type InputDescriptor struct {
	ID          string          `json:"id"`
	Purpose     string          `json:"purpose,omitempty"`
	Constraints *Constraints    `json:"constraints,omitempty"`
}

// Constraints represent limits on fields or schema paths.
type Constraints struct {
	Fields []FieldConstraint `json:"fields,omitempty"`
}

// FieldConstraint represents specific VC property constraints.
type FieldConstraint struct {
	Path   []string `json:"path"`
	Filter any      `json:"filter,omitempty"`
}

// Ports (Interfaces) for the Identity Hub Component.

// DIDResolver resolves W3C DIDs into DID documents.
type DIDResolver interface {
	Resolve(ctx context.Context, did string) (*DIDDocument, error)
}

// CredentialVerifier validates Verifiable Credentials or Presentations.
type CredentialVerifier interface {
	VerifyCredential(ctx context.Context, vc *VerifiableCredential) (bool, error)
	VerifyPresentation(ctx context.Context, vp *VerifiablePresentation) (bool, error)
}

// PresentationExchange manages the negotiation and proof requests (W3C Presentation Exchange).
type PresentationExchange interface {
	CreatePresentationQuery(ctx context.Context, definition *PresentationDefinition) (*VerifiablePresentation, error)
	ValidatePresentation(ctx context.Context, vp *VerifiablePresentation, definition *PresentationDefinition) (bool, error)
}
