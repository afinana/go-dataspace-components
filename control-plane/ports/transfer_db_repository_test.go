package ports

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/afinana/go-dataspace-components/control-plane/domain"
)

func TestPostgresTransferStore(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open stub db connection: %v", err)
	}
	defer db.Close()

	store := NewPostgresTransferStore(db)

	tp := &domain.TransferProcess{
		ID:                  "transfer-test-1",
		ContractAgreementID: "agreement-123",
		CorrelationID:       "transfer-corr-123",
		AssetID:             "asset-123",
		State:               domain.StateTransferStarted,
		DataDestination: domain.DataAddress{
			Type: "HttpProxy",
			Properties: map[string]string{
				"endpoint": "http://localhost:9000",
			},
		},
		DataSource: domain.DataAddress{
			Type: "HttpData",
			Properties: map[string]string{
				"endpoint": "http://localhost:8080/mock-backend",
			},
		},
		ErrorDetail: "none",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	destBytes, _ := json.Marshal(tp.DataDestination)
	sourceBytes, _ := json.Marshal(tp.DataSource)

	// 1. Test Save
	mock.ExpectExec("INSERT INTO transfer_processes").
		WithArgs(
			tp.ID,
			tp.ContractAgreementID,
			tp.CorrelationID,
			tp.AssetID,
			int(tp.State),
			destBytes,
			sourceBytes,
			tp.ErrorDetail,
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = store.Save(context.Background(), tp)
	if err != nil {
		t.Errorf("failed to save transfer process: %v", err)
	}

	// 2. Test FindByID
	rows := sqlmock.NewRows([]string{
		"id", "contract_agreement_id", "correlation_id", "asset_id", "state", "data_destination", "data_source", "error_detail", "created_at", "updated_at",
	}).AddRow(
		tp.ID, tp.ContractAgreementID, tp.CorrelationID, tp.AssetID, int(tp.State), destBytes, sourceBytes, tp.ErrorDetail, tp.CreatedAt, tp.UpdatedAt,
	)

	mock.ExpectQuery("SELECT (.+) FROM transfer_processes WHERE id = \\$1").
		WithArgs(tp.ID).
		WillReturnRows(rows)

	fetched, err := store.FindByID(context.Background(), tp.ID)
	if err != nil {
		t.Errorf("failed to find transfer process by ID: %v", err)
	}
	if fetched.ID != tp.ID {
		t.Errorf("expected ID %s, got %s", tp.ID, fetched.ID)
	}
	if fetched.AssetID != tp.AssetID {
		t.Errorf("expected asset ID %s, got %s", tp.AssetID, fetched.AssetID)
	}

	// 3. Test FindByCorrelationID
	rows = sqlmock.NewRows([]string{
		"id", "contract_agreement_id", "correlation_id", "asset_id", "state", "data_destination", "data_source", "error_detail", "created_at", "updated_at",
	}).AddRow(
		tp.ID, tp.ContractAgreementID, tp.CorrelationID, tp.AssetID, int(tp.State), destBytes, sourceBytes, tp.ErrorDetail, tp.CreatedAt, tp.UpdatedAt,
	)

	mock.ExpectQuery("SELECT (.+) FROM transfer_processes WHERE correlation_id = \\$1").
		WithArgs(tp.CorrelationID).
		WillReturnRows(rows)

	fetchedCorr, err := store.FindByCorrelationID(context.Background(), tp.CorrelationID)
	if err != nil {
		t.Errorf("failed to find transfer process by correlation ID: %v", err)
	}
	if fetchedCorr.ID != tp.ID {
		t.Errorf("expected ID %s, got %s", tp.ID, fetchedCorr.ID)
	}

	// 4. Test Update
	mock.ExpectExec("UPDATE transfer_processes").
		WithArgs(
			tp.ID,
			tp.ContractAgreementID,
			tp.CorrelationID,
			tp.AssetID,
			int(tp.State),
			destBytes,
			sourceBytes,
			tp.ErrorDetail,
			sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = store.Update(context.Background(), tp)
	if err != nil {
		t.Errorf("failed to update transfer process: %v", err)
	}

	// 5. Test FindByID Not Found
	mock.ExpectQuery("SELECT (.+) FROM transfer_processes WHERE id = \\$1").
		WithArgs("invalid-id").
		WillReturnError(sql.ErrNoRows)

	_, err = store.FindByID(context.Background(), "invalid-id")
	if !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("expected sql.ErrNoRows error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
