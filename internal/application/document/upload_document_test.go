package document_test

import (
	"bytes"
	"context"
	"testing"

	applicationdocument "github.com/efangly/thanes-lims-backend/internal/application/document"
	"github.com/efangly/thanes-lims-backend/internal/domain/document"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUploadDocumentUseCase_CallsFileStorageWithVersionedKey(t *testing.T) {
	docs := new(mockDocRepo)
	history := new(mockHistoryRepo)
	storage := new(mockFileStorage)
	idgen := new(mockIDGen)

	idgen.On("Next", mock.Anything, "document", (*int)(nil)).Return(int64(1), nil)
	storage.On("Upload", mock.Anything, "docs/DOC-00001/1/sop.pdf", mock.Anything, int64(1024), "application/pdf").Return(nil)
	docs.On("Create", mock.Anything, mock.MatchedBy(func(d document.Document) bool {
		return d.ID == "DOC-00001" && d.StorageKey == "docs/DOC-00001/1/sop.pdf" && d.Version == "1"
	})).Return(document.Document{ID: "DOC-00001", Version: "1", StorageKey: "docs/DOC-00001/1/sop.pdf"}, nil)
	history.On("Append", mock.Anything, mock.AnythingOfType("document.DocHistory")).Return(document.DocHistory{}, nil)

	uc := applicationdocument.NewUploadDocumentUseCase(docs, history, storage, idgen, nil, nil)
	d, err := uc.Execute(context.Background(), applicationdocument.UploadDocumentInput{
		Name: "SOP Sample Handling", Type: document.TypeSOP, Filename: "sop.pdf",
		ContentType: "application/pdf", Size: 1024, Content: bytes.NewReader([]byte("x")), UploadedBy: "priya",
	})

	assert.NoError(t, err)
	assert.Equal(t, "docs/DOC-00001/1/sop.pdf", d.StorageKey)
	storage.AssertCalled(t, "Upload", mock.Anything, "docs/DOC-00001/1/sop.pdf", mock.Anything, int64(1024), "application/pdf")
}
