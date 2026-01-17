package dto

import (
	"fmt"
	"math"
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
	Name           string  `json:"name"`
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
		return fmt.Errorf("Description cannot be empty")
	}
	if r.RadiusInMeters != nil && *r.RadiusInMeters <= 0 {
		return fmt.Errorf("radius cannot be <= 0")
	}
	if r.Status != nil && *r.Status == "" {
		return fmt.Errorf("status cannot be empty")
	}
	err := ValidateCoordinates(r.Latitude, r.Longitude)
	if err != nil {
		return err
	}
	return nil
}

func ValidateCoordinates(latitude, longitude string) error {
	if len(latitude) > maxLenLatitude {
		return fmt.Errorf("latitude incorrect")
	}
	if len(longitude) > maxLenLongitude {
		return fmt.Errorf("longitude incorrect")
	}
	latStr := strings.Replace(latitude, ",", ".", 1)
	lonStr := strings.Replace(longitude, ",", ".", 1)
	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		return fmt.Errorf("latitude incorrect parse")
	}
	if math.IsNaN(lat) {
		return fmt.Errorf("latitude incorrect parse")
	}
	long, err := strconv.ParseFloat(lonStr, 64)
	if err != nil {
		return fmt.Errorf("longitude incorrect parse")
	}
	if math.IsNaN(long) {
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
			return fmt.Errorf("longitude: incorrect format")
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
