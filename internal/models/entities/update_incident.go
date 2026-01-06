package entities

import "time"

type UpdateIncident struct {
	Name         *string
	Type         *string
	Description  *string
	Radius       *int
	Status       *string
	IsActive     bool
	ResolvedTime *time.Time
}
