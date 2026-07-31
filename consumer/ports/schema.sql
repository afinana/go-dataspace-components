CREATE TABLE IF NOT EXISTS consumer_contract_negotiations (
    id VARCHAR(255) PRIMARY KEY,
    correlation_id VARCHAR(255) UNIQUE,
    counter_party VARCHAR(255) NOT NULL,
    callback_address VARCHAR(512),
    type VARCHAR(50) NOT NULL,
    state INT NOT NULL,
    contract_offer JSONB,
    agreement JSONB,
    error_detail TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS consumer_transfer_processes (
    id VARCHAR(255) PRIMARY KEY,
    contract_agreement_id VARCHAR(255) NOT NULL,
    correlation_id VARCHAR(255) UNIQUE,
    callback_address VARCHAR(512),
    asset_id VARCHAR(255) NOT NULL,
    state INT NOT NULL,
    data_destination JSONB NOT NULL,
    data_source JSONB,
    error_detail TEXT,
    edr_token TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
