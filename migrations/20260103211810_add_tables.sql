-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS incidents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    type VARCHAR(100) NOT NULL,
    latitude DECIMAL(10, 8) CHECK (latitude BETWEEN -90 AND 90),
    longitude DECIMAL(11, 8) CHECK (longitude BETWEEN -180 AND 180),
    coordinates GEOGRAPHY(POINT, 4326) GENERATED ALWAYS AS (ST_SetSRID(ST_MakePoint(longitude, latitude), 4326)
    ) STORED,
    description TEXT,
    radius INTEGER NOT NULL,
    is_active BOOLEAN DEFAULT true,
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'resolved')),
    created_date TIMESTAMP DEFAULT NOW(),
    updated_date TIMESTAMP,
    resolved_date TIMESTAMP
);
CREATE TABLE IF NOT EXISTS checks(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(100) NOT NULL, 
    latitude DECIMAL(10, 8) NOT NULL CHECK (latitude BETWEEN -90 AND 90),
    longitude DECIMAL(11, 8) NOT NULL CHECK (longitude BETWEEN -180 AND 180),
    coordinates GEOGRAPHY(POINT, 4326) GENERATED ALWAYS AS (ST_SetSRID(ST_MakePoint(longitude, latitude), 4326)
    ) STORED,
    is_danger BOOLEAN DEFAULT false,
    detected_incident_ids UUID[] DEFAULT '{}',
    created_date TIMESTAMP DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS incidents;
DROP TABLE IF EXISTS checks;
-- +goose StatementEnd
