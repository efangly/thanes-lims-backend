package partnerdevice

import (
	"time"

	"github.com/efangly/thanes-lims-backend/internal/domain/environment"
)

type CreatePartnerDeviceRequest struct {
	Serial   string `json:"serial" validate:"required"`
	Location string `json:"location" validate:"required"`
	Active   bool   `json:"active"`
}

type UpdatePartnerDeviceRequest struct {
	Location string `json:"location" validate:"required"`
	Active   bool   `json:"active"`
}

type PartnerDeviceResponse struct {
	Serial   string `json:"serial"`
	Location string `json:"location"`
	Active   bool   `json:"active"`
}

func toResponse(d environment.PartnerDevice) PartnerDeviceResponse {
	return PartnerDeviceResponse{Serial: d.Serial, Location: d.Location, Active: d.Active}
}

// SnapshotResponse is the polled metadata + latest reading + derived alert
// Level for one Partner Device, as served by both GET .../snapshot and the
// SSE stream.
type SnapshotResponse struct {
	Serial          string    `json:"serial"`
	Location        string    `json:"location"`
	Name            string    `json:"name"`
	Status          bool      `json:"status"`
	Firmware        string    `json:"firmware"`
	Online          bool      `json:"online"`
	TempDisplay     float64   `json:"temp_display"`
	HumidityDisplay float64   `json:"humidity_display"`
	SendTime        time.Time `json:"send_time"`
	Level           string    `json:"level"`
	FetchedAt       time.Time `json:"fetched_at"`
	Stale           bool      `json:"stale"`
}

func toSnapshotResponse(s environment.PartnerDeviceSnapshot) SnapshotResponse {
	return SnapshotResponse{
		Serial:          s.Serial,
		Location:        s.Location,
		Name:            s.Name,
		Status:          s.Status,
		Firmware:        s.Firmware,
		Online:          s.Online,
		TempDisplay:     s.TempDisplay,
		HumidityDisplay: s.HumidityDisplay,
		SendTime:        s.SendTime,
		Level:           string(s.Level),
		FetchedAt:       s.FetchedAt,
		Stale:           s.Stale,
	}
}
