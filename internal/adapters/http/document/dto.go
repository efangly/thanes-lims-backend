package document

import (
	"path"
	"time"

	"github.com/efangly/thanes-lims-backend/internal/domain/document"
)

type DocumentResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	Version     string    `json:"version"`
	CreatedBy   string    `json:"created_by"`
	IssuedAt    time.Time `json:"issued_at"`
	AccessLevel string    `json:"access_level"`
	Locked      bool      `json:"locked"`
	// Filename is the stored file's base name with extension, derived from
	// StorageKey. The frontend uses the extension to decide whether a file
	// kind can be previewed inline (ADR-0013).
	Filename           string  `json:"filename"`
	EquipmentID        *string `json:"equipment_id"`
	CalibrationEventID *int64  `json:"calibration_event_id"`
}

func toResponse(d document.Document) DocumentResponse {
	return DocumentResponse{
		ID:                 d.ID,
		Name:               d.Name,
		Type:               string(d.Type),
		Version:            d.Version,
		CreatedBy:          d.CreatedBy,
		IssuedAt:           d.IssuedAt,
		AccessLevel:        d.AccessLevel,
		Locked:             d.Locked,
		Filename:           path.Base(d.StorageKey),
		EquipmentID:        d.EquipmentID,
		CalibrationEventID: d.CalibrationEventID,
	}
}

type HistoryResponse struct {
	Version string    `json:"version"`
	Change  string    `json:"change"`
	Date    time.Time `json:"date"`
	Who     string    `json:"who"`
}

func toHistoryResponse(h document.DocHistory) HistoryResponse {
	return HistoryResponse{Version: h.Version, Change: h.Change, Date: h.Date, Who: h.Who}
}

type LockRequest struct {
	Locked bool `json:"locked"`
}

// UpdateDocumentRequest is a partial update: a nil field is left
// unchanged. For the optional links, an explicit empty string / 0 clears
// the link.
type UpdateDocumentRequest struct {
	Name               *string `json:"name"`
	Type               *string `json:"type"`
	AccessLevel        *string `json:"access_level"`
	EquipmentID        *string `json:"equipment_id"`
	CalibrationEventID *int64  `json:"calibration_event_id"`
	ChangeNote         string  `json:"change_note"`
}
