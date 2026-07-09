package ports

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	dp "github.com/afinana/go-dataspace-components/data-plane/domain"
)

// PostgresDataFlowStore persists active data flows to PostgreSQL.
type PostgresDataFlowStore struct {
	db *sql.DB
}

// NewPostgresDataFlowStore creates a new storage repository instance.
func NewPostgresDataFlowStore(db *sql.DB) *PostgresDataFlowStore {
	return &PostgresDataFlowStore{db: db}
}

// Save stores the active data flow in PostgreSQL.
func (s *PostgresDataFlowStore) Save(ctx context.Context, token string, req *dp.DataFlowRequest) error {
	if req.ID == "" {
		return errors.New("data flow request ID cannot be empty")
	}

	sourceBytes, err := json.Marshal(req.SourceDataAddress)
	if err != nil {
		return fmt.Errorf("failed to marshal source data address: %w", err)
	}

	destBytes, err := json.Marshal(req.DestinationDataAddress)
	if err != nil {
		return fmt.Errorf("failed to marshal destination data address: %w", err)
	}

	propBytes, err := json.Marshal(req.Properties)
	if err != nil {
		return fmt.Errorf("failed to marshal flow properties: %w", err)
	}

	query := `
		INSERT INTO data_flows 
			(id, token, contract_agreement_id, source_data_address, destination_data_address, properties) 
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO UPDATE SET
			token = EXCLUDED.token,
			contract_agreement_id = EXCLUDED.contract_agreement_id,
			source_data_address = EXCLUDED.source_data_address,
			destination_data_address = EXCLUDED.destination_data_address,
			properties = EXCLUDED.properties,
			updated_at = CURRENT_TIMESTAMP;
	`

	_, err = s.db.ExecContext(ctx, query,
		req.ID,
		token,
		req.ContractAgreementID,
		sourceBytes,
		destBytes,
		propBytes,
	)
	if err != nil {
		return fmt.Errorf("failed to execute postgres insert for data flow: %w", err)
	}

	return nil
}

// FindByToken retrieves a data flow by security EDR token.
func (s *PostgresDataFlowStore) FindByToken(ctx context.Context, token string) (*dp.DataFlowRequest, error) {
	query := `
		SELECT id, contract_agreement_id, source_data_address, destination_data_address, properties 
		FROM data_flows 
		WHERE token = $1
	`
	var (
		id                  string
		contractAgreementID sql.NullString
		sourceBytes         []byte
		destBytes           []byte
		propBytes           []byte
	)

	err := s.db.QueryRowContext(ctx, query, token).Scan(
		&id, &contractAgreementID, &sourceBytes, &destBytes, &propBytes,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("data flow not found")
		}
		return nil, fmt.Errorf("failed to query data flow by token: %w", err)
	}

	var req dp.DataFlowRequest
	req.ID = id
	req.ContractAgreementID = contractAgreementID.String

	if err := json.Unmarshal(sourceBytes, &req.SourceDataAddress); err != nil {
		return nil, fmt.Errorf("failed to unmarshal source data address: %w", err)
	}
	if err := json.Unmarshal(destBytes, &req.DestinationDataAddress); err != nil {
		return nil, fmt.Errorf("failed to unmarshal destination data address: %w", err)
	}
	if len(propBytes) > 0 {
		if err := json.Unmarshal(propBytes, &req.Properties); err != nil {
			return nil, fmt.Errorf("failed to unmarshal properties: %w", err)
		}
	}

	return &req, nil
}

// FindByFlowID retrieves a data flow and its token by flow ID.
func (s *PostgresDataFlowStore) FindByFlowID(ctx context.Context, id string) (string, *dp.DataFlowRequest, error) {
	query := `
		SELECT token, contract_agreement_id, source_data_address, destination_data_address, properties 
		FROM data_flows 
		WHERE id = $1
	`
	var (
		token               string
		contractAgreementID sql.NullString
		sourceBytes         []byte
		destBytes           []byte
		propBytes           []byte
	)

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&token, &contractAgreementID, &sourceBytes, &destBytes, &propBytes,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil, errors.New("data flow not found")
		}
		return "", nil, fmt.Errorf("failed to query data flow by id: %w", err)
	}

	var req dp.DataFlowRequest
	req.ID = id
	req.ContractAgreementID = contractAgreementID.String

	if err := json.Unmarshal(sourceBytes, &req.SourceDataAddress); err != nil {
		return "", nil, fmt.Errorf("failed to unmarshal source data address: %w", err)
	}
	if err := json.Unmarshal(destBytes, &req.DestinationDataAddress); err != nil {
		return "", nil, fmt.Errorf("failed to unmarshal destination data address: %w", err)
	}
	if len(propBytes) > 0 {
		if err := json.Unmarshal(propBytes, &req.Properties); err != nil {
			return "", nil, fmt.Errorf("failed to unmarshal properties: %w", err)
		}
	}

	return token, &req, nil
}

// ListAll retrieves all active data flows mapped by token.
func (s *PostgresDataFlowStore) ListAll(ctx context.Context) (map[string]*dp.DataFlowRequest, error) {
	query := `
		SELECT id, token, contract_agreement_id, source_data_address, destination_data_address, properties 
		FROM data_flows
	`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query all data flows: %w", err)
	}
	defer rows.Close()

	result := make(map[string]*dp.DataFlowRequest)
	for rows.Next() {
		var (
			id                  string
			token               string
			contractAgreementID sql.NullString
			sourceBytes         []byte
			destBytes           []byte
			propBytes           []byte
		)

		err := rows.Scan(&id, &token, &contractAgreementID, &sourceBytes, &destBytes, &propBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to scan data flow: %w", err)
		}

		var req dp.DataFlowRequest
		req.ID = id
		req.ContractAgreementID = contractAgreementID.String

		if err := json.Unmarshal(sourceBytes, &req.SourceDataAddress); err != nil {
			return nil, fmt.Errorf("failed to unmarshal source data address: %w", err)
		}
		if err := json.Unmarshal(destBytes, &req.DestinationDataAddress); err != nil {
			return nil, fmt.Errorf("failed to unmarshal destination data address: %w", err)
		}
		if len(propBytes) > 0 {
			if err := json.Unmarshal(propBytes, &req.Properties); err != nil {
				return nil, fmt.Errorf("failed to unmarshal properties: %w", err)
			}
		}

		result[token] = &req
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

// Delete removes a data flow mapping when terminated.
func (s *PostgresDataFlowStore) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM data_flows WHERE id = $1`
	res, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to execute postgres delete for data flow: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("data flow not found")
	}

	return nil
}
