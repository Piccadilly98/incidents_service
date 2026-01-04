-- +goose Up
-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_incidents_coords ON incidents USING GIST (coordinates);
CREATE INDEX IF NOT EXISTS idx_checks_coords ON checks USING GIST (coordinates);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_incidents_coords;
DROP INDEX IF EXISTS idx_checks_coords;
-- +goose StatementEnd
