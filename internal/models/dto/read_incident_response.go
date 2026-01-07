package dto

import (
	"time"

	"github.com/Piccadilly98/incidents_service/internal/models/entities"
)

type IncidentBaseResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type IncidentUserResponse struct {
	IncidentBaseResponse
	Latitude       string   `json:"latitude"`
	Longitude      string   `json:"longitude"`
	Radius         int      `json:"radius"`
	IsActive       bool     `json:"is_active"`
	DistanceMeters *float64 `json:"distance_meters,omitempty"`
}

type IncidentAdminResponse struct {
	IncidentUserResponse
	Description  *string    `json:"description"`
	UpdatedDate  *time.Time `json:"updated_date"`
	ResolvedDate *time.Time `json:"resolved_date"`
	CreatedDate  time.Time  `json:"created_date"`
	Status       string     `json:"status"`
}

func CreateUserResponse(entittie *entities.ReadIncident, distanceMeters *float64) *IncidentUserResponse {
	res := &IncidentUserResponse{
		IncidentBaseResponse: IncidentBaseResponse{
			ID:   entittie.Id,
			Type: entittie.Type,
			Name: entittie.Name,
		},
		Latitude:       entittie.Latitude,
		Longitude:      entittie.Longitude,
		Radius:         entittie.Radius,
		IsActive:       entittie.IsActive,
		DistanceMeters: distanceMeters,
	}
	return res
}
func CreateAdminResponse(entittie *entities.ReadIncident, distanceMeters *float64) *IncidentAdminResponse {
	res := &IncidentAdminResponse{
		IncidentUserResponse: *CreateUserResponse(entittie, distanceMeters),
		Description:          entittie.Description,
		UpdatedDate:          entittie.UpdatedDate,
		ResolvedDate:         entittie.ResolvedDate,
		CreatedDate:          entittie.CreatedDate,
		Status:               entittie.Status,
	}
	return res
}
