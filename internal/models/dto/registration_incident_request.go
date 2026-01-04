package dto

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Piccadilly98/incidents_service/internal/models/entities"
)

const (
	maxLatitude  = 90.0
	minLatitude  = -90.0
	maxLongitude = 180.0
	minLongitude = -180.0

	maxLatitudeDecimals = 8
	maxLenLatitude      = 12

	maxLongitudeDecimals = 8
	maxLenLongitude      = 13
)

type RegistrationIncidentRequest struct {
	Name           string  `json:"incident_name"`
	Type           string  `json:"type"`
	Latitude       string  `json:"latitude"`
	Longitude      string  `json:"longitude"`
	Description    *string `json:"description"`
	RadiusInMeters *int    `json:"radius"`
	Status         *string `json:"status"`
}

func (r *RegistrationIncidentRequest) Validate() error {
	if r.Name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if r.Type == "" {
		return fmt.Errorf("type cannot be empty")
	}
	if r.Latitude == "" {
		return fmt.Errorf("Latitude cannot be empty")
	}
	if r.Longitude == "" {
		return fmt.Errorf("Longitude cannot be empty")
	}
	if r.Description != nil && *r.Description == "" {
		return fmt.Errorf("desctiption cannot be empty")
	}
	if r.RadiusInMeters != nil && *r.RadiusInMeters <= 0 {
		return fmt.Errorf("radius canot be <= 0")
	}
	if r.Status != nil && *r.Status == "" {
		return fmt.Errorf("status cannot be empty")
	}
	err := r.ValidateCoordinates()
	if err != nil {
		return err
	}
	return nil
}

func (r *RegistrationIncidentRequest) ValidateCoordinates() error {
	if len(r.Latitude) > maxLenLatitude {
		return fmt.Errorf("latitude incorrect")
	}
	if len(r.Longitude) > maxLenLongitude {
		return fmt.Errorf("longitude incorrect")
	}
	latStr := strings.Replace(r.Latitude, ",", ".", 1)
	lonStr := strings.Replace(r.Longitude, ",", ".", 1)
	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		return fmt.Errorf("latitude incorrect parse")
	}
	long, err := strconv.ParseFloat(lonStr, 64)
	if err != nil {
		return fmt.Errorf("longitude incorrect parse")
	}

	if lat > maxLatitude || lat < minLatitude {
		return fmt.Errorf("latitude incorrect compare")
	}
	if long > maxLongitude || long < minLongitude {
		return fmt.Errorf("longitude incorrect compare")
	}

	latArr := strings.Split(latStr, ".")
	longArr := strings.Split(lonStr, ".")
	if len(latArr) == 2 {
		if len(latArr[1]) > maxLatitudeDecimals {
			return fmt.Errorf("latitide: incorrect format")
		}
	}
	if len(longArr) == 2 {
		if len(longArr[1]) > maxLongitudeDecimals {
			return fmt.Errorf("longitude: invalid format")
		}
	}
	return nil
}

func (r *RegistrationIncidentRequest) ToBaseEntity() *entities.RegistrationIncidentEntitie {
	return &entities.RegistrationIncidentEntitie{
		Name:        r.Name,
		Type:        r.Type,
		Latitude:    r.Latitude,
		Longitude:   r.Longitude,
		Description: r.Description,
	}
}
