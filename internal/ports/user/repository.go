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
	// CountByRole counts Users currently holding role - used by the
	// last-admin guard in UpdateUser (see ADR 0002).
	CountByRole(ctx context.Context, role user.Role) (int64, error)
}
