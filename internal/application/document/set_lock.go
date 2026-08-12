package document

import (
	"context"

	"github.com/efangly/thanes-lims-backend/internal/domain/document"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	domainuser "github.com/efangly/thanes-lims-backend/internal/domain/user"
	portdocument "github.com/efangly/thanes-lims-backend/internal/ports/document"
)

type SetLockUseCase struct {
	documents portdocument.Repository
}

func NewSetLockUseCase(documents portdocument.Repository) *SetLockUseCase {
	return &SetLockUseCase{documents: documents}
}

// Execute is restricted to admin/qa - locking is a document-governance
// action, not a generic edit permission.
func (uc *SetLockUseCase) Execute(ctx context.Context, id string, locked bool, actorRole domainuser.Role) (document.Document, error) {
	if actorRole != domainuser.RoleAdmin && actorRole != domainuser.RoleQA {
		return document.Document{}, shared.ErrForbidden
	}

	d, err := uc.documents.FindByID(ctx, id)
	if err != nil {
		return document.Document{}, err
	}
	d.Locked = locked
	return uc.documents.Update(ctx, d)
}
