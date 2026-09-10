package user

import (
	"context"
	"fmt"

	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	domainuser "github.com/efangly/thanes-lims-backend/internal/domain/user"
	portuser "github.com/efangly/thanes-lims-backend/internal/ports/user"
)

// errSelfTarget is returned when an admin tries to suspend, retire, or
// otherwise lock out their own account - always a mistake, and it would let
// someone strand the system with no usable admin.
func errSelfTarget(action string) error {
	return fmt.Errorf("%w: cannot %s your own account", shared.ErrValidation, action)
}

// guardNotLastActiveAdmin blocks an action (suspend/retire/role change) on
// target when target is the only remaining active Admin - see ADR 0010.
func guardNotLastActiveAdmin(ctx context.Context, users portuser.UserRepository, target domainuser.User, action string) error {
	if target.Role != domainuser.RoleAdmin || target.Status != domainuser.StatusActive {
		return nil
	}
	count, err := users.CountActiveByRole(ctx, domainuser.RoleAdmin)
	if err != nil {
		return err
	}
	if count <= 1 {
		return fmt.Errorf("%w: cannot %s the last remaining active admin", shared.ErrValidation, action)
	}
	return nil
}
