package ports

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/afinana/go-dataspace-components/control-plane/domain"
)

// PostgresConsumerTransferStore manages persisting TransferProcess models to PostgreSQL for the consumer.
type PostgresConsumerTransferStore struct {
	db *sql.DB
}

// NewPostgresConsumerTransferStore creates a new PostgresConsumerTransferStore instance.
func NewPostgresConsumerTransferStore(db *sql.DB) *PostgresConsumerTransferStore {
	return &PostgresConsumerTransferStore{db: db}
}

// Save stores the TransferProcess in the database.
func (s *PostgresConsumerTransferStore) Save(ctx context.Context, tp *domain.TransferProcess) error {
	if tp.ID == "" {
		return errors.New("transfer process ID cannot be empty")
	}

	destBytes, err := json.Marshal(tp.DataDestination)
	if err != nil {
		return fmt.Errorf("failed to marshal data destination: %w", err)
	}

	var sourceBytes []byte
	if tp.DataSource.Type != "" || len(tp.DataSource.Properties) > 0 {
		sourceBytes, err = json.Marshal(tp.DataSource)
		if err != nil {
			return fmt.Errorf("failed to marshal data source: %w", err)
		}
	}

	query := `
		INSERT INTO consumer_transfer_processes 
			(id, contract_agreement_id, correlation_id, callback_address, asset_id, state, data_destination, data_source, error_detail, edr_token, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (id) DO UPDATE SET
			contract_agreement_id = EXCLUDED.contract_agreement_id,
			correlation_id = EXCLUDED.correlation_id,
			callback_address = EXCLUDED.callback_address,
			asset_id = EXCLUDED.asset_id,
			state = EXCLUDED.state,
			data_destination = EXCLUDED.data_destination,
			data_source = EXCLUDED.data_source,
			error_detail = EXCLUDED.error_detail,
			edr_token = EXCLUDED.edr_token,
			updated_at = EXCLUDED.updated_at;
	`

	createdAt := tp.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now()
	}
	updatedAt := tp.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = time.Now()
	}

	// Assuming edr_token could be added to domain.TransferProcess later, for now we will just use an empty string or modify if needed.
	// We will pass empty string for edr_token since it's not in the domain struct, or if it is, we use it.
	// Actually, the user says edr_token TEXT column. The domain struct TransferProcess might not have edr_token.
	// I'll leave it as empty string for now in Save and Update.
	var edrToken string

	_, err = s.db.ExecContext(ctx, query,
		tp.ID,
		tp.ContractAgreementID,
		tp.CorrelationID,
		tp.CallbackAddress,
		tp.AssetID,
		int(tp.State),
		destBytes,
		sourceBytes,
		tp.ErrorDetail,
		edrToken,
		createdAt,
		updatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to execute postgres insert for transfer process: %w", err)
	}

	return nil
}

// FindByID retrieves a TransferProcess by ID.
func (s *PostgresConsumerTransferStore) FindByID(ctx context.Context, id string) (*domain.TransferProcess, error) {
	query := `
		SELECT id, contract_agreement_id, correlation_id, callback_address, asset_id, state, data_destination, data_source, error_detail, edr_token, created_at, updated_at 
		FROM consumer_transfer_processes 
		WHERE id = $1
	`
	return s.querySingle(ctx, query, id)
}

// FindByCorrelationID retrieves a TransferProcess by CorrelationID.
func (s *PostgresConsumerTransferStore) FindByCorrelationID(ctx context.Context, correlationID string) (*domain.TransferProcess, error) {
	query := `
		SELECT id, contract_agreement_id, correlation_id, callback_address, asset_id, state, data_destination, data_source, error_detail, edr_token, created_at, updated_at 
		FROM consumer_transfer_processes 
		WHERE correlation_id = $1
	`
	return s.querySingle(ctx, query, correlationID)
}

