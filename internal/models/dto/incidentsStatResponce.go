package dto

import (
	"time"

	"github.com/Piccadilly98/incidents_service/internal/models/entities"
)

type IncidentsStatResponse struct {
	TotalUniqueUser int            `json:"total_unique_user"`
	TimeStatWindow  int            `json:"time_stat_window"`
	FromDate        time.Time      `json:"from_date"`
	ToDate          time.Time      `json:"to_date"`
	TotalIncidents  int            `json:"total_incidents"`
	IncidentsStat   []IncidentStat `json:"incidents_stat"`
}

type IncidentStat struct {
	IncidentBaseResponse
	UserCount int `json:"user_count"`
}

func ToIncidentsStatResponse(entities []*entities.IncidentStat, timeWindow int, timeRequest time.Time, totalUniqueUsers int) *IncidentsStatResponse {
	res := &IncidentsStatResponse{
		TotalUniqueUser: totalUniqueUsers,
		TimeStatWindow:  timeWindow,
		FromDate:        timeRequest.UTC().Add(-(time.Duration(timeWindow) * time.Second)),
		ToDate:          timeRequest.UTC(),
		TotalIncidents:  len(entities),
	}

	for _, entitie := range entities {
		res.IncidentsStat = append(
			res.IncidentsStat,
			IncidentStat{
				IncidentBaseResponse: IncidentBaseResponse{
					ID:   entitie.ID,
					Name: entitie.Name,
					Type: entitie.Type,
				},
				UserCount: entitie.UserCount,
			})
	}

	return res
}
