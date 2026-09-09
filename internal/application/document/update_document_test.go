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

func strptr(s string) *string { return &s }

func TestUpdateDocumentUseCase_UpdatesName(t *testing.T) {
	docs := new(mockDocRepo)
	history := new(mockHistoryRepo)

	docs.On("FindByID", mock.Anything, "DOC-00001").Return(document.Document{ID: "DOC-00001", Name: "old", Version: "1"}, nil)
	docs.On("Update", mock.Anything, mock.MatchedBy(func(d document.Document) bool {
		return d.Name == "new"
	})).Return(document.Document{ID: "DOC-00001", Name: "new", Version: "1"}, nil)
	history.On("Append", mock.Anything, mock.AnythingOfType("document.DocHistory")).Return(document.DocHistory{}, nil)

	uc := applicationdocument.NewUpdateDocumentUseCase(docs, history, nil, nil)
	d, err := uc.Execute(context.Background(), applicationdocument.UpdateDocumentInput{
		DocumentID: "DOC-00001", Name: strptr("new"), UpdatedBy: "priya",
	})

	assert.NoError(t, err)
	assert.Equal(t, "new", d.Name)
}

func TestUpdateDocumentUseCase_InvalidType(t *testing.T) {
	docs := new(mockDocRepo)
	history := new(mockHistoryRepo)

	docs.On("FindByID", mock.Anything, "DOC-00001").Return(document.Document{ID: "DOC-00001"}, nil)

	uc := applicationdocument.NewUpdateDocumentUseCase(docs, history, nil, nil)
	_, err := uc.Execute(context.Background(), applicationdocument.UpdateDocumentInput{
		DocumentID: "DOC-00001", Type: strptr("bogus"),
	})

	assert.ErrorIs(t, err, shared.ErrValidation)
	docs.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
}

func TestUpdateDocumentUseCase_LockedRejected(t *testing.T) {
	docs := new(mockDocRepo)
	history := new(mockHistoryRepo)

	docs.On("FindByID", mock.Anything, "DOC-00001").Return(document.Document{ID: "DOC-00001", Locked: true}, nil)

	uc := applicationdocument.NewUpdateDocumentUseCase(docs, history, nil, nil)
	_, err := uc.Execute(context.Background(), applicationdocument.UpdateDocumentInput{
		DocumentID: "DOC-00001", Name: strptr("new"),
	})

	assert.ErrorIs(t, err, shared.ErrForbidden)
	docs.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
}

func TestUpdateDocumentUseCase_ClearsEquipmentLink(t *testing.T) {
	docs := new(mockDocRepo)
	history := new(mockHistoryRepo)

	eq := "EQ-1"
	docs.On("FindByID", mock.Anything, "DOC-00001").Return(document.Document{ID: "DOC-00001", EquipmentID: &eq, Version: "1"}, nil)
	docs.On("Update", mock.Anything, mock.MatchedBy(func(d document.Document) bool {
		return d.EquipmentID == nil
	})).Return(document.Document{ID: "DOC-00001", Version: "1"}, nil)
	history.On("Append", mock.Anything, mock.AnythingOfType("document.DocHistory")).Return(document.DocHistory{}, nil)

	uc := applicationdocument.NewUpdateDocumentUseCase(docs, history, nil, nil)
	_, err := uc.Execute(context.Background(), applicationdocument.UpdateDocumentInput{
		DocumentID: "DOC-00001", EquipmentID: strptr(""),
	})

	assert.NoError(t, err)
	docs.AssertCalled(t, "Update", mock.Anything, mock.Anything)
}
