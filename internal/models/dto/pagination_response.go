package dto

type PaginationResponse struct {
	Incidents      []*IncidentAdminResponse `json:"incidents"`
	CountIncidents int                      `json:"incidents_count"`
}

func ToPaginationResponse(incidents []*IncidentAdminResponse) *PaginationResponse {
	return &PaginationResponse{
		Incidents:      incidents,
		CountIncidents: len(incidents),
	}
}
