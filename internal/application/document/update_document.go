package document

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/efangly/thanes-lims-backend/internal/domain/document"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	portdocument "github.com/efangly/thanes-lims-backend/internal/ports/document"
)

// UpdateDocumentUseCase edits a Document's metadata. File content and
// version are out of scope - those go through CreateNewVersionUseCase.
type UpdateDocumentUseCase struct {
	documents    portdocument.Repository
	history      portdocument.HistoryRepository
	equipment    portdocument.EquipmentDirectory
	calibrations portdocument.CalibrationEventDirectory
}

func NewUpdateDocumentUseCase(documents portdocument.Repository, history portdocument.HistoryRepository, equipment portdocument.EquipmentDirectory, calibrations portdocument.CalibrationEventDirectory) *UpdateDocumentUseCase {
	return &UpdateDocumentUseCase{documents: documents, history: history, equipment: equipment, calibrations: calibrations}
}

// UpdateDocumentInput carries only the fields the caller sent: a nil
// pointer means "leave unchanged". For the optional links, a pointer to
// an empty string / zero means "clear the link".
type UpdateDocumentInput struct {
	DocumentID         string
	Name               *string
	Type               *string
	AccessLevel        *string
	EquipmentID        *string
	CalibrationEventID *int64
	ChangeNote         string
	UpdatedBy          string
}

func (uc *UpdateDocumentUseCase) Execute(ctx context.Context, in UpdateDocumentInput) (document.Document, error) {
	d, err := uc.documents.FindByID(ctx, in.DocumentID)
	if err != nil {
		return document.Document{}, err
	}
	if d.Locked {
		return document.Document{}, shared.ErrForbidden
	}

	if in.Name != nil {
		d.Name = *in.Name
	}
	if in.Type != nil {
		t := document.Type(*in.Type)
		if !t.Valid() {
			return document.Document{}, shared.ErrValidation
		}
		d.Type = t
	}
	if in.AccessLevel != nil {
		d.AccessLevel = *in.AccessLevel
	}

	if in.EquipmentID != nil {
		if id := strings.TrimSpace(*in.EquipmentID); id != "" {
			if uc.equipment != nil {
				if _, err := uc.equipment.FindByID(ctx, id); err != nil {
					if errors.Is(err, shared.ErrNotFound) {
						return document.Document{}, fmt.Errorf("%w: equipment %q not found", shared.ErrValidation, id)
					}
					return document.Document{}, err
				}
			}
			d.EquipmentID = &id
		} else {
			d.EquipmentID = nil
		}
	}

	if in.CalibrationEventID != nil {
		if id := *in.CalibrationEventID; id != 0 {
			if uc.calibrations != nil {
				if _, err := uc.calibrations.FindByID(ctx, id); err != nil {
					if errors.Is(err, shared.ErrNotFound) {
						return document.Document{}, fmt.Errorf("%w: calibration event %d not found", shared.ErrValidation, id)
					}
					return document.Document{}, err
				}
			}
			d.CalibrationEventID = &id
		} else {
			d.CalibrationEventID = nil
		}
	}

	updated, err := uc.documents.Update(ctx, d)
	if err != nil {
		return document.Document{}, err
	}

	change := in.ChangeNote
	if strings.TrimSpace(change) == "" {
		change = "แก้ไขข้อมูลเอกสาร"
	}
	if _, err := uc.history.Append(ctx, document.DocHistory{
		DocumentID: updated.ID,
		Version:    updated.Version,
		Change:     change,
		Date:       time.Now(),
		Who:        in.UpdatedBy,
	}); err != nil {
		return document.Document{}, err
	}

	return updated, nil
}
