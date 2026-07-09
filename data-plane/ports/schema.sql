-- PostgreSQL schema for the Data Plane service.
-- Stores active data flow sessions for proxy authentication and tracking.

CREATE TABLE IF NOT EXISTS data_flows (
    id VARCHAR(255) PRIMARY KEY,
    token VARCHAR(255) NOT NULL UNIQUE,
    contract_agreement_id VARCHAR(255),
    source_data_address JSONB NOT NULL,
    destination_data_address JSONB NOT NULL,
    properties JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_df_token ON data_flows (token);
