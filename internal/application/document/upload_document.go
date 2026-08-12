package document

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/efangly/thanes-lims-backend/internal/domain/document"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	portdocument "github.com/efangly/thanes-lims-backend/internal/ports/document"
	portfilestorage "github.com/efangly/thanes-lims-backend/internal/ports/filestorage"
	portidgen "github.com/efangly/thanes-lims-backend/internal/ports/idgen"
)

type UploadDocumentUseCase struct {
	documents portdocument.Repository
	history   portdocument.HistoryRepository
	storage   portfilestorage.FileStorage
	idgen     portidgen.SequenceGenerator
}

func NewUploadDocumentUseCase(documents portdocument.Repository, history portdocument.HistoryRepository, storage portfilestorage.FileStorage, idgen portidgen.SequenceGenerator) *UploadDocumentUseCase {
	return &UploadDocumentUseCase{documents: documents, history: history, storage: storage, idgen: idgen}
}

type UploadDocumentInput struct {
	Name        string
	Type        document.Type
	Filename    string
	ContentType string
	Size        int64
	Content     io.Reader
	AccessLevel string
	UploadedBy  string
}

// Execute stores the object under a version-scoped key (docs/{id}/{version}/{filename})
// rather than overwriting - every version stays retrievable for audit integrity.
func (uc *UploadDocumentUseCase) Execute(ctx context.Context, in UploadDocumentInput) (document.Document, error) {
	if !in.Type.Valid() {
		return document.Document{}, shared.ErrValidation
	}

	seq, err := uc.idgen.Next(ctx, "document", nil)
	if err != nil {
		return document.Document{}, err
	}

	id := fmt.Sprintf("DOC-%05d", seq)
	const initialVersion = "1"
	key := fmt.Sprintf("docs/%s/%s/%s", id, initialVersion, in.Filename)

	if err := uc.storage.Upload(ctx, key, in.Content, in.Size, in.ContentType); err != nil {
		return document.Document{}, err
	}

	now := time.Now()
	created, err := uc.documents.Create(ctx, document.Document{
		ID:          id,
		Name:        in.Name,
		Type:        in.Type,
		Version:     initialVersion,
		CreatedBy:   in.UploadedBy,
		IssuedAt:    now,
		AccessLevel: in.AccessLevel,
		StorageKey:  key,
	})
	if err != nil {
		return document.Document{}, err
	}

	_, err = uc.history.Append(ctx, document.DocHistory{
		DocumentID: created.ID,
		Version:    initialVersion,
		Change:     "อัปโหลดเอกสารเข้าสู่ระบบ",
		Date:       now,
		Who:        in.UploadedBy,
	})
	if err != nil {
		return document.Document{}, err
	}

	return created, nil
}
