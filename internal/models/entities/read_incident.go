package entities

import "time"

type ReadIncident struct {
	Id           string
	Name         string
	Type         string
	Description  *string
	Latitude     string
	Longitude    string
	Radius       int
	IsActive     bool
	Status       string
	Coordinates  string
	CreatedDate  time.Time
	UpdatedDate  *time.Time
	ResolvedDate *time.Time
}
