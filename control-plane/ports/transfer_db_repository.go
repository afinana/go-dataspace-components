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

// PostgresTransferStore manages persisting TransferProcess models to PostgreSQL.
type PostgresTransferStore struct {
	db *sql.DB
}

// NewPostgresTransferStore creates a new TransferProcessStore instance.
func NewPostgresTransferStore(db *sql.DB) *PostgresTransferStore {
	return &PostgresTransferStore{db: db}
}

// Save stores the TransferProcess in the database.
func (s *PostgresTransferStore) Save(ctx context.Context, tp *domain.TransferProcess) error {
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
		INSERT INTO transfer_processes 
			(id, contract_agreement_id, correlation_id, asset_id, state, data_destination, data_source, error_detail, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (id) DO UPDATE SET
			contract_agreement_id = EXCLUDED.contract_agreement_id,
			correlation_id = EXCLUDED.correlation_id,
			asset_id = EXCLUDED.asset_id,
			state = EXCLUDED.state,
			data_destination = EXCLUDED.data_destination,
			data_source = EXCLUDED.data_source,
			error_detail = EXCLUDED.error_detail,
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

	_, err = s.db.ExecContext(ctx, query,
		tp.ID,
		tp.ContractAgreementID,
		tp.CorrelationID,
		tp.AssetID,
		int(tp.State),
		destBytes,
		sourceBytes,
		tp.ErrorDetail,
		createdAt,
		updatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to execute postgres insert for transfer process: %w", err)
	}

	return nil
}

// FindByID retrieves a TransferProcess by ID.
func (s *PostgresTransferStore) FindByID(ctx context.Context, id string) (*domain.TransferProcess, error) {
	query := `
		SELECT id, contract_agreement_id, correlation_id, asset_id, state, data_destination, data_source, error_detail, created_at, updated_at 
		FROM transfer_processes 
		WHERE id = $1
	`
	return s.querySingle(ctx, query, id)
}

// FindByCorrelationID retrieves a TransferProcess by CorrelationID.
func (s *PostgresTransferStore) FindByCorrelationID(ctx context.Context, correlationID string) (*domain.TransferProcess, error) {
	query := `
		SELECT id, contract_agreement_id, correlation_id, asset_id, state, data_destination, data_source, error_detail, created_at, updated_at 
		FROM transfer_processes 
		WHERE correlation_id = $1
	`
	return s.querySingle(ctx, query, correlationID)
}

// Update updates an existing TransferProcess.
func (s *PostgresTransferStore) Update(ctx context.Context, tp *domain.TransferProcess) error {
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
		UPDATE transfer_processes 
		SET contract_agreement_id = $2, correlation_id = $3, asset_id = $4, state = $5, data_destination = $6, data_source = $7, error_detail = $8, updated_at = $9 
		WHERE id = $1
	`
	updatedAt := tp.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = time.Now()
	}

	res, err := s.db.ExecContext(ctx, query,
		tp.ID,
		tp.ContractAgreementID,
		tp.CorrelationID,
		tp.AssetID,
		int(tp.State),
		destBytes,
		sourceBytes,
		tp.ErrorDetail,
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
func (s *PostgresTransferStore) ListAll(ctx context.Context) ([]domain.TransferProcess, error) {
	query := `
		SELECT id, contract_agreement_id, correlation_id, asset_id, state, data_destination, data_source, error_detail, created_at, updated_at 
		FROM transfer_processes
	`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query transfer processes: %w", err)
	}
	defer rows.Close()

	var result []domain.TransferProcess
	for rows.Next() {
		tp, err := scanTransfer(rows)
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

func (s *PostgresTransferStore) querySingle(ctx context.Context, query string, arg string) (*domain.TransferProcess, error) {
	rows, err := s.db.QueryContext(ctx, query, arg)
	if err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, sql.ErrNoRows
	}

	tp, err := scanTransfer(rows)
	if err != nil {
		return nil, err
	}

	return tp, nil
}

func scanTransfer(scanner rowScanner) (*domain.TransferProcess, error) {
	var (
		id                  string
		contractAgreementID string
		correlationID       sql.NullString
		assetID             string
		state               int
		destBytes           []byte
		sourceBytes         []byte
		errorDetail         sql.NullString
		createdAt           time.Time
		updatedAt           time.Time
	)

	err := scanner.Scan(&id, &contractAgreementID, &correlationID, &assetID, &state, &destBytes, &sourceBytes, &errorDetail, &createdAt, &updatedAt)
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
		AssetID:             assetID,
		State:               domain.TransferState(state),
		DataDestination:     dest,
		DataSource:          source,
		ErrorDetail:         errorDetail.String,
		CreatedAt:           createdAt,
		UpdatedAt:           updatedAt,
	}, nil
}
