package entities

import "time"

type RegistrationIncidentEntitie struct {
	Name         string
	Type         string
	Description  *string
	Latitude     string
	Longitude    string
	Radius       int
	IsActive     bool
	Status       string
	ResolvedTime *time.Time
}
