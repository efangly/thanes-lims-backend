package rbac

import (
	"context"
	"time"

	"github.com/efangly/thanes-lims-backend/internal/domain/rbac"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// permissionRow is the Scan target for the permissions/role_permissions/
// roles join - it has no dedicated Model/TableName since it isn't a
// standalone GORM-managed table row.
type permissionRow struct {
	ID        int64
	Module    string
	Action    string
	CreatedAt time.Time
}

func (r *Repository) FindPermissionsByRoleName(ctx context.Context, roleName string) ([]rbac.Permission, error) {
	var rows []permissionRow
	err := r.db.WithContext(ctx).
		Table("permissions AS p").
		Select("p.id, p.module, p.action, p.created_at").
		Joins("JOIN role_permissions rp ON rp.permission_id = p.id").
		Joins("JOIN roles r ON r.id = rp.role_id").
		Where("r.name = ?", roleName).
		Order("p.module, p.action").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	out := make([]rbac.Permission, len(rows))
	for i, row := range rows {
		out[i] = rbac.Permission{
			ID:        row.ID,
			Module:    rbac.Module(row.Module),
			Action:    rbac.Action(row.Action),
			CreatedAt: row.CreatedAt,
		}
	}
	return out, nil
}
