package ports_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/afinana/go-dataspace-components/catalog/domain"
	catalogports "github.com/afinana/go-dataspace-components/catalog/ports"
	"github.com/afinana/go-dataspace-components/internal/pkg/kvstore"
)

func TestOxigraphCatalogStore_RegisterAndGet(t *testing.T) {
	// Mock Oxigraph HTTP server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/update":
			w.WriteHeader(http.StatusNoContent)
		case "/query":
			w.Header().Set("Content-Type", "application/sparql-results+json")
			w.Write([]byte(`{
				"results": {
					"bindings": [
						{
							"payload": {
								"type": "literal",
								"value": "{\"id\":\"dataset-1\",\"type\":\"dcat:Dataset\",\"title\":\"Test RDF Dataset\",\"distributions\":[]}"
							}
						}
					]
				}
			}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer mockServer.Close()

	kvCache := kvstore.NewMemoryKVStore(1 * time.Minute)
	defer kvCache.Close()

	store := catalogports.NewOxigraphCatalogStore(mockServer.URL, "did:web:provider").WithCache(kvCache, 1*time.Minute)

	ctx := context.Background()
	dataset := &domain.Dataset{
		ID:    "dataset-1",
		Type:  "dcat:Dataset",
		Title: "Test RDF Dataset",
	}

	err := store.RegisterDataset(ctx, dataset)
	if err != nil {
		t.Fatalf("RegisterDataset failed: %v", err)
	}

	retrieved, err := store.GetDataset(ctx, "dataset-1")
	if err != nil {
		t.Fatalf("GetDataset failed: %v", err)
	}

	if retrieved.ID != "dataset-1" || retrieved.Title != "Test RDF Dataset" {
		t.Errorf("unexpected dataset retrieved: %+v", retrieved)
	}
}

func TestOxigraphCatalogStore_DeleteAndList(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/update":
			w.WriteHeader(http.StatusNoContent)
		case "/query":
			w.Header().Set("Content-Type", "application/sparql-results+json")
			w.Write([]byte(`{
				"results": {
					"bindings": [
						{
							"payload": {
								"type": "literal",
								"value": "{\"id\":\"ds-1\",\"title\":\"DS 1\"}"
							}
						}
					]
				}
			}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer mockServer.Close()

	store := catalogports.NewOxigraphCatalogStore(mockServer.URL, "did:web:provider")
	ctx := context.Background()

	err := store.DeleteDataset(ctx, "ds-1")
	if err != nil {
		t.Fatalf("DeleteDataset failed: %v", err)
	}

	datasets, err := store.ListDatasets(ctx)
	if err != nil {
		t.Fatalf("ListDatasets failed: %v", err)
	}

	if len(datasets) != 1 || datasets[0].ID != "ds-1" {
		t.Errorf("unexpected datasets list: %+v", datasets)
	}

	catalog, err := store.GetCatalog(ctx, "req-1")
	if err != nil || catalog == nil {
		t.Fatalf("GetCatalog failed: %v", err)
	}
}
