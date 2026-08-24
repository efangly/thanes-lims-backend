package sample

import (
	"time"

	"github.com/efangly/thanes-lims-backend/internal/domain/sample"
)

type CreateSampleRequest struct {
	Name       string  `json:"name" validate:"required"`
	Type       string  `json:"type" validate:"required"`
	Custodian  string  `json:"custodian" validate:"required"`
	LocationID *string `json:"location_id"`
}

type UpdateStatusRequest struct {
	Status string `json:"status" validate:"required"`
}

type AssignLocationRequest struct {
	LocationID string `json:"location_id" validate:"required"`
}

type AppendCoCStepRequest struct {
	Title string `json:"title" validate:"required"`
	Meta  string `json:"meta"`
}

type SampleResponse struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Type       string    `json:"type"`
	Custodian  string    `json:"custodian"`
	LocationID *string   `json:"location_id"`
	Status     string    `json:"status"`
	ReceivedAt time.Time `json:"received_at"`
}

func toSampleResponse(s sample.Sample) SampleResponse {
	return SampleResponse{
		ID:         s.ID,
		Name:       s.Name,
		Type:       string(s.Type),
		Custodian:  s.Custodian,
		LocationID: s.LocationID,
		Status:     string(s.Status),
		ReceivedAt: s.ReceivedAt,
	}
}

type CoCStepResponse struct {
	ID         int64     `json:"id"`
	State      string    `json:"state"`
	Icon       string    `json:"icon"`
	Title      string    `json:"title"`
	Meta       string    `json:"meta"`
	Who        string    `json:"who"`
	OccurredAt time.Time `json:"occurred_at"`
}

func toCoCStepResponse(s sample.CoCStep) CoCStepResponse {
	return CoCStepResponse{
		ID:         s.ID,
		State:      string(s.State),
		Icon:       string(s.Icon),
		Title:      s.Title,
		Meta:       s.Meta,
		Who:        s.Who,
		OccurredAt: s.OccurredAt,
	}
}
