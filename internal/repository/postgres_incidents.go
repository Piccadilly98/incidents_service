package repository

import (
	"context"
	"fmt"

	"github.com/Piccadilly98/incidents_service/internal/models/entities"
	"github.com/lib/pq"
)

// FOR UPDATE !!

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
		query += fmt.Sprintf("SET name=$%d", indexArg)
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

	query += fmt.Sprintf(" WHERE id = $%d RETURNING id, name, type, latitude, longitude, coordinates, description, radius, is_active, status, created_date, updated_date, resolved_date", indexArg)
	args = append(args, id)
	return query, args
}

func (pr *PostgresRepository) DeleteIncidentByID(ctx context.Context, id string, exec Executor) error {
	if exec == nil {
		exec = pr.db
	}

	_, err := exec.ExecContext(ctx, `DELETE FROM incidents WHERE id = $1;`, id)
	return err
}

func (pr *PostgresRepository) GetCountRows(ctx context.Context, exec Executor) (int, error) {
	if exec == nil {
		exec = pr.db
	}
	var result int
	err := exec.QueryRowContext(ctx, `SELECT COUNT(*) FROM incidents;`).Scan(&result)
	if err != nil {
		return 0, err
	}
	return result, nil
}

func (pr *PostgresRepository) GetPaginationIncidentsInfo(ctx context.Context, entit *entities.PaginationIncidents, exec Executor) ([]*entities.ReadIncident, error) {
	if exec == nil {
		exec = pr.db
	}
	incidents := []*entities.ReadIncident{}
	query, args := pr.getQueryAndArgsForPagination(entit)
	rows, err := exec.QueryContext(ctx, query, args...)

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		res := &entities.ReadIncident{}

		err := rows.Scan(&res.Id,
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
			&res.ResolvedDate,
		)

		if err != nil {
			return nil, err
		}

		incidents = append(incidents, res)
	}
	return incidents, err
}

func (pr *PostgresRepository) getQueryAndArgsForPagination(entit *entities.PaginationIncidents) (string, []any) {
	args := []any{}
	indexArg := 1

	query := "SELECT * FROM incidents"
	if entit.ID != "" {
		query += fmt.Sprintf(" WHERE id=$%d", indexArg)
		args = append(args, entit.ID)
		indexArg++
	}
	if entit.Status != "" {
		if indexArg == 1 {
			query += fmt.Sprintf(" WHERE status=$%d", indexArg)
			args = append(args, entit.Status)
			indexArg++
		} else {
			query += fmt.Sprintf(" AND status=$%d", indexArg)
			args = append(args, entit.Status)
			indexArg++
		}
	}

	if entit.Type != "" {
		if indexArg == 1 {
			query += fmt.Sprintf(" WHERE type=$%d", indexArg)
			args = append(args, entit.Type)
			indexArg++
		} else {
			query += fmt.Sprintf(" AND type=$%d", indexArg)
			args = append(args, entit.Type)
			indexArg++
		}
	}

	if entit.Name != "" {
		if indexArg == 1 {
			query += fmt.Sprintf(" WHERE name=$%d", indexArg)
			args = append(args, entit.Name)
			indexArg++
		} else {
			query += fmt.Sprintf(" AND name=$%d", indexArg)
			args = append(args, entit.Name)
			indexArg++
		}
	}
	if entit.Radius != nil {
		if indexArg == 1 {
			query += fmt.Sprintf(" WHERE radius=$%d", indexArg)
			args = append(args, *entit.Radius)
			indexArg++
		} else {
			query += fmt.Sprintf(" AND radius=$%d", indexArg)
			args = append(args, *entit.Radius)
			indexArg++
		}
	}
	if entit.Limit != 0 {
		query += fmt.Sprintf(" LIMIT $%d", indexArg)
		args = append(args, entit.Limit)
		indexArg++
	}
	if entit.Offset != 0 {
		query += fmt.Sprintf(" OFFSET $%d", indexArg)
		args = append(args, entit.Offset)
		indexArg++
	}
	query += ";"
	return query, args
}

func (pr *PostgresRepository) RegistrationCheck(ctx context.Context, userID, latitude, longitude string, exec Executor) (string, error) {
	if exec == nil {
		exec = pr.db
	}
	var checkId string

	err := exec.QueryRowContext(ctx,
		`INSERT INTO checks(user_id, latitude, longitude)
		VALUES($1, $2, $3)
		RETURNING id;
		`, userID, latitude, longitude).Scan(&checkId)
	if err != nil {
		return "", err
	}

	return checkId, nil
}

func (pr *PostgresRepository) GetDetectedIncidents(ctx context.Context, longitude, latitude string, exec Executor) ([]*entities.DistanceCheck, error) {
	if exec == nil {
		exec = pr.db
	}

	rows, err := exec.QueryContext(ctx,
		`SELECT 
		id, 
		name, 
		type, 
		latitude, 
		longitude, 
		coordinates, 
		description, 
		radius, 
		is_active, 
		status, 
		created_date, 
		updated_date, 
		resolved_date,
		ST_Distance(
			coordinates,
			ST_MakePoint($1, $2)::geography
		)AS distance
		FROM incidents
	WHERE is_active = true 
	AND ST_DWithin(
		coordinates,
		ST_MakePoint($1, $2)::geography,
		radius
		)
	ORDER BY distance;;`, longitude, latitude,
	)

	if err != nil {
		return nil, err
	}

	incidents := []*entities.DistanceCheck{}

	for rows.Next() {
		defer rows.Close()
		res := &entities.DistanceCheck{}

		err := rows.Scan(
			&res.Incident.Id,
			&res.Incident.Name,
			&res.Incident.Type,
			&res.Incident.Latitude,
			&res.Incident.Longitude,
			&res.Incident.Coordinates,
			&res.Incident.Description,
			&res.Incident.Radius,
			&res.Incident.IsActive,
			&res.Incident.Status,
			&res.Incident.CreatedDate,
			&res.Incident.UpdatedDate,
			&res.Incident.ResolvedDate,
			&res.Distance,
		)

		if err != nil {
			return nil, err
		}

		incidents = append(incidents, res)
	}

	return incidents, nil
}

func (pr *PostgresRepository) UpdateCheckByID(ctx context.Context, dangersIds []string, checkId string, isDanger bool, exec Executor) error {
	if exec == nil {
		exec = pr.db
	}
	if dangersIds == nil {
		dangersIds = []string{}
	}
	_, err := exec.ExecContext(ctx,
		`UPDATE checks
		SET is_danger = $1, detected_incident_ids = $2 
		WHERE id = $3;`,
		isDanger,
		pq.Array(dangersIds),
		checkId,
	)
	return err
}
