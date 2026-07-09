package ports

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	cp "github.com/afinana/go-dataspace-components/control-plane/domain"
	dp "github.com/afinana/go-dataspace-components/data-plane/domain"
)

func TestPostgresDataFlowStore(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open stub db connection: %v", err)
	}
	defer db.Close()

	store := NewPostgresDataFlowStore(db)

	req := &dp.DataFlowRequest{
		ID:                  "flow-1",
		ContractAgreementID: "agreement-123",
		SourceDataAddress: cp.DataAddress{
			Type: "HttpData",
			Properties: map[string]string{
				"endpoint": "http://localhost:8081/mock-backend",
			},
		},
		DestinationDataAddress: cp.DataAddress{
			Type: "HttpProxy",
		},
		Properties: map[string]string{
			"auth_token": "token-123",
		},
	}
	token := "token-123"

	sourceBytes, _ := json.Marshal(req.SourceDataAddress)
	destBytes, _ := json.Marshal(req.DestinationDataAddress)
	propBytes, _ := json.Marshal(req.Properties)

	// 1. Test Save
	mock.ExpectExec("INSERT INTO data_flows").
		WithArgs(
			req.ID,
			token,
			req.ContractAgreementID,
			sourceBytes,
			destBytes,
			propBytes,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = store.Save(context.Background(), token, req)
	if err != nil {
		t.Errorf("failed to save data flow: %v", err)
	}

	// 2. Test FindByToken
	rows := sqlmock.NewRows([]string{
		"id", "contract_agreement_id", "source_data_address", "destination_data_address", "properties",
	}).AddRow(
		req.ID, req.ContractAgreementID, sourceBytes, destBytes, propBytes,
	)

	mock.ExpectQuery("SELECT id, contract_agreement_id, source_data_address, destination_data_address, properties FROM data_flows WHERE token = \\$1").
		WithArgs(token).
		WillReturnRows(rows)

	fetched, err := store.FindByToken(context.Background(), token)
	if err != nil {
		t.Errorf("failed to find data flow by token: %v", err)
	}
	if fetched.ID != req.ID {
		t.Errorf("expected flow ID %s, got %s", req.ID, fetched.ID)
	}

	// 3. Test FindByFlowID
	rows = sqlmock.NewRows([]string{
		"token", "contract_agreement_id", "source_data_address", "destination_data_address", "properties",
	}).AddRow(
		token, req.ContractAgreementID, sourceBytes, destBytes, propBytes,
	)

	mock.ExpectQuery("SELECT token, contract_agreement_id, source_data_address, destination_data_address, properties FROM data_flows WHERE id = \\$1").
		WithArgs(req.ID).
		WillReturnRows(rows)

	fetchedToken, fetchedReq, err := store.FindByFlowID(context.Background(), req.ID)
	if err != nil {
		t.Errorf("failed to find data flow by flow ID: %v", err)
	}
	if fetchedToken != token {
		t.Errorf("expected token %s, got %s", token, fetchedToken)
	}
	if fetchedReq.ID != req.ID {
		t.Errorf("expected flow ID %s, got %s", req.ID, fetchedReq.ID)
	}

	// 4. Test ListAll
	rows = sqlmock.NewRows([]string{
		"id", "token", "contract_agreement_id", "source_data_address", "destination_data_address", "properties",
	}).AddRow(
		req.ID, token, req.ContractAgreementID, sourceBytes, destBytes, propBytes,
	)

	mock.ExpectQuery("SELECT id, token, contract_agreement_id, source_data_address, destination_data_address, properties FROM data_flows").
		WillReturnRows(rows)

	allFlows, err := store.ListAll(context.Background())
	if err != nil {
		t.Errorf("failed to list all data flows: %v", err)
	}
	if len(allFlows) != 1 {
		t.Errorf("expected 1 active flow, got %d", len(allFlows))
	}
	if allFlows[token].ID != req.ID {
		t.Errorf("expected flow ID %s under token %s", req.ID, token)
	}

	// 5. Test Delete
	mock.ExpectExec("DELETE FROM data_flows WHERE id = \\$1").
		WithArgs(req.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = store.Delete(context.Background(), req.ID)
	if err != nil {
		t.Errorf("failed to delete data flow: %v", err)
	}

	// 6. Test FindByToken Not Found
	mock.ExpectQuery("SELECT (.+) FROM data_flows WHERE token = \\$1").
		WithArgs("invalid-token").
		WillReturnError(sql.ErrNoRows)

	_, err = store.FindByToken(context.Background(), "invalid-token")
	if err == nil {
		t.Error("expected error for non-existent token, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
