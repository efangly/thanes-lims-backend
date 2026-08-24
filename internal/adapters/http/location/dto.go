package location

import (
	domainlocation "github.com/efangly/thanes-lims-backend/internal/domain/location"
)

type CreateCabinetRequest struct {
	Name string `json:"name" validate:"required"`
}

type GenerateChildrenRequest struct {
	Prefix string `json:"prefix" validate:"required"`
	Count  int    `json:"count" validate:"required,min=1"`
}

type LocationResponse struct {
	ID        string  `json:"id"`
	ParentID  *string `json:"parent_id"`
	Name      string  `json:"name"`
	LevelType string  `json:"level_type"`
}

func toLocationResponse(l domainlocation.Location) LocationResponse {
	return LocationResponse{
		ID:        l.ID,
		ParentID:  l.ParentID,
		Name:      l.Name,
		LevelType: string(l.LevelType),
	}
}

type FullPathResponse struct {
	FullPath string `json:"full_path"`
}
