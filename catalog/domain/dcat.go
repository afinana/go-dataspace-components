package domain

import (
	"context"
	"time"
)

// Catalog represents a W3C DCAT-AP compliant Data Catalog.
type Catalog struct {
	ID          string        `json:"id"`
	Type        string        `json:"type"` // e.g. "dcat:Catalog"
	Title       string        `json:"title"`
	Description string        `json:"description,omitempty"`
	Publisher   string        `json:"publisher"` // URI of publishing entity (DID/URL)
	Datasets    []Dataset     `json:"datasets,omitempty"`
	Services    []DataService `json:"services,omitempty"`
}

// Dataset represents a conceptual dataset published in the catalog.
type Dataset struct {
	ID            string         `json:"id"`
	Type          string         `json:"type"` // e.g. "dcat:Dataset"
	Title         string         `json:"title"`
	Description   string         `json:"description,omitempty"`
	Version       string         `json:"version,omitempty"`
	Keywords      []string       `json:"keywords,omitempty"`
	Publisher     string         `json:"publisher,omitempty"`
	Issued        *time.Time     `json:"issued,omitempty"`
	Modified      *time.Time     `json:"modified,omitempty"`
	Distributions []Distribution `json:"distributions"`
}

// Distribution represents a concrete representation of a Dataset (e.g., REST API, File stream).
type Distribution struct {
	ID             string       `json:"id"`
	Type           string       `json:"type"` // e.g. "dcat:Distribution"
	Title          string       `json:"title"`
	Format         string       `json:"format"` // e.g. "application/json", "application/octet-stream"
	AccessURL      string       `json:"accessUrl"`
	DownloadURL    string       `json:"downloadUrl,omitempty"`
	DataServiceRef string       `json:"dataService,omitempty"` // ID references to a DataService if applicable
	Policy         *ODRLPolicy  `json:"policy,omitempty"`       // ODRL usage terms/contracts linked to this distribution
}

// DataService represents operations/endpoints providing access to datasets.
type DataService struct {
	ID             string   `json:"id"`
	Type           string   `json:"type"` // e.g. "dcat:DataService"
	Title          string   `json:"title"`
	EndpointURL    string   `json:"endpointUrl"`
	EndpointType   string   `json:"endpointDescription,omitempty"` // e.g. "http-api", "grpc", "s3"
	ServedDatasets []string `json:"servedDatasets,omitempty"`      // List of Dataset IDs served
}

// ODRLPolicy models Open Digital Rights Language policies attached to assets.
type ODRLPolicy struct {
	ID         string       `json:"id"`
	Type       string       `json:"type"` // e.g. "odrl:Offer", "odrl:Agreement"
	Target     string       `json:"target"`
	Assigner   string       `json:"assigner,omitempty"`
	Assignee   string       `json:"assignee,omitempty"`
	Permission []Permission `json:"permission,omitempty"`
	Prohibition []Prohibition `json:"prohibition,omitempty"`
	Obligation  []Duty       `json:"obligation,omitempty"`
}

type Permission struct {
	Action     string       `json:"action"` // e.g. "odrl:use", "odrl:read"
	Constraint []Constraint `json:"constraint,omitempty"`
}

type Prohibition struct {
	Action     string       `json:"action"`
	Constraint []Constraint `json:"constraint,omitempty"`
}

type Duty struct {
	Action     string       `json:"action"`
	Constraint []Constraint `json:"constraint,omitempty"`
}

type Constraint struct {
	LeftOperand  string `json:"leftOperand"`  // e.g. "spatial", "dateTime"
	Operator     string `json:"operator"`     // e.g. "eq", "lt", "odrl:in"
	RightOperand string `json:"rightOperand"` // e.g. "EU", "2026-12-31T23:59:59Z"
}

// Ports (Interfaces) for the Catalog Component.

// AssetRegistry manages localized storage and registration of assets.
type AssetRegistry interface {
	RegisterDataset(ctx context.Context, dataset *Dataset) error
	GetDataset(ctx context.Context, id string) (*Dataset, error)
	DeleteDataset(ctx context.Context, id string) error
	ListDatasets(ctx context.Context) ([]Dataset, error)
}

// CatalogQueryService handles querying catalogs for users or external connectors.
type CatalogQueryService interface {
	GetCatalog(ctx context.Context, requesterID string) (*Catalog, error)
	QueryDatasets(ctx context.Context, filter map[string]string) ([]Dataset, error)
}
