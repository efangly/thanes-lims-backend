package document

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/efangly/thanes-lims-backend/internal/domain/document"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	portdocument "github.com/efangly/thanes-lims-backend/internal/ports/document"
	portfilestorage "github.com/efangly/thanes-lims-backend/internal/ports/filestorage"
	portidgen "github.com/efangly/thanes-lims-backend/internal/ports/idgen"
)

type UploadDocumentUseCase struct {
	documents    portdocument.Repository
	history      portdocument.HistoryRepository
	storage      portfilestorage.FileStorage
	idgen        portidgen.SequenceGenerator
	equipment    portdocument.EquipmentDirectory
	calibrations portdocument.CalibrationEventDirectory
}

func NewUploadDocumentUseCase(documents portdocument.Repository, history portdocument.HistoryRepository, storage portfilestorage.FileStorage, idgen portidgen.SequenceGenerator, equipment portdocument.EquipmentDirectory, calibrations portdocument.CalibrationEventDirectory) *UploadDocumentUseCase {
	return &UploadDocumentUseCase{documents: documents, history: history, storage: storage, idgen: idgen, equipment: equipment, calibrations: calibrations}
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
	// EquipmentID optionally links the new Document to an Equipment (Phase
	// 5). Empty = no link.
	EquipmentID string
	// CalibrationEventID optionally links the new Document to a
	// CalibrationEvent (Phase 6, e.g. a certificate). 0 = no link.
	CalibrationEventID int64
}

// Execute stores the object under a version-scoped key (docs/{id}/{version}/{filename})
// rather than overwriting - every version stays retrievable for audit integrity.
func (uc *UploadDocumentUseCase) Execute(ctx context.Context, in UploadDocumentInput) (document.Document, error) {
	if !in.Type.Valid() {
		return document.Document{}, shared.ErrValidation
	}

	var equipmentID *string
	if id := strings.TrimSpace(in.EquipmentID); id != "" {
		if uc.equipment != nil {
			if _, err := uc.equipment.FindByID(ctx, id); err != nil {
				if errors.Is(err, shared.ErrNotFound) {
					return document.Document{}, fmt.Errorf("%w: equipment %q not found", shared.ErrValidation, id)
				}
				return document.Document{}, err
			}
		}
		equipmentID = &id
	}

	var calibrationEventID *int64
	if in.CalibrationEventID != 0 {
		if uc.calibrations != nil {
			if _, err := uc.calibrations.FindByID(ctx, in.CalibrationEventID); err != nil {
				if errors.Is(err, shared.ErrNotFound) {
					return document.Document{}, fmt.Errorf("%w: calibration event %d not found", shared.ErrValidation, in.CalibrationEventID)
				}
				return document.Document{}, err
			}
		}
		id := in.CalibrationEventID
		calibrationEventID = &id
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
		ID:                 id,
		Name:               in.Name,
		Type:               in.Type,
		Version:            initialVersion,
		CreatedBy:          in.UploadedBy,
		IssuedAt:           now,
		AccessLevel:        in.AccessLevel,
		StorageKey:         key,
		EquipmentID:        equipmentID,
		CalibrationEventID: calibrationEventID,
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
