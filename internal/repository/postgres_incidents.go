package repository

import (
	"context"

	"github.com/Piccadilly98/incidents_service/internal/models/entities"
)

func (pr *PostgresRepository) RegistrationIncident(ctx context.Context, entit *entities.RegistrationIncidentEntitie, exec Executor) (string, error) {
	var id string
	if exec == nil {
		exec = pr.db
	}
	err := exec.QueryRowContext(ctx, `
	INSERT INTO incidents(name, type, description, latitude,longitude, radius, is_active, status, resolved_date)
	VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9)
	RETURNING id;
	`,
		entit.Name,
		entit.Type,
		entit.Description,
		entit.Latitude,
		entit.Longitude,
		entit.Radius,
		entit.IsActive,
		entit.Status,
		entit.ResolvedTime,
	).Scan(&id)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (pr *PostgresRepository) GetInfoByIncidentID(ctx context.Context, id string, exec Executor) (*entities.ReadIncident, error) {
	if exec == nil {
		exec = pr.db
	}
	res := &entities.ReadIncident{}
	err := exec.QueryRowContext(ctx, `
	SELECT id, name, type, latitude, longitude, coordinates, description, radius, is_active, status, created_date, updated_date, resolved_date FROM incidents
	WHERE id = $1;`, id).Scan(&res.Id, &res.Name, &res.Type, &res.Latitude, &res.Longitude, &res.Coordinates, &res.Description, &res.Radius, &res.IsActive, &res.Status, &res.CreatedDate, &res.UpdatedDate, &res.ResolvedDate)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (pr *PostgresRepository) GetExistByIncidentID(ctx context.Context, id string, exec Executor) (bool, error) {
	if exec == nil {
		exec = pr.db
	}

	var exists bool

	err := exec.QueryRowContext(ctx,
		`SELECT EXISTS (SELECT 1 FROM incidents WHERE id = $1)`,
		id).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}
