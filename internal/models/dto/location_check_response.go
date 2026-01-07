package dto

type LocationCheckResponse struct {
	ID                  string                  `json:"check_id"`
	UserID              string                  `json:"user_id"`
	Latitude            string                  `json:"latitude"`
	Longitude           string                  `json:"longitude"`
	IsDanger            bool                    `json:"is_danger"`
	DetectedIncidentsID []*IncidentUserResponse `json:"detected_incidents"`
}
