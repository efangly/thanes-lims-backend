package partnerdevice

import (
	"time"

	"github.com/efangly/thanes-lims-backend/internal/domain/environment"
	portenvironment "github.com/efangly/thanes-lims-backend/internal/ports/environment"
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

// DiscoverDeviceResponse is one SMtrack device as returned by browsing a
// ward, before it's necessarily mapped to a Location here. The reading
// fields are pointers (not zero values) so "no telemetry yet" is
// distinguishable from a genuine 0.0 reading.
type DiscoverDeviceResponse struct {
	Serial          string     `json:"serial"`
	Name            string     `json:"name"`
	Status          bool       `json:"status"`
	Firmware        string     `json:"firmware"`
	Online          bool       `json:"online"`
	TempDisplay     *float64   `json:"temp_display,omitempty"`
	HumidityDisplay *float64   `json:"humidity_display,omitempty"`
	SendTime        *time.Time `json:"send_time,omitempty"`
}

type DiscoverDevicesResponse struct {
	Devices []DiscoverDeviceResponse `json:"devices"`
	Total   int                      `json:"total"`
	Page    int                      `json:"page"`
	Limit   int                      `json:"limit"`
}

func toDiscoverResponse(l portenvironment.PartnerDeviceListing) DiscoverDevicesResponse {
	out := DiscoverDevicesResponse{
		Devices: make([]DiscoverDeviceResponse, len(l.Devices)),
		Total:   l.Total,
		Page:    l.Page,
		Limit:   l.Limit,
	}
	for i, d := range l.Devices {
		item := DiscoverDeviceResponse{
			Serial:   d.Device.Serial,
			Name:     d.Device.Name,
			Status:   d.Device.Status,
			Firmware: d.Device.Firmware,
			Online:   d.Device.Online,
		}
		if d.HasLatest {
			item.TempDisplay = &d.Latest.TempDisplay
			item.HumidityDisplay = &d.Latest.HumidityDisplay
			item.SendTime = &d.Latest.SendTime
		}
		out.Devices[i] = item
	}
	return out
}
