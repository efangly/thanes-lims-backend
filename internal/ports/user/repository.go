package user

import (
	"context"

	"github.com/efangly/thanes-lims-backend/internal/domain/user"
)

type UserRepository interface {
	Create(ctx context.Context, u user.User) (user.User, error)
	FindByID(ctx context.Context, id int64) (user.User, error)
	FindByEmail(ctx context.Context, email string) (user.User, error)
	List(ctx context.Context) ([]user.User, error)
	Update(ctx context.Context, u user.User) (user.User, error)
	// Retire soft-deletes the User (Retired - irreversible; the row and any
	// FK references to it stay intact). See CONTEXT.md "Retired".
	Retire(ctx context.Context, id int64) error
	// CountByRole counts Users currently holding role.
	CountByRole(ctx context.Context, role user.Role) (int64, error)
	// CountActiveByRole counts Users holding role whose Status is active -
	// the last-active-admin guard shared by role change, suspend and retire
	// (see ADR 0010).
	CountActiveByRole(ctx context.Context, role user.Role) (int64, error)
}

// CustodianChecker answers "does retiring this User orphan any Custodian
// FK?" without application/user taking a dependency on the sample or
// inventory domains (see ADR 0010). Implemented by
// internal/adapters/postgres/user.CustodianChecker.
type CustodianChecker interface {
	// CountCustodianRefs returns how many non-Retired Samples and non-Retired
	// Inventory Items still name userID as their Custodian.
	CountCustodianRefs(ctx context.Context, userID int64) (samples int64, inventoryItems int64, err error)
}
