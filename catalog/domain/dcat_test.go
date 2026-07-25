package domain

import (
	"testing"
	"time"
)

func TestCatalog_DomainModels(t *testing.T) {
	now := time.Now()
	cat := Catalog{
		ID:          "catalog-01",
		Type:        "dcat:Catalog",
		Title:       "Test Catalog",
		Description: "Unit Test Catalog",
		Publisher:   "did:web:example.com",
		Datasets: []Dataset{
			{
				ID:          "ds-01",
				Type:        "dcat:Dataset",
				Title:       "Test Dataset",
				Description: "Sample Data",
				Version:     "1.0.0",
				Keywords:    []string{"test", "data"},
				Publisher:   "did:web:example.com",
				Issued:      &now,
				Modified:    &now,
				Distributions: []Distribution{
					{
						ID:             "dist-01",
						Type:           "dcat:Distribution",
						Title:          "REST API",
						Format:         "application/json",
						AccessURL:      "http://127.0.0.1:8080/api",
						DownloadURL:    "http://127.0.0.1:8080/download",
						DataServiceRef: "ds-ref-01",
						Policy: &ODRLPolicy{
							ID:       "pol-01",
							Type:     "odrl:Offer",
							Target:   "ds-01",
							Assigner: "did:web:example.com",
							Assignee: "did:web:consumer.com",
							Permission: []Permission{
								{
									Action: "odrl:use",
									Constraint: []Constraint{
										{LeftOperand: "spatial", Operator: "eq", RightOperand: "EU"},
									},
								},
							},
							Prohibition: []Prohibition{
								{
									Action: "odrl:distribute",
								},
							},
							Obligation: []Duty{
								{
									Action: "odrl:compensate",
								},
							},
						},
					},
				},
			},
		},
		Services: []DataService{
			{
				ID:             "ds-ref-01",
				Type:           "dcat:DataService",
				Title:          "Test Service",
				EndpointURL:    "http://127.0.0.1:8080/endpoint",
				EndpointType:   "http-api",
				ServedDatasets: []string{"ds-01"},
			},
		},
	}

	if cat.ID != "catalog-01" {
		t.Errorf("expected catalog ID 'catalog-01', got %s", cat.ID)
	}
	if len(cat.Datasets) != 1 {
		t.Errorf("expected 1 dataset, got %d", len(cat.Datasets))
	}
	if cat.Datasets[0].Distributions[0].Policy.Permission[0].Constraint[0].RightOperand != "EU" {
		t.Errorf("expected constraint right operand 'EU'")
	}
}
