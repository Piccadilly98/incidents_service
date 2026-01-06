package entities

type PaginationIncidents struct {
	Offset int
	Limit  int
	Status string
	Name   string
	Type   string
	Radius *int
	ID     string
}
