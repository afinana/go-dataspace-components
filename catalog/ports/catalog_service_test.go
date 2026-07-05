package ports

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/afinana/go-dataspace-components/catalog/domain"
	"go.opentelemetry.io/otel/trace/noop"
)

func TestInMemoryCatalogService(t *testing.T) {
	ctx := context.Background()
	service := NewInMemoryCatalogService("did:web:local-connector")

	dataset := &domain.Dataset{
		ID:          "ds-1",
		Type:        "dcat:Dataset",
		Title:       "Test Dataset",
		Description: "A dataset for testing",
	}

	// Test Register
	err := service.RegisterDataset(ctx, dataset)
	if err != nil {
		t.Fatalf("failed to register dataset: %v", err)
	}

	// Test Get
	fetched, err := service.GetDataset(ctx, "ds-1")
	if err != nil {
		t.Fatalf("failed to get dataset: %v", err)
	}
	if fetched.Title != "Test Dataset" {
		t.Errorf("expected title 'Test Dataset', got '%s'", fetched.Title)
	}

	// Test List
	list, err := service.ListDatasets(ctx)
	if err != nil {
		t.Fatalf("failed to list datasets: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("expected list size 1, got %d", len(list))
	}

	// Test GetCatalog
	catalog, err := service.GetCatalog(ctx, "")
	if err != nil {
		t.Fatalf("failed to get catalog: %v", err)
	}
	if len(catalog.Datasets) != 1 {
		t.Errorf("expected 1 dataset in catalog, got %d", len(catalog.Datasets))
	}

	// Test Delete
	err = service.DeleteDataset(ctx, "ds-1")
	if err != nil {
		t.Fatalf("failed to delete dataset: %v", err)
	}

	_, err = service.GetDataset(ctx, "ds-1")
	if err == nil {
		t.Error("expected error getting deleted dataset, got nil")
	}
}

func TestPostgresCatalogStore(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open stub db connection: %v", err)
	}
	defer db.Close()

	store := NewPostgresCatalogStore(db, "did:web:local-connector")

	dataset := &domain.Dataset{
		ID:          "ds-postgres",
		Type:        "dcat:Dataset",
		Title:       "Postgres Dataset",
		Description: "Dataset stored in Postgres",
	}

	payloadBytes, _ := json.Marshal(dataset)

	// 1. Test RegisterDataset
	mock.ExpectExec("INSERT INTO datasets").
		WithArgs(dataset.ID, payloadBytes).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = store.RegisterDataset(context.Background(), dataset)
	if err != nil {
		t.Errorf("failed to register dataset: %v", err)
	}

	// 2. Test GetDataset
	rows := sqlmock.NewRows([]string{"payload"}).AddRow(payloadBytes)
	mock.ExpectQuery("SELECT payload FROM datasets WHERE id = ?").
		WithArgs(dataset.ID).
		WillReturnRows(rows)

	fetched, err := store.GetDataset(context.Background(), dataset.ID)
	if err != nil {
		t.Errorf("failed to get dataset: %v", err)
	}
	if fetched.Title != "Postgres Dataset" {
		t.Errorf("expected title 'Postgres Dataset', got '%s'", fetched.Title)
	}

	// 3. Test DeleteDataset
	mock.ExpectExec("DELETE FROM datasets WHERE id = ?").
		WithArgs(dataset.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = store.DeleteDataset(context.Background(), dataset.ID)
	if err != nil {
		t.Errorf("failed to delete dataset: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestCatalogAPIHandler(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	tracer := noop.NewTracerProvider().Tracer("noop")
	service := NewInMemoryCatalogService("did:web:local-connector")
	handler := NewCatalogAPIHandler(logger, tracer, service, service)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	// 1. Test POST /catalog/datasets
	dataset := domain.Dataset{
		ID:          "ds-test",
		Type:        "dcat:Dataset",
		Title:       "HTTP Test Dataset",
		Description: "Testing HTTP handlers",
	}
	body, _ := json.Marshal(dataset)
	req := httptest.NewRequest("POST", "/catalog/datasets", bytes.NewReader(body))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", w.Code)
	}

	// 2. Test GET /catalog/datasets/ds-test
	req = httptest.NewRequest("GET", "/catalog/datasets/ds-test", nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	var fetched domain.Dataset
	json.NewDecoder(w.Body).Decode(&fetched)
	if fetched.ID != "ds-test" {
		t.Errorf("expected dataset ID 'ds-test', got '%s'", fetched.ID)
	}

	// 3. Test GET /catalog
	req = httptest.NewRequest("GET", "/catalog", nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	var catalog domain.Catalog
	json.NewDecoder(w.Body).Decode(&catalog)
	if len(catalog.Datasets) != 1 {
		t.Errorf("expected 1 dataset in catalog, got %d", len(catalog.Datasets))
	}

	// 4. Test DELETE /catalog/datasets/ds-test
	req = httptest.NewRequest("DELETE", "/catalog/datasets/ds-test", nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	// 5. Verify deleted
	req = httptest.NewRequest("GET", "/catalog/datasets/ds-test", nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404 after deletion, got %d", w.Code)
	}
}
