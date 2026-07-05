-- PostgreSQL schema definition for the Sovereign Catalog service.
-- Stores W3C DCAT-AP datasets and data services.

CREATE TABLE IF NOT EXISTS datasets (
    id VARCHAR(255) PRIMARY KEY,
    payload JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS data_services (
    id VARCHAR(255) PRIMARY KEY,
    payload JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Index for fast retrieval on JSONB payload attributes
CREATE INDEX IF NOT EXISTS idx_datasets_payload ON datasets USING gin (payload);
CREATE INDEX IF NOT EXISTS idx_data_services_payload ON data_services USING gin (payload);
