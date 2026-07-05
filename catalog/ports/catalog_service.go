package ports

import (
	"context"
	"errors"
	"sync"

	"github.com/afinana/go-dataspace-components/catalog/domain"
)

// InMemoryCatalogService implements domain.AssetRegistry and domain.CatalogQueryService.
type InMemoryCatalogService struct {
	mu        sync.RWMutex
	datasets  map[string]domain.Dataset
	services  map[string]domain.DataService
	publisher string
}

// NewInMemoryCatalogService creates a new instance of InMemoryCatalogService.
func NewInMemoryCatalogService(publisher string) *InMemoryCatalogService {
	return &InMemoryCatalogService{
		datasets:  make(map[string]domain.Dataset),
		services:  make(map[string]domain.DataService),
		publisher: publisher,
	}
}

// RegisterDataset stores a W3C DCAT-AP dataset in memory.
func (s *InMemoryCatalogService) RegisterDataset(ctx context.Context, dataset *domain.Dataset) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if dataset.ID == "" {
		return errors.New("dataset ID cannot be empty")
	}
	s.datasets[dataset.ID] = *dataset
	return nil
}

// GetDataset retrieves a W3C DCAT-AP dataset by ID.
func (s *InMemoryCatalogService) GetDataset(ctx context.Context, id string) (*domain.Dataset, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	dataset, exists := s.datasets[id]
	if !exists {
		return nil, errors.New("dataset not found")
	}
	return &dataset, nil
}

// DeleteDataset removes a W3C DCAT-AP dataset from memory.
func (s *InMemoryCatalogService) DeleteDataset(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.datasets[id]; !exists {
		return errors.New("dataset not found")
	}
	delete(s.datasets, id)
	return nil
}

// ListDatasets returns all registered datasets.
func (s *InMemoryCatalogService) ListDatasets(ctx context.Context) ([]domain.Dataset, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]domain.Dataset, 0, len(s.datasets))
	for _, d := range s.datasets {
		list = append(list, d)
	}
	return list, nil
}

// GetCatalog constructs and returns the W3C DCAT-AP Catalog.
func (s *InMemoryCatalogService) GetCatalog(ctx context.Context, requesterID string) (*domain.Catalog, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	datasets := make([]domain.Dataset, 0, len(s.datasets))
	for _, d := range s.datasets {
		datasets = append(datasets, d)
	}

	services := make([]domain.DataService, 0, len(s.services))
	for _, srv := range s.services {
		services = append(services, srv)
	}

	return &domain.Catalog{
		ID:          "catalog-main",
		Type:        "dcat:Catalog",
		Title:       "Main Dataspace Catalog",
		Description: "A sovereign DCAT-AP compliant catalog for dataspace operations",
		Publisher:   s.publisher,
		Datasets:    datasets,
		Services:    services,
	}, nil
}

// QueryDatasets queries datasets matching specific metadata filters.
func (s *InMemoryCatalogService) QueryDatasets(ctx context.Context, filter map[string]string) ([]domain.Dataset, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var matched []domain.Dataset
	for _, d := range s.datasets {
		match := true
		for k, v := range filter {
			if k == "keyword" {
				keywordMatch := false
				for _, kw := range d.Keywords {
					if kw == v {
						keywordMatch = true
						break
					}
				}
				if !keywordMatch {
					match = false
					break
				}
			}
		}
		if match {
			matched = append(matched, d)
		}
	}
	return matched, nil
}

// RegisterService registers a data service in the catalog (helper for testing/population).
func (s *InMemoryCatalogService) RegisterService(ctx context.Context, srv *domain.DataService) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if srv.ID == "" {
		return errors.New("service ID cannot be empty")
	}
	s.services[srv.ID] = *srv
	return nil
}
