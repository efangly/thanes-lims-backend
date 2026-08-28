package location

import (
	domainlocation "github.com/efangly/thanes-lims-backend/internal/domain/location"
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

type LocationResponse struct {
	ID          string  `json:"id"`
	ParentID    *string `json:"parent_id"`
	Name        string  `json:"name"`
	Kind        string  `json:"kind"`
	LevelType   string  `json:"level_type"`
	BarcodeCode string  `json:"barcode_code,omitempty"`
}

func toLocationResponse(l domainlocation.Location) LocationResponse {
	return LocationResponse{
		ID:          l.ID,
		ParentID:    l.ParentID,
		Name:        l.Name,
		Kind:        string(l.Kind),
		LevelType:   string(l.LevelType),
		BarcodeCode: l.BarcodeCode,
	}
}

type FullPathResponse struct {
	FullPath string `json:"full_path"`
}
