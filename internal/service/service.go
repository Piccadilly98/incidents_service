package service

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/Piccadilly98/incidents_service/internal/models/dto"
	"github.com/Piccadilly98/incidents_service/internal/models/entities"
)

func (s *Service) RegistrationIncident(ctx context.Context, req *dto.RegistrationIncidentRequest) (*dto.IncidentAdminResponse, error) {
	err := req.Validate()
	if err != nil {
		return nil, err
	}
	entit, err := s.FromDtoToEntitie(req)
	if err != nil {
		return nil, err
	}
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()
	id, err := s.db.RegistrationIncident(ctx, entit, tx)
	if err != nil {
		return nil, err
	}
	res, err := s.db.GetInfoByIncidentID(ctx, id, tx)
	if err != nil {
		return nil, err
	}
	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	if s.cache != nil {
		if res.IsActive {
			err := s.cache.SetActiveIncident(ctx, res)
			if err != nil {
				s.cacheLogger.Printf("ERROR IN SET WITH ID: %s, err: %s\n", res.Id, err.Error())
			}
		}
	}
	s.changeLogger.Printf("INFO: Create new incident with id: %s", id)
	return dto.CreateAdminResponse(res, nil), nil
}

func (s *Service) FromDtoToEntitie(req *dto.RegistrationIncidentRequest) (*entities.RegistrationIncidentEntitie, error) {
	if len(req.Name) > 100 {
		return nil, fmt.Errorf("very long name")
	}
	if len(req.Type) > 100 {
		return nil, fmt.Errorf("very long type")
	}
	entit := req.ToBaseEntity()
	var err error
	entit.Status, err = s.processingStatus(req.Status)
	if err != nil {
		return nil, err
	}
	entit.Radius, err = s.processingRadius(req.RadiusInMeters)
	if err != nil {
		return nil, err
	}
	entit.IsActive, err = s.processingIsActive(entit.Status)
	if err != nil {
		return nil, err
	}
	entit.ResolvedTime = s.processingResolvedTime(entit.Status)
	return entit, nil
}

func (s *Service) processingStatus(statusReq *string) (string, error) {
	status := StatusActive
	if statusReq != nil {
		if len(*statusReq) > 20 {
			return "", fmt.Errorf("very long status")
		}
		if *statusReq != StatusActive && *statusReq != StatusResolved && *statusReq != StatusArchived {
			return "", fmt.Errorf("invalid status")
		}
		status = *statusReq
	}
	return status, nil
}

func (s *Service) processingRadius(radiusResp *int) (int, error) {
	radius := s.config.DefaultRadius

	if radiusResp != nil {
		if *radiusResp > s.config.MaxRadius {
			return 0, fmt.Errorf("radius cannot be > %d", s.config.MaxRadius)
		}
		if *radiusResp <= 0 {
			return 0, fmt.Errorf("radius cannot be <= 0")
		}
		radius = *radiusResp
	}

	return radius, nil
}

func (s *Service) processingIsActive(status string) (bool, error) {
	switch status {
	case StatusActive:
		return true, nil
	case StatusResolved:
		return false, nil
	case StatusArchived:
		return false, nil
	default:
		return false, fmt.Errorf("unexpected status")
	}
}

func (s *Service) processingResolvedTime(status string) *time.Time {
	var res *time.Time
	if status == StatusResolved || status == StatusArchived {
		now := time.Now().UTC()
		res = &now
	}

	return res
}

func (s *Service) GetIncidentInfoByID(ctx context.Context, id string) (*dto.IncidentAdminResponse, error) {
	var read *entities.ReadIncident
	var err error
	if s.cache != nil {
		read, err = s.cache.GetActiveIncident(ctx, id)
		if err != nil {
			read = nil
			s.cacheLogger.Printf("ERROR IN GET WITH ID: %s, err: %s\n", id, err.Error())
		}
	}
	if read == nil {
		read, err = s.db.GetInfoByIncidentID(ctx, id, nil)
		if err != nil {
			return nil, err
		}
		if s.cache != nil {
			if read.IsActive {
				err := s.cache.SetActiveIncident(ctx, read)
				if err != nil {
					s.cacheLogger.Printf("ERROR IN SET WITH ID: %s, err: %s\n", read.Id, err.Error())
				}
			}
		}
	}
	return dto.CreateAdminResponse(read, nil), nil
}

