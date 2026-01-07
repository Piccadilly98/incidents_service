package dto

type PaginationResponse struct {
	Incidents      []*IncidentAdminResponse `json:"incidents"`
	CountIncidents int                      `json:"incidents_count"`
	TotalPages     int                      `json:"total_pages"`
	PageNum        *int                     `json:"page_num,omitempty"`
	TotalIncidents int                      `json:"total_incidents"`
}

func ToPaginationResponse(incidents []*IncidentAdminResponse, totalPages, totalIncidents int, pageNum *int) *PaginationResponse {
	return &PaginationResponse{
		Incidents:      incidents,
		CountIncidents: len(incidents),
		TotalPages:     totalPages,
		TotalIncidents: totalIncidents,
		PageNum:        pageNum,
	}
}
