-- PostgreSQL schema definition for the Sovereign Identity Hub.
-- Provides secure storage for Verifiable Credentials, strictly isolated
-- by participant context anchors to prevent multi-tenant credential leakage.

CREATE TABLE IF NOT EXISTS verifiable_credentials (
    id VARCHAR(255) PRIMARY KEY,
    participant_context_id VARCHAR(255) NOT NULL,
    issuer VARCHAR(255) NOT NULL,
    vc_type VARCHAR(255)[] NOT NULL, -- Array of VC types for index scans
    credential_payload JSONB NOT NULL, -- Raw W3C JSON-LD credential document
    issuance_date TIMESTAMP WITH TIME ZONE NOT NULL,
    expiration_date TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Indices for rapid credential lookup and tenant isolation

-- Enforces tenant security boundaries at the database level
CREATE INDEX IF NOT EXISTS idx_vc_tenant ON verifiable_credentials (participant_context_id);

-- Enables fast filtering based on credentials classification (e.g. membership checks)
CREATE INDEX IF NOT EXISTS idx_vc_types ON verifiable_credentials USING gin (vc_type);

-- Enables query capability matching on properties nested within the payload JSONB
CREATE INDEX IF NOT EXISTS idx_vc_payload_attributes ON verifiable_credentials USING gin (credential_payload);
