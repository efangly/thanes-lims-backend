package document

import (
	"context"
	"time"

	"github.com/efangly/thanes-lims-backend/internal/domain/document"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	portdocument "github.com/efangly/thanes-lims-backend/internal/ports/document"
)

// DeleteDocumentUseCase soft-deletes a Document (Retired, per ADR 0003).
// The row, its doc_history and the stored file all stay in place so the
// Document can be restored and the audit trail never gets a gap.
type DeleteDocumentUseCase struct {
	documents portdocument.Repository
	history   portdocument.HistoryRepository
}

func NewDeleteDocumentUseCase(documents portdocument.Repository, history portdocument.HistoryRepository) *DeleteDocumentUseCase {
	return &DeleteDocumentUseCase{documents: documents, history: history}
}

// Execute retires the Document. A locked Document cannot be deleted - the
// same rule CreateNewVersionUseCase enforces for edits.
func (uc *DeleteDocumentUseCase) Execute(ctx context.Context, id, actor string) error {
	d, err := uc.documents.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if d.Locked {
		return shared.ErrForbidden
	}

	if err := uc.documents.Delete(ctx, id); err != nil {
		return err
	}

	_, err = uc.history.Append(ctx, document.DocHistory{
		DocumentID: d.ID,
		Version:    d.Version,
		Change:     "ลบเอกสารออกจากระบบ (Retired)",
		Date:       time.Now(),
		Who:        actor,
	})
	return err
}
