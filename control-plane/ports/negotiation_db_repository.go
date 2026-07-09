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

// PostgresNegotiationStore manages persisting ContractNegotiation models to PostgreSQL.
type PostgresNegotiationStore struct {
	db *sql.DB
}

// NewPostgresNegotiationStore creates a new ContractNegotiationStore instance.
func NewPostgresNegotiationStore(db *sql.DB) *PostgresNegotiationStore {
	return &PostgresNegotiationStore{db: db}
}

// Save stores the ContractNegotiation in the database.
func (s *PostgresNegotiationStore) Save(ctx context.Context, cn *domain.ContractNegotiation) error {
	if cn.ID == "" {
		return errors.New("contract negotiation ID cannot be empty")
	}

	offerBytes, err := json.Marshal(cn.ContractOffer)
	if err != nil {
		return fmt.Errorf("failed to marshal contract offer: %w", err)
	}

	var agreementBytes []byte
	if cn.Agreement != nil {
		agreementBytes, err = json.Marshal(cn.Agreement)
		if err != nil {
			return fmt.Errorf("failed to marshal agreement: %w", err)
		}
	}

	query := `
		INSERT INTO contract_negotiations 
			(id, correlation_id, counter_party, type, state, contract_offer, agreement, error_detail, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (id) DO UPDATE SET
			correlation_id = EXCLUDED.correlation_id,
			counter_party = EXCLUDED.counter_party,
			type = EXCLUDED.type,
			state = EXCLUDED.state,
			contract_offer = EXCLUDED.contract_offer,
			agreement = EXCLUDED.agreement,
			error_detail = EXCLUDED.error_detail,
			updated_at = EXCLUDED.updated_at;
	`

	createdAt := cn.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now()
	}
	updatedAt := cn.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = time.Now()
	}

	_, err = s.db.ExecContext(ctx, query,
		cn.ID,
		cn.CorrelationID,
		cn.CounterParty,
		string(cn.Type),
		int(cn.State),
		offerBytes,
		agreementBytes,
		cn.ErrorDetail,
		createdAt,
		updatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to execute postgres insert for contract negotiation: %w", err)
	}

	return nil
}

// FindByID retrieves a ContractNegotiation by ID.
func (s *PostgresNegotiationStore) FindByID(ctx context.Context, id string) (*domain.ContractNegotiation, error) {
	query := `
		SELECT id, correlation_id, counter_party, type, state, contract_offer, agreement, error_detail, created_at, updated_at 
		FROM contract_negotiations 
		WHERE id = $1
	`
	return s.querySingle(ctx, query, id)
}

// FindByCorrelationID retrieves a ContractNegotiation by CorrelationID.
func (s *PostgresNegotiationStore) FindByCorrelationID(ctx context.Context, correlationID string) (*domain.ContractNegotiation, error) {
	query := `
		SELECT id, correlation_id, counter_party, type, state, contract_offer, agreement, error_detail, created_at, updated_at 
		FROM contract_negotiations 
		WHERE correlation_id = $1
	`
	return s.querySingle(ctx, query, correlationID)
}

// Update updates an existing ContractNegotiation state and timestamps.
func (s *PostgresNegotiationStore) Update(ctx context.Context, cn *domain.ContractNegotiation) error {
	offerBytes, err := json.Marshal(cn.ContractOffer)
	if err != nil {
		return fmt.Errorf("failed to marshal contract offer: %w", err)
	}

	var agreementBytes []byte
	if cn.Agreement != nil {
		agreementBytes, err = json.Marshal(cn.Agreement)
		if err != nil {
			return fmt.Errorf("failed to marshal agreement: %w", err)
		}
	}

	query := `
		UPDATE contract_negotiations 
		SET correlation_id = $2, counter_party = $3, type = $4, state = $5, contract_offer = $6, agreement = $7, error_detail = $8, updated_at = $9 
		WHERE id = $1
	`
	updatedAt := cn.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = time.Now()
	}

	res, err := s.db.ExecContext(ctx, query,
		cn.ID,
		cn.CorrelationID,
		cn.CounterParty,
		string(cn.Type),
		int(cn.State),
		offerBytes,
		agreementBytes,
		cn.ErrorDetail,
		updatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to execute postgres update for contract negotiation: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("contract negotiation not found")
	}

	return nil
}

// ListAll returns all stored ContractNegotiations.
func (s *PostgresNegotiationStore) ListAll(ctx context.Context) ([]domain.ContractNegotiation, error) {
	query := `
		SELECT id, correlation_id, counter_party, type, state, contract_offer, agreement, error_detail, created_at, updated_at 
		FROM contract_negotiations
	`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query contract negotiations: %w", err)
	}
	defer rows.Close()

	var result []domain.ContractNegotiation
	for rows.Next() {
		cn, err := scanNegotiation(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, *cn)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (s *PostgresNegotiationStore) querySingle(ctx context.Context, query string, arg string) (*domain.ContractNegotiation, error) {
	rows, err := s.db.QueryContext(ctx, query, arg)
	if err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, sql.ErrNoRows
	}

	cn, err := scanNegotiation(rows)
	if err != nil {
		return nil, err
	}

	return cn, nil
}

// scanNegotiation is a helper to scan rows into domain.ContractNegotiation.
type rowScanner interface {
	Scan(dest ...any) error
}

func scanNegotiation(scanner rowScanner) (*domain.ContractNegotiation, error) {
	var (
		id             string
		correlationID  sql.NullString
		counterParty   string
		typeName       string
		state          int
		offerBytes     []byte
		agreementBytes []byte
		errorDetail    sql.NullString
		createdAt      time.Time
		updatedAt      time.Time
	)

	err := scanner.Scan(&id, &correlationID, &counterParty, &typeName, &state, &offerBytes, &agreementBytes, &errorDetail, &createdAt, &updatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to scan contract negotiation: %w", err)
	}

	var offer *domain.ContractOffer
	if len(offerBytes) > 0 {
		offer = &domain.ContractOffer{}
		if err := json.Unmarshal(offerBytes, offer); err != nil {
			return nil, fmt.Errorf("failed to unmarshal contract offer: %w", err)
		}
	}

	var agreement *domain.ContractAgreement
	if len(agreementBytes) > 0 {
		agreement = &domain.ContractAgreement{}
		if err := json.Unmarshal(agreementBytes, agreement); err != nil {
			return nil, fmt.Errorf("failed to unmarshal contract agreement: %w", err)
		}
	}

	return &domain.ContractNegotiation{
		ID:            id,
		CorrelationID: correlationID.String,
		CounterParty:  counterParty,
		Type:          domain.NegotiationType(typeName),
		State:         domain.NegotiationState(state),
		ContractOffer: offer,
		Agreement:     agreement,
		ErrorDetail:   errorDetail.String,
		CreatedAt:     createdAt,
		UpdatedAt:     updatedAt,
	}, nil
}
