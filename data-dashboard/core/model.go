package core

import "time"

// Dataset represents a metadata descriptor of registered assets.
type Dataset struct {
	ID            string         `json:"id"`
	Type          string         `json:"type"`
	Title         string         `json:"title"`
	Description   string         `json:"description,omitempty"`
	Version       string         `json:"version,omitempty"`
	Publisher     string         `json:"publisher,omitempty"`
	Keywords      []string       `json:"keywords,omitempty"`
	Distributions []Distribution `json:"distributions"`
}

// Distribution represents physical distribution coordinates.
type Distribution struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Title     string `json:"title"`
	Format    string `json:"format"`
	AccessURL string `json:"accessUrl"`
}

// Catalog represents a collection of Datasets.
type Catalog struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Publisher   string    `json:"publisher"`
	Datasets    []Dataset `json:"datasets"`
}

// ContractNegotiation tracks active negotiation states.
type ContractNegotiation struct {
	ID            string    `json:"id"`
	CorrelationID string    `json:"correlationId"`
	CounterParty  string    `json:"counterParty"`
	State         string    `json:"state"`
	CreatedAt     time.Time `json:"createdAt"`
}

// TransferProcess tracks active transfer states.
type TransferProcess struct {
	ID                  string    `json:"id"`
	ContractAgreementID string    `json:"contractAgreementId"`
	AssetID             string    `json:"assetId"`
	State               string    `json:"state"`
	CreatedAt           time.Time `json:"createdAt"`
}

// VerifiableCredential represents a claim stored in the identity hub.
type VerifiableCredential struct {
	ID                string         `json:"id"`
	Type              []string       `json:"type"`
	Issuer            string         `json:"issuer"`
	IssuanceDate      time.Time      `json:"issuanceDate"`
	CredentialSubject map[string]any `json:"credentialSubject"`
}
