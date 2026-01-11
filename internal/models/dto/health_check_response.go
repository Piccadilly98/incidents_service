package dto

type HealthCheckResponse struct {
	ServerStatus string   `json:"server_status"`
	Errors       []string `json:"errors,omitempty"`
}