func (s *Service) UpdateIncidentByID(ctx context.Context, id string, req *dto.UpdateRequest) (*dto.IncidentAdminResponse, error) {
	err := req.Validate()
	if err != nil {
		return nil, err
	}
	var read *entities.ReadIncident
	if s.cache != nil {
		read, err = s.cache.GetActiveIncident(ctx, id)
		if err != nil {
			read = nil
			s.cacheLogger.Printf("ERROR IN SET WITH ID: %s, err: %s\n", id, err.Error())
		}
	}
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	if read == nil {
		read, err = s.db.GetInfoByIncidentID(ctx, id, tx)
		if err != nil {
			return nil, err
		}
	}
	if err := s.processingIncidentIDForUpdate(read, req, id); err != nil {
		return nil, err
	}
	res := s.toUpdateEntity(read, req)
	model, err := s.db.UpdateIncidentByID(ctx, id, res, tx)
	if err != nil {
		return nil, err
	}
	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	if s.cache != nil {
		if res.IsActive {
			err := s.cache.SetActiveIncident(ctx, model)
			if err != nil {
				s.cacheLogger.Printf("ERROR IN SET WITH ID: %s, err: %s\n", model.Id, err.Error())
			}
		}
	}
	s.changeLogger.Printf("INFO: incident %s updated successfully", id)
	return dto.CreateAdminResponse(model, nil), nil
}

func (s *Service) processingIncidentIDForUpdate(res *entities.ReadIncident, req *dto.UpdateRequest, id string) error {
	hasChanges := false
	if res.Status == StatusArchived {
		if req.Name != nil || req.Type != nil || req.Radius != nil || req.Status != nil {
			return fmt.Errorf("unable to update archived incident")
		}
	}
	if req.Description != nil && res.Description != nil {
		if *req.Description != *res.Description {
			hasChanges = true
		}
	} else if (req.Description != nil && res.Description == nil) || (req.Description == nil && res.Description != nil) {
		hasChanges = true
	}
	if req.Radius != nil {
		if *req.Radius > s.config.MaxRadius {
			return fmt.Errorf("radius cannot be > %d", s.config.MaxRadius)
		}
		if *req.Radius <= 0 {
			return fmt.Errorf("radius cannot be <= 0")
		}
		if *req.Radius != res.Radius {
			hasChanges = true
			s.changeLogger.Printf("INFO: incident id: %s, radius changed from %d to %d", id, res.Radius, *req.Radius)
		}
	}
	if req.Name != nil {
		if *req.Name != res.Name {
			hasChanges = true
		}
	}
	if req.Type != nil {
		if *req.Type != res.Type {
			hasChanges = true
		}
	}
	if req.Status != nil {
		if *req.Status != StatusActive && *req.Status != StatusResolved && *req.Status != StatusArchived {
			return fmt.Errorf("invalid status")
		}
		if *req.Status != res.Status {
			hasChanges = true
		}
	}
	if !hasChanges {
		return fmt.Errorf("no data for update")
	}
	return nil
}

func (s *Service) toUpdateEntity(res *entities.ReadIncident, req *dto.UpdateRequest) *entities.UpdateIncident {
	isActive := res.IsActive
	var resolvedTime *time.Time = res.ResolvedDate
	if req.Status != nil {
		switch *req.Status {
		case StatusResolved, StatusArchived:
			isActive = false
			now := time.Now().UTC()
			resolvedTime = &now
		case StatusActive:
			isActive = true
			resolvedTime = nil
		}
	}
	return req.ToEntity(resolvedTime, isActive)
}

func (s *Service) DeactivateIncidentByID(ctx context.Context, id string) (*dto.IncidentAdminResponse, error) {
	var err error
	var read *entities.ReadIncident
	if s.cache != nil {
		read, err = s.cache.GetActiveIncident(ctx, id)
		if err != nil {
			read = nil
			s.cacheLogger.Printf("ERROR IN SET WITH ID: %s, err: %s\n", id, err.Error())
		}
	}
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	if read == nil {
		read, err = s.db.GetInfoByIncidentID(ctx, id, tx)
		if err != nil {
			return nil, err
		}
	}
	if read.Status == StatusArchived {
		return nil, fmt.Errorf("incident already archived")
	}
	archived := StatusArchived
	req := &dto.UpdateRequest{
		Status: &archived,
	}

	if err := s.processingIncidentIDForUpdate(read, req, id); err != nil {
		return nil, err
	}

	updateEntity := s.toUpdateEntity(read, req)

	updated, err := s.db.UpdateIncidentByID(ctx, id, updateEntity, tx)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}
	if s.cache != nil {
		err := s.cache.DeleteActiveIncident(ctx, id)
		if err != nil {
			s.cacheLogger.Printf("ERROR IN DEL WITH ID: %s, err: %s\n", id, err.Error())
		}
	}
	s.changeLogger.Printf("INFO: incident %s deactivated", id)

	return dto.CreateAdminResponse(updated, nil), nil
}

