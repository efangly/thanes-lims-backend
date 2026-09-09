package document_test

import (
	"context"
	"testing"

	applicationdocument "github.com/efangly/thanes-lims-backend/internal/application/document"
	"github.com/efangly/thanes-lims-backend/internal/domain/document"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRestoreDocumentUseCase_RestoresAndAppendsHistory(t *testing.T) {
	docs := new(mockDocRepo)
	history := new(mockHistoryRepo)

	docs.On("Restore", mock.Anything, "DOC-00001").Return(nil)
	docs.On("FindByID", mock.Anything, "DOC-00001").Return(document.Document{ID: "DOC-00001", Version: "1"}, nil)
	history.On("Append", mock.Anything, mock.AnythingOfType("document.DocHistory")).Return(document.DocHistory{}, nil)

	uc := applicationdocument.NewRestoreDocumentUseCase(docs, history)
	d, err := uc.Execute(context.Background(), "DOC-00001", "priya")

	assert.NoError(t, err)
	assert.Equal(t, "DOC-00001", d.ID)
	history.AssertCalled(t, "Append", mock.Anything, mock.Anything)
}

func TestRestoreDocumentUseCase_NotRetiredConflict(t *testing.T) {
	docs := new(mockDocRepo)
	history := new(mockHistoryRepo)

	docs.On("Restore", mock.Anything, "DOC-00001").Return(shared.ErrConflict)

	uc := applicationdocument.NewRestoreDocumentUseCase(docs, history)
	_, err := uc.Execute(context.Background(), "DOC-00001", "priya")

	assert.ErrorIs(t, err, shared.ErrConflict)
	history.AssertNotCalled(t, "Append", mock.Anything, mock.Anything)
}

func TestRestoreDocumentUseCase_NotFound(t *testing.T) {
	docs := new(mockDocRepo)
	history := new(mockHistoryRepo)

	docs.On("Restore", mock.Anything, "DOC-09999").Return(shared.ErrNotFound)

	uc := applicationdocument.NewRestoreDocumentUseCase(docs, history)
	_, err := uc.Execute(context.Background(), "DOC-09999", "priya")

	assert.ErrorIs(t, err, shared.ErrNotFound)
}
