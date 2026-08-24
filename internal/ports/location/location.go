package location

import (
	"context"

	"github.com/efangly/thanes-lims-backend/internal/domain/location"
)

type LocationRepository interface {
	Create(ctx context.Context, l location.Location) (location.Location, error)
	// CreateMany inserts the batch produced by "generate children" in one
	// round-trip; all-or-nothing.
	CreateMany(ctx context.Context, ls []location.Location) ([]location.Location, error)
	GetByID(ctx context.Context, id string) (location.Location, error)
	// ListChildren lists the direct children of parentID. A nil parentID
	// lists root Locations (Cabinets).
	ListChildren(ctx context.Context, parentID *string) ([]location.Location, error)
	// FindChildByName looks up a direct child of parentID by name, used to
	// pre-check sibling/root name uniqueness before Create. Returns
	// shared.ErrNotFound if no such child exists.
	FindChildByName(ctx context.Context, parentID *string, name string) (location.Location, error)
	HasChildren(ctx context.Context, id string) (bool, error)
	Delete(ctx context.Context, id string) error
	// FullPath returns the human-readable ancestor chain down to id (e.g.
	// "Fridge-A / Shelf-2 / Slot-4"), computed from the tree on every call.
	FullPath(ctx context.Context, id string) (string, error)
}
