-- +goose Up
CREATE EXTENSION IF NOT EXISTS postgis;

-- +goose Down
SELECT 'down SQL query';
