-- +goose Up
CREATE TABLE metrics (
    id VARCHAR(255) PRIMARY KEY,
    metric_type VARCHAR(255) NOT NULL,
    delta BIGINT,
    value DOUBLE precision,
    hash VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE metrics;
