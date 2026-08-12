package environment

import "time"

type EnvAlert struct {
	ID          int64
	Location    string
	Level       Level
	Title       string
	Message     string
	TriggeredAt time.Time
	ResolvedAt  *time.Time
}
