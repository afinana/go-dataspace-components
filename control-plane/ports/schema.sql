-- PostgreSQL schema for the Control Plane services.
-- Stores Contract Negotiations and Transfer Processes with full state machine tracking.

CREATE TABLE IF NOT EXISTS contract_negotiations (
    id VARCHAR(255) PRIMARY KEY,
    correlation_id VARCHAR(255),
    counter_party VARCHAR(255) NOT NULL,
    type VARCHAR(20) NOT NULL,  -- CONSUMER or PROVIDER
    state INT NOT NULL DEFAULT 0,
    contract_offer JSONB,
    agreement JSONB,
    error_detail TEXT DEFAULT '',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_cn_correlation ON contract_negotiations (correlation_id);
CREATE INDEX IF NOT EXISTS idx_cn_state ON contract_negotiations (state);

CREATE TABLE IF NOT EXISTS transfer_processes (
    id VARCHAR(255) PRIMARY KEY,
    contract_agreement_id VARCHAR(255) NOT NULL,
    correlation_id VARCHAR(255),
    asset_id VARCHAR(255) NOT NULL,
    state INT NOT NULL DEFAULT 0,
    data_destination JSONB NOT NULL,
    data_source JSONB,
    error_detail TEXT DEFAULT '',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_tp_correlation ON transfer_processes (correlation_id);
CREATE INDEX IF NOT EXISTS idx_tp_state ON transfer_processes (state);
CREATE INDEX IF NOT EXISTS idx_tp_agreement ON transfer_processes (contract_agreement_id);
