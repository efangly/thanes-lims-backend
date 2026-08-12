package shared

import "time"

// BuddhistYear converts a Gregorian time to its Thai Buddhist Era year (CE + 543).
func BuddhistYear(t time.Time) int {
	return t.Year() + 543
}
