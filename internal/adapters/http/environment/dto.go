package environment

import (
	"time"

	applicationenvironment "github.com/efangly/thanes-lims-backend/internal/application/environment"
	"github.com/efangly/thanes-lims-backend/internal/domain/environment"
)

type RecordReadingRequest struct {
	Location string  `json:"location" validate:"required"`
	Value    float64 `json:"value"`
}

type GaugeResponse struct {
	Location string   `json:"location"`
	Unit     string   `json:"unit"`
	RangeMin float64  `json:"range_min"`
	RangeMax float64  `json:"range_max"`
	Value    *float64 `json:"value,omitempty"`
	Level    string   `json:"level"`
}

func toGaugeResponse(s applicationenvironment.GaugeStatus) GaugeResponse {
	resp := GaugeResponse{
		Location: s.Gauge.Location,
		Unit:     s.Gauge.Unit,
		RangeMin: s.Gauge.RangeMin,
		RangeMax: s.Gauge.RangeMax,
		Level:    string(s.Level),
	}
	if s.Reading != nil {
		v := s.Reading.Value
		resp.Value = &v
	}
	return resp
}

type ReadingResponse struct {
	Value      float64   `json:"value"`
	RecordedAt time.Time `json:"recorded_at"`
}

func toReadingResponse(r environment.SensorReading) ReadingResponse {
	return ReadingResponse{Value: r.Value, RecordedAt: r.RecordedAt}
}

type AlertResponse struct {
	ID          int64      `json:"id"`
	Location    string     `json:"location"`
	Level       string     `json:"level"`
	Title       string     `json:"title"`
	Message     string     `json:"message"`
	TriggeredAt time.Time  `json:"triggered_at"`
	ResolvedAt  *time.Time `json:"resolved_at,omitempty"`
}

func toAlertResponse(a environment.EnvAlert) AlertResponse {
	return AlertResponse{
		ID: a.ID, Location: a.Location, Level: string(a.Level),
		Title: a.Title, Message: a.Message, TriggeredAt: a.TriggeredAt, ResolvedAt: a.ResolvedAt,
	}
}
