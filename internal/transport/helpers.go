package transport

import "time"

// validateDateParam returns the date string if it's a valid YYYY-MM-DD, otherwise "".
func validateDateParam(s string) string {
	if s == "" {
		return ""
	}
	_, err := time.Parse("2006-01-02", s)
	if err != nil {
		return ""
	}
	return s
}
