package dto

import "fmt"

type LocationCheckRequest struct {
	UserID    string `json:"user_id"`
	Latitude  string `json:"latitude"`
	Longitude string `json:"longitude"`
}

func (l *LocationCheckRequest) Validate() error {
	if l.UserID == "" {
		return fmt.Errorf("user_id cannot be empty")
	}
	return ValidateCoordinates(l.Latitude, l.Longitude)
}
