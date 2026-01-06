package repository

import (
	"context"
	"fmt"

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

func (pr *PostgresRepository) UpdateIncidentByID(ctx context.Context, id string, entit *entities.UpdateIncident, exec Executor) (*entities.ReadIncident, error) {
	if exec == nil {
		exec = pr.db
	}
	query, args := pr.getQueryAndArgsForUpdate(entit, id)
	res := &entities.ReadIncident{}

	err := exec.QueryRowContext(ctx, query, args...).Scan(
		&res.Id,
		&res.Name,
		&res.Type,
		&res.Latitude,
		&res.Longitude,
		&res.Coordinates,
		&res.Description,
		&res.Radius,
		&res.IsActive,
		&res.Status,
		&res.CreatedDate,
		&res.UpdatedDate,
		&res.ResolvedDate)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (pr *PostgresRepository) getQueryAndArgsForUpdate(entit *entities.UpdateIncident, id string) (string, []any) {
	args := []any{}
	query := "UPDATE incidents "
	indexArg := 1

	if entit.Name != nil {
		query += fmt.Sprintf(", name=$%d", indexArg)
		indexArg++
		args = append(args, *entit.Name)
	}
	if entit.Type != nil {
		if indexArg == 1 {
			query += fmt.Sprintf("SET type=$%d", indexArg)
			indexArg++
			args = append(args, *entit.Type)
		} else {
			query += fmt.Sprintf(", type=$%d", indexArg)
			indexArg++
			args = append(args, *entit.Type)
		}
	}
	if entit.Description != nil {
		if indexArg == 1 {
			query += fmt.Sprintf("SET description=$%d", indexArg)
			indexArg++
			args = append(args, *entit.Description)
		} else {
			query += fmt.Sprintf(", description=$%d", indexArg)
			indexArg++
			args = append(args, *entit.Description)
		}
	}
	if entit.Radius != nil {
		if indexArg == 1 {
			query += fmt.Sprintf("SET radius=$%d", indexArg)
			indexArg++
			args = append(args, *entit.Radius)
		} else {
			query += fmt.Sprintf(", radius=$%d", indexArg)
			indexArg++
			args = append(args, *entit.Radius)
		}
	}
	if entit.Status != nil {
		if indexArg == 1 {
			query += fmt.Sprintf("SET status=$%d", indexArg)
			indexArg++
			args = append(args, *entit.Status)
		} else {
			query += fmt.Sprintf(", status=$%d", indexArg)
			indexArg++
			args = append(args, *entit.Status)
		}
	}

	if indexArg == 1 {
		query += fmt.Sprintf("SET is_active=$%d", indexArg)
		indexArg++
		args = append(args, entit.IsActive)
	} else {
		query += fmt.Sprintf(", is_active=$%d", indexArg)
		indexArg++
		args = append(args, entit.IsActive)
	}

	if indexArg == 1 {
		query += fmt.Sprintf("SET resolved_date=$%d", indexArg)
		indexArg++
		args = append(args, entit.ResolvedTime)
	} else {
		query += fmt.Sprintf(", resolved_date=$%d", indexArg)
		indexArg++
		args = append(args, entit.ResolvedTime)
	}

	if indexArg == 1 {
		query += "SET updated_date=NOW()"

	} else {
		query += ", updated_date=NOW()"
	}

	query += fmt.Sprintf(" WHERE id = $%d \nRETURNING id, name, type, latitude, longitude, coordinates, description, radius, is_active, status, created_date, updated_date, resolved_date", indexArg)
	args = append(args, id)
	return query, args
}
