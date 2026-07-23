package ports

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/afinana/go-dataspace-components/catalog/domain"
	"github.com/afinana/go-dataspace-components/internal/pkg/kvstore"
)

// PostgresCatalogStore implements domain.AssetRegistry and domain.CatalogQueryService.
type PostgresCatalogStore struct {
	db        *sql.DB
	publisher string
	cache     kvstore.KVStore
	cacheTTL  time.Duration
}

// NewPostgresCatalogStore creates a new storage repository instance.
func NewPostgresCatalogStore(db *sql.DB, publisher string) *PostgresCatalogStore {
	return &PostgresCatalogStore{
		db:        db,
		publisher: publisher,
		cacheTTL:  5 * time.Minute,
	}
}

// WithCache attaches an L1 KV Store cache to the Postgres catalog store.
func (s *PostgresCatalogStore) WithCache(cache kvstore.KVStore, ttl time.Duration) *PostgresCatalogStore {
	s.cache = cache
	if ttl > 0 {
		s.cacheTTL = ttl
	}
	return s
}

// RegisterDataset persists a W3C DCAT-AP dataset in PostgreSQL and updates the KV cache.
func (s *PostgresCatalogStore) RegisterDataset(ctx context.Context, dataset *domain.Dataset) error {
	if dataset.ID == "" {
		return errors.New("dataset ID cannot be empty")
	}

	payloadBytes, err := json.Marshal(dataset)
	if err != nil {
		return fmt.Errorf("failed to marshal dataset payload for database write: %w", err)
	}

	query := `
		INSERT INTO datasets (id, payload)
		VALUES ($1, $2)
		ON CONFLICT (id) DO UPDATE SET
			payload = EXCLUDED.payload,
			updated_at = CURRENT_TIMESTAMP;
	`
	_, err = s.db.ExecContext(ctx, query, dataset.ID, payloadBytes)
	if err != nil {
		return fmt.Errorf("failed to execute postgres insert for dataset: %w", err)
	}

	// Update KV cache if present
	if s.cache != nil {
		_ = s.cache.Set(ctx, "dataset:"+dataset.ID, payloadBytes, s.cacheTTL)
		_ = s.cache.Delete(ctx, "datasets:all")
		_ = s.cache.Delete(ctx, "catalog:main")
	}

	return nil
}

// GetDataset retrieves a W3C DCAT-AP dataset by ID, checking the KV cache first.
func (s *PostgresCatalogStore) GetDataset(ctx context.Context, id string) (*domain.Dataset, error) {
	// L1 Cache read
	if s.cache != nil {
		if cachedBytes, found, _ := s.cache.Get(ctx, "dataset:"+id); found {
			var dataset domain.Dataset
			if err := json.Unmarshal(cachedBytes, &dataset); err == nil {
				return &dataset, nil
			}
		}
	}

	query := `SELECT payload FROM datasets WHERE id = $1`
	var payload []byte
	err := s.db.QueryRowContext(ctx, query, id).Scan(&payload)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("dataset not found")
		}
		return nil, fmt.Errorf("failed to query dataset: %w", err)
	}

	var dataset domain.Dataset
	if err := json.Unmarshal(payload, &dataset); err != nil {
		return nil, fmt.Errorf("failed to unmarshal dataset payload: %w", err)
	}

	// Populate KV cache
	if s.cache != nil {
		_ = s.cache.Set(ctx, "dataset:"+id, payload, s.cacheTTL)
	}

	return &dataset, nil
}

// DeleteDataset removes a W3C DCAT-AP dataset by ID and invalidates cache.
func (s *PostgresCatalogStore) DeleteDataset(ctx context.Context, id string) error {
	query := `DELETE FROM datasets WHERE id = $1`
	res, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to execute postgres delete for dataset: %w", err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("dataset not found")
	}

	if s.cache != nil {
		_ = s.cache.Delete(ctx, "dataset:"+id)
		_ = s.cache.Delete(ctx, "datasets:all")
		_ = s.cache.Delete(ctx, "catalog:main")
	}

	return nil
}

