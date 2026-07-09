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

func TestPostgresNegotiationStore(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open stub db connection: %v", err)
	}
	defer db.Close()

	store := NewPostgresNegotiationStore(db)

	cn := &domain.ContractNegotiation{
		ID:            "negotiation-test-1",
		CorrelationID: "correlation-123",
		CounterParty:  "did:web:counterparty",
		Type:          domain.TypeConsumer,
		State:         domain.StateRequested,
		ContractOffer: &domain.ContractOffer{
			ID:      "offer-1",
			AssetID: "asset-1",
		},
		Agreement: &domain.ContractAgreement{
			ID:         "agreement-1",
			ProviderID: "provider-1",
			ConsumerID: "consumer-1",
			AssetID:    "asset-1",
		},
		ErrorDetail: "no error",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	offerBytes, _ := json.Marshal(cn.ContractOffer)
	agreementBytes, _ := json.Marshal(cn.Agreement)

	// 1. Test Save
	mock.ExpectExec("INSERT INTO contract_negotiations").
		WithArgs(
			cn.ID,
			cn.CorrelationID,
			cn.CounterParty,
			string(cn.Type),
			int(cn.State),
			offerBytes,
			agreementBytes,
			cn.ErrorDetail,
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = store.Save(context.Background(), cn)
	if err != nil {
		t.Errorf("failed to save contract negotiation: %v", err)
	}

	// 2. Test FindByID
	rows := sqlmock.NewRows([]string{
		"id", "correlation_id", "counter_party", "type", "state", "contract_offer", "agreement", "error_detail", "created_at", "updated_at",
	}).AddRow(
		cn.ID, cn.CorrelationID, cn.CounterParty, string(cn.Type), int(cn.State), offerBytes, agreementBytes, cn.ErrorDetail, cn.CreatedAt, cn.UpdatedAt,
	)

	mock.ExpectQuery("SELECT (.+) FROM contract_negotiations WHERE id = \\$1").
		WithArgs(cn.ID).
		WillReturnRows(rows)

	fetched, err := store.FindByID(context.Background(), cn.ID)
	if err != nil {
		t.Errorf("failed to find contract negotiation by ID: %v", err)
	}
	if fetched.ID != cn.ID {
		t.Errorf("expected ID %s, got %s", cn.ID, fetched.ID)
	}
	if fetched.CorrelationID != cn.CorrelationID {
		t.Errorf("expected correlation ID %s, got %s", cn.CorrelationID, fetched.CorrelationID)
	}

	// 3. Test FindByCorrelationID
	rows = sqlmock.NewRows([]string{
		"id", "correlation_id", "counter_party", "type", "state", "contract_offer", "agreement", "error_detail", "created_at", "updated_at",
	}).AddRow(
		cn.ID, cn.CorrelationID, cn.CounterParty, string(cn.Type), int(cn.State), offerBytes, agreementBytes, cn.ErrorDetail, cn.CreatedAt, cn.UpdatedAt,
	)

	mock.ExpectQuery("SELECT (.+) FROM contract_negotiations WHERE correlation_id = \\$1").
		WithArgs(cn.CorrelationID).
		WillReturnRows(rows)

	fetchedCorr, err := store.FindByCorrelationID(context.Background(), cn.CorrelationID)
	if err != nil {
		t.Errorf("failed to find contract negotiation by correlation ID: %v", err)
	}
	if fetchedCorr.ID != cn.ID {
		t.Errorf("expected ID %s, got %s", cn.ID, fetchedCorr.ID)
	}

	// 4. Test Update
	mock.ExpectExec("UPDATE contract_negotiations").
		WithArgs(
			cn.ID,
			cn.CorrelationID,
			cn.CounterParty,
			string(cn.Type),
			int(cn.State),
			offerBytes,
			agreementBytes,
			cn.ErrorDetail,
			sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = store.Update(context.Background(), cn)
	if err != nil {
		t.Errorf("failed to update contract negotiation: %v", err)
	}

	// 5. Test FindByID Not Found
	mock.ExpectQuery("SELECT (.+) FROM contract_negotiations WHERE id = \\$1").
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
