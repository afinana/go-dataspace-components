package core

import "time"

// DataAddress defines raw data source infrastructure parameters.
type DataAddress struct {
	Type             string            `json:"type"`
	BaseURL          string            `json:"baseUrl,omitempty"`
	ProxyQueryParams string            `json:"proxyQueryParams,omitempty"`
	ProxyPath        string            `json:"proxyPath,omitempty"`
	AuthKey          string            `json:"authKey,omitempty"`
	Bucket           string            `json:"bucket,omitempty"`
	Region           string            `json:"region,omitempty"`
	Properties       map[string]string `json:"properties,omitempty"`
}

// AssetEntry aggregates asset metadata and Data Address.
type AssetEntry struct {
	ID          string            `json:"id"`
	Title       string            `json:"title"`
	Version     string            `json:"version,omitempty"`
	ContentType string            `json:"contentType,omitempty"`
	Description string            `json:"description,omitempty"`
	Keywords    []string          `json:"keywords,omitempty"`
	DataAddress DataAddress       `json:"dataAddress"`
	Properties  map[string]string `json:"properties,omitempty"`
	CreatedAt   time.Time         `json:"createdAt"`
}

// Constraint represents an ODRL policy constraint.
type Constraint struct {
	LeftOperand  string `json:"leftOperand"`  // e.g. "spatial"
	Operator     string `json:"operator"`     // "eq", "in", "neq"
	RightOperand string `json:"rightOperand"` // e.g. "https://w3id.org/idsa/code/EU"
}

// PolicyRule defines permissions, prohibitions, or obligations.
type PolicyRule struct {
	Action      string       `json:"action"` // e.g. "USE"
	Constraints []Constraint `json:"constraints,omitempty"`
}

// PolicyDefinition specifies legal or technical usage boundaries.
type PolicyDefinition struct {
	ID           string       `json:"id"`
	Permissions  []PolicyRule `json:"permissions"`
	Prohibitions []PolicyRule `json:"prohibitions"`
	Obligations  []PolicyRule `json:"obligations"`
	CreatedAt    time.Time    `json:"createdAt"`
}

// Criterion represents an asset selector key-value filter.
type Criterion struct {
	OperandLeft  string `json:"operandLeft"`  // e.g. "asset:prop:id"
	Operator     string `json:"operator"`     // "="
	OperandRight string `json:"operandRight"` // e.g. "asset-01-sensor-data"
}

// ContractDefinition links Assets with Access & Contract policies.
type ContractDefinition struct {
	ID               string      `json:"id"`
	AccessPolicyID   string      `json:"accessPolicyId"`
	ContractPolicyID string      `json:"contractPolicyId"`
	AssetsSelector   []Criterion `json:"assetsSelector"`
	CreatedAt        time.Time   `json:"createdAt"`
}

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
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Publisher string    `json:"publisher"`
	Datasets  []Dataset `json:"datasets"`
}

// ContractNegotiation tracks active negotiation states.
type ContractNegotiation struct {
	ID            string    `json:"id"`
	CorrelationID string    `json:"correlationId"`
	CounterParty  string    `json:"counterParty"`
	State         string    `json:"state"` // INITIALIZED, REQUESTED, AGREED, FINALIZED, TERMINATED
	CreatedAt     time.Time `json:"createdAt"`
}

// TransferProcess tracks active transfer states.
type TransferProcess struct {
	ID                  string      `json:"id"`
	ContractAgreementID string      `json:"contractAgreementId"`
	AssetID             string      `json:"assetId"`
	State               string      `json:"state"` // INITIALIZED, REQUESTED, AGREED, STARTED, TERMINATED
	DataDestination     DataAddress `json:"dataDestination"`
	CreatedAt           time.Time   `json:"createdAt"`
}

// VerifiableCredential represents a claim stored in the identity hub.
type VerifiableCredential struct {
	ID                string         `json:"id"`
	Type              []string       `json:"type"`
	Issuer            string         `json:"issuer"`
	IssuanceDate      time.Time      `json:"issuanceDate"`
	CredentialSubject map[string]any `json:"credentialSubject"`
}
