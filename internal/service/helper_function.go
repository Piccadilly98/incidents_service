package service

import "time"

func getPtrTime(t time.Time) *time.Time {
	return &t
}
