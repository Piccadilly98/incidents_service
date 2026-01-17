package dto

import (
	"fmt"
	"time"

	"github.com/Piccadilly98/incidents_service/internal/models/entities"
)

type UpdateRequest struct {
	Name        *string `json:"name"`
	Type        *string `json:"type"`
	Description *string `json:"description"`
	Radius      *int    `json:"radius"`
	Status      *string `json:"status"`
}

func (u *UpdateRequest) Validate() error {
	if !(u.Name != nil || u.Type != nil ||
		u.Description != nil ||
		u.Radius != nil ||
		u.Status != nil) {
		return fmt.Errorf("no data for update")
	}
	if u.Name != nil {
		if *u.Name == "" {
			return fmt.Errorf("name cannot be empty")
		}
	}
	if u.Type != nil {
		if *u.Type == "" {
			return fmt.Errorf("type cannot be empty")
		}
	}
	if u.Description != nil {
		if *u.Description == "" {
			return fmt.Errorf("description cannot be empty")
		}
	}
	if u.Radius != nil {
		if *u.Radius <= 0 {
			return fmt.Errorf("radius cannot be <= 0")
		}
	}
	if u.Status != nil {
		if *u.Status == "" {
			return fmt.Errorf("status cannot be empty")
		}
	}
	return nil
}

func (u *UpdateRequest) ToEntity(resolvedTime *time.Time, isActive bool) *entities.UpdateIncident {
	return &entities.UpdateIncident{
		Name:         u.Name,
		Type:         u.Type,
		ResolvedTime: resolvedTime,
		IsActive:     isActive,
		Description:  u.Description,
		Radius:       u.Radius,
		Status:       u.Status,
	}
}
