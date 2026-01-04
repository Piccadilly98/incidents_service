package dto

import (
	"fmt"

	"github.com/Piccadilly98/incidents_service/internal/models/entities"
)

type UpdateRequest struct {
	Name        *string `json:"incident_name"`
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
		return fmt.Errorf("not data for update")
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
			return fmt.Errorf("radius canot be <= 0")
		}
	}
	if u.Status != nil {
		if *u.Status == "" {
			return fmt.Errorf("status cannot be empty")
		}
	}
	return nil
}

func (u *UpdateRequest) ToBaseEntity() *entities.UpdateIncident {
	return &entities.UpdateIncident{
		Name:        u.Name,
		Type:        u.Type,
		Description: u.Description,
		Radius:      u.Radius,
		Status:      u.Status,
	}
}
