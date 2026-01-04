package entities

type UpdateIncident struct {
	Name        *string
	Type        *string
	Description *string
	Radius      *int
	Status      *string
	IsEnded     bool
}
