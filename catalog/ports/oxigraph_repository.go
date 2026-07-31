package ports

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/afinana/go-dataspace-components/catalog/domain"
	"github.com/afinana/go-dataspace-components/internal/pkg/kvstore"
)

// OxigraphCatalogStore implements domain.AssetRegistry and domain.CatalogQueryService using an Oxigraph Triple Store SPARQL endpoint.
type OxigraphCatalogStore struct {
	oxigraphURL string
	publisher   string
	httpClient  *http.Client
	cache       kvstore.KVStore
	cacheTTL    time.Duration
}

// NewOxigraphCatalogStore creates a new Oxigraph triple store repository instance.
func NewOxigraphCatalogStore(oxigraphURL string, publisher string) *OxigraphCatalogStore {
	if oxigraphURL == "" {
		oxigraphURL = "http://localhost:7878"
	}
	return &OxigraphCatalogStore{
		oxigraphURL: oxigraphURL,
		publisher:   publisher,
		httpClient:  &http.Client{Timeout: 10 * time.Second},
		cacheTTL:    5 * time.Minute,
	}
}

// WithCache attaches an L1 KV Store cache to the Oxigraph store.
func (s *OxigraphCatalogStore) WithCache(cache kvstore.KVStore, ttl time.Duration) *OxigraphCatalogStore {
	s.cache = cache
	if ttl > 0 {
		s.cacheTTL = ttl
	}
	return s
}

// RegisterDataset stores dataset metadata as RDF triples in Oxigraph via SPARQL UPDATE.
func (s *OxigraphCatalogStore) RegisterDataset(ctx context.Context, dataset *domain.Dataset) error {
	if dataset.ID == "" {
		return fmt.Errorf("dataset ID cannot be empty")
	}

	payloadBytes, err := json.Marshal(dataset)
	if err != nil {
		return fmt.Errorf("failed to marshal dataset for oxigraph write: %w", err)
	}

	// Construct SPARQL INSERT DATA query replacing existing dataset subgraph
	sparqlUpdate := fmt.Sprintf(`
		PREFIX dcat: <http://www.w3.org/ns/dcat#>
		PREFIX dct: <http://purl.org/dc/terms/>
		PREFIX xsd: <http://www.w3.org/2001/XMLSchema#>

		DELETE { <urn:dataset:%s> ?p ?o }
		WHERE { <urn:dataset:%s> ?p ?o };

		INSERT DATA {
			<urn:dataset:%s> a dcat:Dataset ;
				dct:title "%s" ;
				dct:description "%s" ;
				dct:publisher "%s" ;
				dct:jsonPayload "%s" .
		}
	`, dataset.ID, dataset.ID, dataset.ID, dataset.Title, dataset.Description, dataset.Publisher, url.QueryEscape(string(payloadBytes)))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.oxigraphURL+"/update", bytes.NewBufferString(sparqlUpdate))
	if err != nil {
		return fmt.Errorf("failed to build SPARQL update request: %w", err)
	}
	req.Header.Set("Content-Type", "application/sparql-update")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute SPARQL update against Oxigraph: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("oxigraph SPARQL update failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Update KV cache if configured
	if s.cache != nil {
		_ = s.cache.Set(ctx, "dataset:"+dataset.ID, payloadBytes, s.cacheTTL)
		_ = s.cache.Delete(ctx, "datasets:all")
		_ = s.cache.Delete(ctx, "catalog:main")
	}

	return nil
}

