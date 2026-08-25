package audit

import (
	"time"

	domainaudit "github.com/efangly/thanes-lims-backend/internal/domain/audit"
)

// EntryResponse is one audit trail entry as returned by GET /audit/logs.
type EntryResponse struct {
	ID         int64          `json:"id"`
	ActorID    *int64         `json:"actor_id,omitempty"`
	ActorRole  string         `json:"actor_role"`
	Method     string         `json:"method"`
	Path       string         `json:"path"`
	Resource   string         `json:"resource"`
	ResourceID string         `json:"resource_id"`
	StatusCode int            `json:"status_code"`
	Metadata   map[string]any `json:"metadata,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
}

func toEntryResponse(e domainaudit.AuditLog) EntryResponse {
	return EntryResponse{
		ID:         e.ID,
		ActorID:    e.ActorID,
		ActorRole:  e.ActorRole,
		Method:     e.Method,
		Path:       e.Path,
		Resource:   e.Resource,
		ResourceID: e.ResourceID,
		StatusCode: e.StatusCode,
		Metadata:   e.Metadata,
		CreatedAt:  e.CreatedAt,
	}
}

// PageMeta is the pagination side-channel returned in the response
// envelope's `meta` field.
type PageMeta struct {
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
	Total int64 `json:"total"`
}
