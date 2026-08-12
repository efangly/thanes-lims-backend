package document_test

import (
	"bytes"
	"context"
	"testing"

	applicationdocument "github.com/efangly/thanes-lims-backend/internal/application/document"
	"github.com/efangly/thanes-lims-backend/internal/domain/document"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateNewVersionUseCase_LockedDocumentRejected(t *testing.T) {
	docs := new(mockDocRepo)
	history := new(mockHistoryRepo)
	storage := new(mockFileStorage)

	docs.On("FindByID", mock.Anything, "DOC-00001").Return(document.Document{ID: "DOC-00001", Locked: true}, nil)

	uc := applicationdocument.NewCreateNewVersionUseCase(docs, history, storage)
	_, err := uc.Execute(context.Background(), applicationdocument.CreateNewVersionInput{
		DocumentID: "DOC-00001", Filename: "sop-v2.pdf", ContentType: "application/pdf", Size: 10, Content: bytes.NewReader([]byte("y")),
	})

	assert.ErrorIs(t, err, shared.ErrForbidden)
	storage.AssertNotCalled(t, "Upload", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestCreateNewVersionUseCase_BumpsVersion(t *testing.T) {
	docs := new(mockDocRepo)
	history := new(mockHistoryRepo)
	storage := new(mockFileStorage)

	docs.On("FindByID", mock.Anything, "DOC-00001").Return(document.Document{ID: "DOC-00001", Version: "1", Locked: false}, nil)
	storage.On("Upload", mock.Anything, "docs/DOC-00001/2/sop-v2.pdf", mock.Anything, int64(10), "application/pdf").Return(nil)
	docs.On("Update", mock.Anything, mock.MatchedBy(func(d document.Document) bool {
		return d.Version == "2" && d.StorageKey == "docs/DOC-00001/2/sop-v2.pdf"
	})).Return(document.Document{ID: "DOC-00001", Version: "2"}, nil)
	history.On("Append", mock.Anything, mock.AnythingOfType("document.DocHistory")).Return(document.DocHistory{}, nil)

	uc := applicationdocument.NewCreateNewVersionUseCase(docs, history, storage)
	d, err := uc.Execute(context.Background(), applicationdocument.CreateNewVersionInput{
		DocumentID: "DOC-00001", Filename: "sop-v2.pdf", ContentType: "application/pdf", Size: 10, Content: bytes.NewReader([]byte("y")), UploadedBy: "priya",
	})

	assert.NoError(t, err)
	assert.Equal(t, "2", d.Version)
}
