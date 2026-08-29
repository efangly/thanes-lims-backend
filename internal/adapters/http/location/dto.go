package location

import (
	domainlocation "github.com/efangly/thanes-lims-backend/internal/domain/location"
	domainsample "github.com/efangly/thanes-lims-backend/internal/domain/sample"
)

type CreateCabinetRequest struct {
	Name string `json:"name" validate:"required"`
	// Kind selects the tree: "sample_storage" (default) or
	// "equipment_storage".
	Kind string `json:"kind" validate:"omitempty,oneof=sample_storage equipment_storage"`
}

type GenerateChildrenRequest struct {
	Prefix string `json:"prefix" validate:"required"`
	Count  int    `json:"count" validate:"required,min=1"`
}

// CreateBoxRequest creates a Box under :id (a Shelf/Slot/Sub-slot). rows are
// named A..Z (max 26), cols are two digits (max 99).
type CreateBoxRequest struct {
	Name string `json:"name" validate:"required"`
	Rows int    `json:"rows" validate:"required,min=1,max=26"`
	Cols int    `json:"cols" validate:"required,min=1,max=99"`
}

// EnlargeBoxRequest grows :id's Grid. Boxes only ever grow.
type EnlargeBoxRequest struct {
	Rows int `json:"rows" validate:"required,min=1,max=26"`
	Cols int `json:"cols" validate:"required,min=1,max=99"`
}

// MoveWithinBoxRequest rearranges Cells inside :id atomically.
type MoveWithinBoxRequest struct {
	Moves []MoveItem `json:"moves" validate:"required,min=1,dive"`
}

type MoveItem struct {
	SampleID string `json:"sample_id" validate:"required"`
	Position string `json:"position" validate:"required"`
}

type LocationResponse struct {
	ID          string  `json:"id"`
	ParentID    *string `json:"parent_id"`
	Name        string  `json:"name"`
	Kind        string  `json:"kind"`
	LevelType   string  `json:"level_type"`
	BarcodeCode string  `json:"barcode_code,omitempty"`
	// Rows/Cols are the Box Grid; omitted for non-Box nodes.
	Rows int `json:"rows,omitempty"`
	Cols int `json:"cols,omitempty"`
}

func toLocationResponse(l domainlocation.Location) LocationResponse {
	return LocationResponse{
		ID:          l.ID,
		ParentID:    l.ParentID,
		Name:        l.Name,
		Kind:        string(l.Kind),
		LevelType:   string(l.LevelType),
		BarcodeCode: l.BarcodeCode,
		Rows:        l.Rows,
		Cols:        l.Cols,
	}
}

type FullPathResponse struct {
	FullPath string `json:"full_path"`
}

// BoxCellResponse is the post-move Cell layout of a Box - one row per active
// Sample it holds.
type BoxCellResponse struct {
	SampleID string `json:"sample_id"`
	Position string `json:"position"`
}

func toMoveResponse(samples []domainsample.Sample) []BoxCellResponse {
	out := make([]BoxCellResponse, len(samples))
	for i, s := range samples {
		pos := ""
		if s.Position != nil {
			pos = *s.Position
		}
		out[i] = BoxCellResponse{SampleID: s.ID, Position: pos}
	}
	return out
}
