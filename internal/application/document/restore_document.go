package document

import (
	"context"
	"time"

	"github.com/efangly/thanes-lims-backend/internal/domain/document"
	portdocument "github.com/efangly/thanes-lims-backend/internal/ports/document"
)

// RestoreDocumentUseCase un-retires a soft-deleted Document.
type RestoreDocumentUseCase struct {
	documents portdocument.Repository
	history   portdocument.HistoryRepository
}

func NewRestoreDocumentUseCase(documents portdocument.Repository, history portdocument.HistoryRepository) *RestoreDocumentUseCase {
	return &RestoreDocumentUseCase{documents: documents, history: history}
}

// Execute clears deleted_at. The repository returns shared.ErrNotFound
// when the id is unknown and shared.ErrConflict when the Document is not
// currently Retired.
func (uc *RestoreDocumentUseCase) Execute(ctx context.Context, id, actor string) (document.Document, error) {
	if err := uc.documents.Restore(ctx, id); err != nil {
		return document.Document{}, err
	}

	restored, err := uc.documents.FindByID(ctx, id)
	if err != nil {
		return document.Document{}, err
	}

	if _, err := uc.history.Append(ctx, document.DocHistory{
		DocumentID: restored.ID,
		Version:    restored.Version,
		Change:     "กู้คืนเอกสาร",
		Date:       time.Now(),
		Who:        actor,
	}); err != nil {
		return document.Document{}, err
	}

	return restored, nil
}
