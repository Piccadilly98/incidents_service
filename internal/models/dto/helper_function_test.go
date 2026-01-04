package dto_test

func getIntPtr(i int) *int {
	return &i
}

func getBoolPtr(b bool) *bool {
	return &b
}

func getPtrStr(str string) *string {
	return &str
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr ||
			contains(s[1:], substr))))
}