func (s *Service) DeleteIncidentByID(ctx context.Context, id string) error {
	err := s.db.DeleteIncidentByID(ctx, id, nil)
	if err != nil {
		return err
	}
	if s.cache != nil {
		err := s.cache.DeleteActiveIncident(ctx, id)
		if err != nil {
			s.cacheLogger.Printf("ERROR IN DEL WITH ID: %s, err: %s\n", id, err.Error())
		}
	}
	s.changeLogger.Printf("CRITICAL: incident %s force deleted", id)
	return nil
}

func (s *Service) GetPagination(ctx context.Context, query *dto.PaginationQueryParams) (*dto.PaginationResponse, error) {
	if query.Radius != nil {
		_, err := s.processingRadius(query.Radius)
		if err != nil {
			return nil, err
		}
	}
	if query.Status != "" {
		if query.Status != StatusActive && query.Status != StatusArchived && query.Status != StatusResolved {
			return nil, fmt.Errorf("invalid status")
		}
	}
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	count, err := s.db.GetCountRows(ctx, tx)
	if err != nil {
		return nil, err
	}
	offset := 0
	limit := 0
	pages := 0
	var pageNum *int
	pages = s.GetCountPages(count)
	if query.PageNum != nil {
		if *query.PageNum > pages {
			return nil, fmt.Errorf("invalid page: max %d", pages)
		}
		pageNum = query.PageNum
		offset = s.config.MaxRowsInPage * (*query.PageNum - 1)
		limit = s.config.MaxRowsInPage
	}
	entit := s.toPaginationEntity(query, offset, limit)
	read, err := s.db.GetPaginationIncidentsInfo(ctx, entit, tx)
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}

	res := []*dto.IncidentAdminResponse{}

	for _, model := range read {
		if model.IsActive {
			if s.cache != nil {
				err := s.cache.SetActiveIncident(ctx, model)
				if err != nil {
					s.cacheLogger.Printf("ERROR IN SET WITH ID: %s, err: %s\n", model.Id, err.Error())
				}
			}
		}
		dto := dto.CreateAdminResponse(model, nil)
		res = append(res, dto)
	}
	return dto.ToPaginationResponse(res, pages, count, pageNum), nil
}

func (s *Service) GetCountPages(countRows int) int {
	res := float64(countRows) / float64(s.config.MaxRowsInPage)
	integerPart, fractionalPart := math.Modf(res)
	if fractionalPart != 0 {
		integerPart++
	}
	return int(integerPart)
}

func (s *Service) toPaginationEntity(query *dto.PaginationQueryParams, offset, limit int) *entities.PaginationIncidents {
	id := ""
	if query.ID != nil {
		id = *query.ID
	}
	res := &entities.PaginationIncidents{
		Offset: offset,
		Limit:  limit,
		Status: query.Status,
		Name:   query.Name,
		Type:   query.Type,
		Radius: query.Radius,
		ID:     id,
	}
	return res
}

func (s *Service) LocationCheck(ctx context.Context, req *dto.LocationCheckRequest) (*dto.LocationCheckResponse, error) {
	err := req.Validate()
	if err != nil {
		return nil, err
	}

	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	checkId, err := s.db.RegistrationCheck(ctx, req.UserID, req.Latitude, req.Longitude, tx)
	if err != nil {
		return nil, err
	}
	if checkId == "" {
		return nil, fmt.Errorf("empty id before db method")
	}
	s.changeLogger.Printf("INFO: Create new check with id: %s", checkId)
	destChecks, err := s.db.GetDetectedIncidents(ctx, req.Longitude, req.Latitude, tx)
	if err != nil {
		return nil, err
	}
	isDanger := false
	if len(destChecks) > 0 {
		isDanger = true
	}
	dangersIds := []string{}
	for _, check := range destChecks {
		dangersIds = append(dangersIds, check.Incident.Id)
	}

	err = s.db.UpdateCheckByID(ctx, dangersIds, checkId, isDanger, tx)
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	res := &dto.LocationCheckResponse{
		ID:        checkId,
		UserID:    req.UserID,
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
		IsDanger:  isDanger,
	}
	userIncidents := []*dto.IncidentUserResponse{}
	for _, incident := range destChecks {
		userIncidents = append(userIncidents, dto.CreateUserResponse(&incident.Incident, &incident.Distance))
	}
	res.DetectedIncidentsID = userIncidents
	if isDanger && s.wm != nil {
		s.wm.AddToQueue(*res, "", "")
	}
	return res, nil
}

func (s *Service) GetChecksStatistics(ctx context.Context) (*dto.IncidentsStatResponse, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	count, err := s.db.GetCountUniqueUsers(ctx, tx)
	if err != nil {
		return nil, err
	}
	fromTime := time.Now().UTC()
	statistics, err := s.db.GetStaticsForIncidentsWithTimeWindow(ctx, tx, s.config.StatsTimeWindow)
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return dto.ToIncidentsStatResponse(statistics, s.config.StatsTimeWindow, fromTime, count), nil

}
