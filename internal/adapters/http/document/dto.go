package document

import (
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
}

func toResponse(d document.Document) DocumentResponse {
	return DocumentResponse{
		ID:          d.ID,
		Name:        d.Name,
		Type:        string(d.Type),
		Version:     d.Version,
		CreatedBy:   d.CreatedBy,
		IssuedAt:    d.IssuedAt,
		AccessLevel: d.AccessLevel,
		Locked:      d.Locked,
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
