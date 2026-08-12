package environment

import "time"

type GaugeModel struct {
	Location string `gorm:"primaryKey"`
	Unit     string
	RangeMin float64
	RangeMax float64
}

func (GaugeModel) TableName() string { return "gauges" }

type ReadingModel struct {
	ID         int64 `gorm:"primaryKey"`
	Location   string
	Value      float64
	RecordedAt time.Time
}

func (ReadingModel) TableName() string { return "sensor_readings" }

type AlertModel struct {
	ID          int64 `gorm:"primaryKey"`
	Location    string
	Level       string
	Title       string
	Message     string
	TriggeredAt time.Time
	ResolvedAt  *time.Time
}

func (AlertModel) TableName() string { return "env_alerts" }
