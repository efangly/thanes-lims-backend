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

func TestDeleteDocumentUseCase_SoftDeletesAndAppendsHistory(t *testing.T) {
	docs := new(mockDocRepo)
	history := new(mockHistoryRepo)

	docs.On("FindByID", mock.Anything, "DOC-00001").Return(document.Document{ID: "DOC-00001", Version: "2"}, nil)
	docs.On("Delete", mock.Anything, "DOC-00001").Return(nil)
	history.On("Append", mock.Anything, mock.MatchedBy(func(h document.DocHistory) bool {
		return h.DocumentID == "DOC-00001" && h.Who == "priya"
	})).Return(document.DocHistory{}, nil)

	uc := applicationdocument.NewDeleteDocumentUseCase(docs, history)
	err := uc.Execute(context.Background(), "DOC-00001", "priya")

	assert.NoError(t, err)
	docs.AssertCalled(t, "Delete", mock.Anything, "DOC-00001")
	history.AssertCalled(t, "Append", mock.Anything, mock.Anything)
}

func TestDeleteDocumentUseCase_LockedRejected(t *testing.T) {
	docs := new(mockDocRepo)
	history := new(mockHistoryRepo)

	docs.On("FindByID", mock.Anything, "DOC-00001").Return(document.Document{ID: "DOC-00001", Locked: true}, nil)

	uc := applicationdocument.NewDeleteDocumentUseCase(docs, history)
	err := uc.Execute(context.Background(), "DOC-00001", "priya")

	assert.ErrorIs(t, err, shared.ErrForbidden)
	docs.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything)
}

func TestDeleteDocumentUseCase_NotFound(t *testing.T) {
	docs := new(mockDocRepo)
	history := new(mockHistoryRepo)

	docs.On("FindByID", mock.Anything, "DOC-09999").Return(document.Document{}, shared.ErrNotFound)

	uc := applicationdocument.NewDeleteDocumentUseCase(docs, history)
	err := uc.Execute(context.Background(), "DOC-09999", "priya")

	assert.ErrorIs(t, err, shared.ErrNotFound)
}