// GetDataset queries dataset triples from Oxigraph via SPARQL SELECT.
func (s *OxigraphCatalogStore) GetDataset(ctx context.Context, id string) (*domain.Dataset, error) {
	if s.cache != nil {
		if cachedBytes, found, _ := s.cache.Get(ctx, "dataset:"+id); found {
			var dataset domain.Dataset
			if err := json.Unmarshal(cachedBytes, &dataset); err == nil {
				return &dataset, nil
			}
		}
	}

	sparqlQuery := fmt.Sprintf(`
		PREFIX dct: <http://purl.org/dc/terms/>
		SELECT ?payload WHERE {
			<urn:dataset:%s> dct:jsonPayload ?payload .
		} LIMIT 1
	`, id)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.oxigraphURL+"/query?query="+url.QueryEscape(sparqlQuery), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build SPARQL query request: %w", err)
	}
	req.Header.Set("Accept", "application/sparql-results+json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute SPARQL query: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("oxigraph query returned status: %d", resp.StatusCode)
	}

	var result struct {
		Results struct {
			Bindings []map[string]struct {
				Value string `json:"value"`
			} `json:"bindings"`
		} `json:"results"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode SPARQL JSON response: %w", err)
	}

	if len(result.Results.Bindings) == 0 {
		return nil, fmt.Errorf("dataset not found in triple store")
	}

	escapedPayload := result.Results.Bindings[0]["payload"].Value
	rawJSON, err := url.QueryUnescape(escapedPayload)
	if err != nil {
		rawJSON = escapedPayload
	}

	var dataset domain.Dataset
	if err := json.Unmarshal([]byte(rawJSON), &dataset); err != nil {
		return nil, fmt.Errorf("failed to unmarshal dataset json payload: %w", err)
	}

	return &dataset, nil
}

// DeleteDataset removes a dataset triple graph from Oxigraph.
func (s *OxigraphCatalogStore) DeleteDataset(ctx context.Context, id string) error {
	sparqlUpdate := fmt.Sprintf(`
		DELETE WHERE { <urn:dataset:%s> ?p ?o }
	`, id)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.oxigraphURL+"/update", bytes.NewBufferString(sparqlUpdate))
	if err != nil {
		return fmt.Errorf("failed to create delete request: %w", err)
	}
	req.Header.Set("Content-Type", "application/sparql-update")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute delete update: %w", err)
	}
	defer resp.Body.Close()

	if s.cache != nil {
		_ = s.cache.Delete(ctx, "dataset:"+id)
		_ = s.cache.Delete(ctx, "datasets:all")
		_ = s.cache.Delete(ctx, "catalog:main")
	}

	return nil
}

// ListDatasets queries all datasets stored in Oxigraph.
func (s *OxigraphCatalogStore) ListDatasets(ctx context.Context) ([]domain.Dataset, error) {
	sparqlQuery := `
		PREFIX dct: <http://purl.org/dc/terms/>
		SELECT ?payload WHERE {
			?s dct:jsonPayload ?payload .
		}
	`

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.oxigraphURL+"/query?query="+url.QueryEscape(sparqlQuery), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build SPARQL query request: %w", err)
	}
	req.Header.Set("Accept", "application/sparql-results+json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute SPARQL list query: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Results struct {
			Bindings []map[string]struct {
				Value string `json:"value"`
			} `json:"bindings"`
		} `json:"results"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode SPARQL JSON response: %w", err)
	}

	var datasets []domain.Dataset
	for _, binding := range result.Results.Bindings {
		escapedPayload := binding["payload"].Value
		rawJSON, err := url.QueryUnescape(escapedPayload)
		if err != nil {
			rawJSON = escapedPayload
		}
		var dataset domain.Dataset
		if err := json.Unmarshal([]byte(rawJSON), &dataset); err == nil {
			datasets = append(datasets, dataset)
		}
	}

	return datasets, nil
}

// GetCatalog aggregates catalog datasets from Oxigraph.
func (s *OxigraphCatalogStore) GetCatalog(ctx context.Context, requesterID string) (*domain.Catalog, error) {
	datasets, err := s.ListDatasets(ctx)
	if err != nil {
		return nil, err
	}

	return &domain.Catalog{
		ID:          "catalog:main",
		Type:        "dcat:Catalog",
		Title:       "Sovereign Oxigraph TripleStore Catalog",
		Description: "RDF Triplestore DCAT-AP Catalog Powered by Oxigraph",
		Publisher:   s.publisher,
		Datasets:    datasets,
	}, nil
}

// QueryDatasets executes filtering over dataset items.
func (s *OxigraphCatalogStore) QueryDatasets(ctx context.Context, filter map[string]string) ([]domain.Dataset, error) {
	return s.ListDatasets(ctx)
}
