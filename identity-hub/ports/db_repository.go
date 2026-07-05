package ports

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/afinana/go-dataspace-components/identity-hub/domain"
)

// PostgresVCStore manages the database connection and queries to persist VCs.
type PostgresVCStore struct {
	db *sql.DB
}

// NewPostgresVCStore creates a new storage repository instance.
func NewPostgresVCStore(db *sql.DB) *PostgresVCStore {
	return &PostgresVCStore{db: db}
}

// Save stores the Verifiable Credential payload in PostgreSQL under tenant context.
func (s *PostgresVCStore) Save(ctx context.Context, participantContextID string, vc *domain.VerifiableCredential) error {
	payloadBytes, err := json.Marshal(vc)
	if err != nil {
		return fmt.Errorf("failed to marshal credential payload for database write: %w", err)
	}

	// Prepare simple types slice for array database insertion
	// Convert types slice to pg array format e.g. {"VerifiableCredential", "XDataShareMembershipCredential"}
	typesList := fmt.Sprintf("{%s}", stringsJoinQuotes(vc.Type))

	query := `
		INSERT INTO verifiable_credentials 
			(id, participant_context_id, issuer, vc_type, credential_payload, issuance_date, expiration_date) 
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE SET 
			credential_payload = EXCLUDED.credential_payload,
			updated_at = CURRENT_TIMESTAMP;
	`

	_, err = s.db.ExecContext(ctx, query,
		vc.ID,
		participantContextID,
		vc.Issuer,
		typesList,
		payloadBytes,
		vc.IssuanceDate,
		vc.ExpirationDate,
	)

	if err != nil {
		return fmt.Errorf("failed to execute postgres insert: %w", err)
	}

	return nil
}

// FindByScope retrieves all credentials matching tenant context and credential types.
func (s *PostgresVCStore) FindByScope(ctx context.Context, participantContextID string, scope string) ([]domain.VerifiableCredential, error) {
	// Query searches JSONB payload or checks vc_type array matching scope
	query := `
		SELECT credential_payload 
		FROM verifiable_credentials 
		WHERE participant_context_id = $1 
		  AND (vc_type @> ARRAY[$2]::VARCHAR[] OR credential_payload->'type' @> jsonb_build_array($2));
	`

	rows, err := s.db.QueryContext(ctx, query, participantContextID, scope)
	if err != nil {
		return nil, fmt.Errorf("failed to query verifiable credentials: %w", err)
	}
	defer rows.Close()

	var result []domain.VerifiableCredential
	for rows.Next() {
		var payload []byte
		if err := rows.Scan(&payload); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		var vc domain.VerifiableCredential
		if err := json.Unmarshal(payload, &vc); err != nil {
			return nil, fmt.Errorf("failed to unmarshal JSONB payload to credential struct: %w", err)
		}

		result = append(result, vc)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows traversal error: %w", err)
	}

	return result, nil
}

// Helper to join types with commas for postgres arrays
func stringsJoinQuotes(items []string) string {
	var escaped []string
	for _, it := range items {
		escaped = append(escaped, fmt.Sprintf(`"%s"`, it))
	}
	// Return comma separated string of elements
	if len(escaped) == 0 {
		return ""
	}
	return joinStringSlice(escaped, ",")
}

func joinStringSlice(slice []string, sep string) string {
	if len(slice) == 0 {
		return ""
	}
	res := slice[0]
	for _, s := range slice[1:] {
		res += sep + s
	}
	return res
}
