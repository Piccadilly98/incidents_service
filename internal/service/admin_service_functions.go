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
	entit.Status, err = s.processingStatus(req)
	if err != nil {
		return nil, err
	}
	entit.Radius, err = s.processingRadius(req)
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

func (s *Service) processingStatus(req *dto.RegistrationIncidentRequest) (string, error) {
	status := StatusActive
	if req.Status != nil {
		if len(*req.Status) > 20 {
			return "", fmt.Errorf("very long status")
		}
		if *req.Status != StatusActive && *req.Status != StatusResolved {
			return "", fmt.Errorf("invalid status")
		}
		status = *req.Status
	}
	return status, nil
}

func (s *Service) processingRadius(req *dto.RegistrationIncidentRequest) (int, error) {
	radius := s.config.DefaultRadius

	if req.RadiusInMeters != nil {
		if *req.RadiusInMeters > s.config.MaxRadius {
			return 0, fmt.Errorf("radius cannot be > %d", s.config.MaxRadius)
		}
		if *req.RadiusInMeters <= 0 {
			return 0, fmt.Errorf("radius cannot be <= 0")
		}
		radius = *req.RadiusInMeters
	}

	return radius, nil
}

func (s *Service) processingIsActive(status string) (bool, error) {
	var res bool
	if status == StatusActive {
		res = true
	} else if status == StatusResolved {
		res = false
	} else {
		return false, fmt.Errorf("unexpected status")
	}
	return res, nil
}

func (s *Service) processingResolvedTime(status string) *time.Time {
	var res *time.Time
	if status == StatusResolved {
		now := time.Now().UTC()
		res = &now
	}

	return res
}

func (s *Service) GetExistsIncidentByID(ctx context.Context, id string) (bool, error) {
	return s.db.GetExistByIncidentID(ctx, id, nil)
}

func (s *Service) GetIncidentInfoByID(ctx context.Context, id string) (*dto.IncidentAdminResponse, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	exist, err := s.db.GetExistByIncidentID(ctx, id, tx)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, fmt.Errorf("invalid incident_id")
	}

	res, err := s.db.GetInfoByIncidentID(ctx, id, tx)
	if err != nil {
		return nil, err
	}
	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	return dto.CreateAdminResponse(res), nil
}