// ListDatasets retrieves all registered W3C DCAT-AP datasets.
func (s *PostgresCatalogStore) ListDatasets(ctx context.Context) ([]domain.Dataset, error) {
	if s.cache != nil {
		if cachedBytes, found, _ := s.cache.Get(ctx, "datasets:all"); found {
			var datasets []domain.Dataset
			if err := json.Unmarshal(cachedBytes, &datasets); err == nil {
				return datasets, nil
			}
		}
	}

	query := `SELECT payload FROM datasets`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query datasets: %w", err)
	}
	defer rows.Close()

	var datasets []domain.Dataset
	for rows.Next() {
		var payload []byte
		if err := rows.Scan(&payload); err != nil {
			return nil, fmt.Errorf("failed to scan dataset row: %w", err)
		}
		var d domain.Dataset
		if err := json.Unmarshal(payload, &d); err != nil {
			return nil, fmt.Errorf("failed to unmarshal dataset: %w", err)
		}
		datasets = append(datasets, d)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if s.cache != nil {
		if payloadBytes, err := json.Marshal(datasets); err == nil {
			_ = s.cache.Set(ctx, "datasets:all", payloadBytes, s.cacheTTL)
		}
	}

	return datasets, nil
}

// GetCatalog constructs and returns the full DCAT Catalog from database or cache.
func (s *PostgresCatalogStore) GetCatalog(ctx context.Context, requesterID string) (*domain.Catalog, error) {
	if s.cache != nil {
		if cachedBytes, found, _ := s.cache.Get(ctx, "catalog:main"); found {
			var cat domain.Catalog
			if err := json.Unmarshal(cachedBytes, &cat); err == nil {
				return &cat, nil
			}
		}
	}

	datasets, err := s.ListDatasets(ctx)
	if err != nil {
		return nil, err
	}

	services, err := s.ListServices(ctx)
	if err != nil {
		return nil, err
	}

	cat := &domain.Catalog{
		ID:          "catalog-main",
		Type:        "dcat:Catalog",
		Title:       "Main Dataspace Catalog",
		Description: "A sovereign DCAT-AP compliant catalog for dataspace operations",
		Publisher:   s.publisher,
		Datasets:    datasets,
		Services:    services,
	}

	if s.cache != nil {
		if payloadBytes, err := json.Marshal(cat); err == nil {
			_ = s.cache.Set(ctx, "catalog:main", payloadBytes, s.cacheTTL)
		}
	}

	return cat, nil
}

// QueryDatasets queries datasets matching specific metadata filters.
func (s *PostgresCatalogStore) QueryDatasets(ctx context.Context, filter map[string]string) ([]domain.Dataset, error) {
	datasets, err := s.ListDatasets(ctx)
	if err != nil {
		return nil, err
	}

	var matched []domain.Dataset
	for _, d := range datasets {
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

// RegisterService persists a W3C DCAT-AP data service in PostgreSQL and updates cache.
func (s *PostgresCatalogStore) RegisterService(ctx context.Context, srv *domain.DataService) error {
	if srv.ID == "" {
		return errors.New("service ID cannot be empty")
	}

	payloadBytes, err := json.Marshal(srv)
	if err != nil {
		return fmt.Errorf("failed to marshal data service payload: %w", err)
	}

	query := `
		INSERT INTO data_services (id, payload)
		VALUES ($1, $2)
		ON CONFLICT (id) DO UPDATE SET
			payload = EXCLUDED.payload,
			updated_at = CURRENT_TIMESTAMP;
	`
	_, err = s.db.ExecContext(ctx, query, srv.ID, payloadBytes)
	if err != nil {
		return fmt.Errorf("failed to execute postgres insert for data service: %w", err)
	}

	if s.cache != nil {
		_ = s.cache.Delete(ctx, "catalog:main")
	}

	return nil
}

// ListServices retrieves all registered W3C DCAT-AP data services.
func (s *PostgresCatalogStore) ListServices(ctx context.Context) ([]domain.DataService, error) {
	query := `SELECT payload FROM data_services`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query data services: %w", err)
	}
	defer rows.Close()

	var services []domain.DataService
	for rows.Next() {
		var payload []byte
		if err := rows.Scan(&payload); err != nil {
			return nil, fmt.Errorf("failed to scan data service row: %w", err)
		}
		var srv domain.DataService
		if err := json.Unmarshal(payload, &srv); err != nil {
			return nil, fmt.Errorf("failed to unmarshal data service: %w", err)
		}
		services = append(services, srv)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return services, nil
}
