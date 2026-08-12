package environment

import "time"

type SensorReading struct {
	ID         int64
	Location   string
	Value      float64
	RecordedAt time.Time
}