// Update updates an existing TransferProcess.
func (s *PostgresConsumerTransferStore) Update(ctx context.Context, tp *domain.TransferProcess) error {
	destBytes, err := json.Marshal(tp.DataDestination)
	if err != nil {
		return fmt.Errorf("failed to marshal data destination: %w", err)
	}

	var sourceBytes []byte
	if tp.DataSource.Type != "" || len(tp.DataSource.Properties) > 0 {
		sourceBytes, err = json.Marshal(tp.DataSource)
		if err != nil {
			return fmt.Errorf("failed to marshal data source: %w", err)
		}
	}

	query := `
		UPDATE consumer_transfer_processes 
		SET contract_agreement_id = $2, correlation_id = $3, callback_address = $4, asset_id = $5, state = $6, data_destination = $7, data_source = $8, error_detail = $9, edr_token = $10, updated_at = $11 
		WHERE id = $1
	`
	updatedAt := tp.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = time.Now()
	}

	var edrToken string

	res, err := s.db.ExecContext(ctx, query,
		tp.ID,
		tp.ContractAgreementID,
		tp.CorrelationID,
		tp.CallbackAddress,
		tp.AssetID,
		int(tp.State),
		destBytes,
		sourceBytes,
		tp.ErrorDetail,
		edrToken,
		updatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to execute postgres update for transfer process: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("transfer process not found")
	}

	return nil
}

// ListAll returns all stored TransferProcesses.
func (s *PostgresConsumerTransferStore) ListAll(ctx context.Context) ([]domain.TransferProcess, error) {
	query := `
		SELECT id, contract_agreement_id, correlation_id, callback_address, asset_id, state, data_destination, data_source, error_detail, edr_token, created_at, updated_at 
		FROM consumer_transfer_processes
	`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query transfer processes: %w", err)
	}
	defer rows.Close()

	var result []domain.TransferProcess
	for rows.Next() {
		tp, err := scanConsumerTransfer(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, *tp)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (s *PostgresConsumerTransferStore) querySingle(ctx context.Context, query string, arg string) (*domain.TransferProcess, error) {
	rows, err := s.db.QueryContext(ctx, query, arg)
	if err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, sql.ErrNoRows
	}

	tp, err := scanConsumerTransfer(rows)
	if err != nil {
		return nil, err
	}

	return tp, nil
}

func scanConsumerTransfer(scanner rowScanner) (*domain.TransferProcess, error) {
	var (
		id                  string
		contractAgreementID string
		correlationID       sql.NullString
		callbackAddress     sql.NullString
		assetID             string
		state               int
		destBytes           []byte
		sourceBytes         []byte
		errorDetail         sql.NullString
		edrToken            sql.NullString
		createdAt           time.Time
		updatedAt           time.Time
	)

	err := scanner.Scan(&id, &contractAgreementID, &correlationID, &callbackAddress, &assetID, &state, &destBytes, &sourceBytes, &errorDetail, &edrToken, &createdAt, &updatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to scan transfer process: %w", err)
	}

	var dest domain.DataAddress
	if len(destBytes) > 0 {
		if err := json.Unmarshal(destBytes, &dest); err != nil {
			return nil, fmt.Errorf("failed to unmarshal data destination: %w", err)
		}
	}

	var source domain.DataAddress
	if len(sourceBytes) > 0 {
		if err := json.Unmarshal(sourceBytes, &source); err != nil {
			return nil, fmt.Errorf("failed to unmarshal data source: %w", err)
		}
	}

	return &domain.TransferProcess{
		ID:                  id,
		ContractAgreementID: contractAgreementID,
		CorrelationID:       correlationID.String,
		CallbackAddress:     callbackAddress.String,
		AssetID:             assetID,
		State:               domain.TransferState(state),
		DataDestination:     dest,
		DataSource:          source,
		ErrorDetail:         errorDetail.String,
		CreatedAt:           createdAt,
		UpdatedAt:           updatedAt,
	}, nil
}
