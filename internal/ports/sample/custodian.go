package sample

import (
	"context"

	"github.com/efangly/thanes-lims-backend/internal/domain/user"
)

// CustodianDirectory resolves a Sample's CustodianUserID to a User. It is
// used to validate the id on write (create) and to render the Custodian's
// name on reads (e.g. the PDF test-result report). Returns shared.ErrNotFound
// when no such User exists.
//
// *postgres/user.Repository satisfies this via its FindByID method.
type CustodianDirectory interface {
	FindByID(ctx context.Context, id int64) (user.User, error)
}
