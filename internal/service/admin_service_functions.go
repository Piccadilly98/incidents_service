package service

import (
	"context"
	"fmt"
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
	s.changeLogger.Printf("INFO: Create new incident with id: %s", id)
	return dto.CreateAdminResponse(res), nil
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

// func (s *Service) GetExistsIncidentByID(ctx context.Context, id string) (bool, error) {
// 	return s.db.GetExistByIncidentID(ctx, id, nil)
// }

func (s *Service) GetIncidentInfoByID(ctx context.Context, id string) (*dto.IncidentAdminResponse, error) {
	res, err := s.db.GetInfoByIncidentID(ctx, id, nil)
	if err != nil {
		return nil, err
	}
	return dto.CreateAdminResponse(res), nil
}

func (s *Service) UpdateIncidentByID(ctx context.Context, id string, req *dto.UpdateRequest) (*dto.IncidentAdminResponse, error) {
	err := req.Validate()
	if err != nil {
		return nil, err
	}

	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	read, err := s.db.GetInfoByIncidentID(ctx, id, tx)
	if err != nil {
		return nil, err
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
	s.changeLogger.Printf("INFO: incident %s updated successfully", id)
	return dto.CreateAdminResponse(model), nil
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
		if *req.Status == StatusResolved || *req.Status == StatusArchived {
			isActive = false
			now := time.Now().UTC()
			resolvedTime = &now
		} else if *req.Status == StatusActive {
			isActive = true
			resolvedTime = nil
		}
	}
	return req.ToEntity(resolvedTime, isActive)
}
