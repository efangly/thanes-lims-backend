package document

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/efangly/thanes-lims-backend/internal/domain/document"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	portdocument "github.com/efangly/thanes-lims-backend/internal/ports/document"
	portfilestorage "github.com/efangly/thanes-lims-backend/internal/ports/filestorage"
)

type CreateNewVersionUseCase struct {
	documents portdocument.Repository
	history   portdocument.HistoryRepository
	storage   portfilestorage.FileStorage
}

func NewCreateNewVersionUseCase(documents portdocument.Repository, history portdocument.HistoryRepository, storage portfilestorage.FileStorage) *CreateNewVersionUseCase {
	return &CreateNewVersionUseCase{documents: documents, history: history, storage: storage}
}

type CreateNewVersionInput struct {
	DocumentID  string
	Filename    string
	ContentType string
	Size        int64
	Content     io.Reader
	ChangeNote  string
	UploadedBy  string
}

func (uc *CreateNewVersionUseCase) Execute(ctx context.Context, in CreateNewVersionInput) (document.Document, error) {
	d, err := uc.documents.FindByID(ctx, in.DocumentID)
	if err != nil {
		return document.Document{}, err
	}
	if d.Locked {
		return document.Document{}, shared.ErrForbidden
	}

	nextVersion := nextVersionString(d.Version)
	key := fmt.Sprintf("docs/%s/%s/%s", d.ID, nextVersion, in.Filename)

	if err := uc.storage.Upload(ctx, key, in.Content, in.Size, in.ContentType); err != nil {
		return document.Document{}, err
	}

	d.Version = nextVersion
	d.StorageKey = key
	updated, err := uc.documents.Update(ctx, d)
	if err != nil {
		return document.Document{}, err
	}

	_, err = uc.history.Append(ctx, document.DocHistory{
		DocumentID: d.ID,
		Version:    nextVersion,
		Change:     in.ChangeNote,
		Date:       time.Now(),
		Who:        in.UploadedBy,
	})
	if err != nil {
		return document.Document{}, err
	}

	return updated, nil
}

func nextVersionString(current string) string {
	n, err := strconv.Atoi(current)
	if err != nil {
		return "2"
	}
	return strconv.Itoa(n + 1)
}
