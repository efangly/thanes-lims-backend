package user

import (
	"context"
	"fmt"

	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	portuser "github.com/efangly/thanes-lims-backend/internal/ports/user"
)

// RetireUserUseCase soft-deletes a User (Retired - irreversible). It is
// blocked while the User is still the Custodian of any non-Retired Sample or
// Inventory Item: those FKs must be reassigned first (see CONTEXT.md
// "Retired" and ADR 0010).
type RetireUserUseCase struct {
	users     portuser.UserRepository
	refresh   portuser.RefreshTokenRepository
	custodian portuser.CustodianChecker
}

func NewRetireUserUseCase(users portuser.UserRepository, refresh portuser.RefreshTokenRepository, custodian portuser.CustodianChecker) *RetireUserUseCase {
	return &RetireUserUseCase{users: users, refresh: refresh, custodian: custodian}
}

func (uc *RetireUserUseCase) Execute(ctx context.Context, actorID, targetID int64) error {
	if actorID == targetID {
		return errSelfTarget("retire")
	}

	target, err := uc.users.FindByID(ctx, targetID)
	if err != nil {
		return err
	}
	if err := guardNotLastActiveAdmin(ctx, uc.users, target, "retire"); err != nil {
		return err
	}

	samples, items, err := uc.custodian.CountCustodianRefs(ctx, targetID)
	if err != nil {
		return err
	}
	if samples > 0 || items > 0 {
		return fmt.Errorf("%w: user is still the custodian of %d sample(s) and %d inventory item(s) - reassign them first",
			shared.ErrConflict, samples, items)
	}

	if err := uc.users.Retire(ctx, targetID); err != nil {
		return err
	}
	// A Retired User must not keep a live Session.
	return uc.refresh.RevokeAllForUser(ctx, targetID)
}
