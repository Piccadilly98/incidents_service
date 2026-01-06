package service

import "time"

func getPtrTime(t time.Time) *time.Time {
	return &t
}
func getIntPtr(i int) *int {
	return &i
}

func getPtrStr(str string) *string {
	return &str
}
func notNil() *time.Time {
	return &time.Time{}
}

func ptrStringEqual(a, b *string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func ptrIntEqual(a, b *int) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}
