package document

import (
	"context"

	"github.com/efangly/thanes-lims-backend/internal/domain/document"
)

type Repository interface {
	Create(ctx context.Context, d document.Document) (document.Document, error)
	FindByID(ctx context.Context, id string) (document.Document, error)
	List(ctx context.Context) ([]document.Document, error)
	Update(ctx context.Context, d document.Document) (document.Document, error)
}

type HistoryRepository interface {
	Append(ctx context.Context, h document.DocHistory) (document.DocHistory, error)
	ListByDocument(ctx context.Context, documentID string) ([]document.DocHistory, error)
}
