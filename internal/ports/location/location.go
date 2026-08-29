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
	// lists root Locations across every Kind - callers that want only one
	// Kind's roots use ListRoots instead.
	ListChildren(ctx context.Context, parentID *string) ([]location.Location, error)
	// ListRoots lists the root (parentless) Locations of a single Kind.
	ListRoots(ctx context.Context, kind location.Kind) ([]location.Location, error)
	// FindChildByName looks up a direct child of parentID by name, used to
	// pre-check sibling/root name uniqueness before Create. Returns
	// shared.ErrNotFound if no such child exists.
	FindChildByName(ctx context.Context, parentID *string, name string) (location.Location, error)
	// FindByBarcode resolves a Location Barcode to its Location. Returns
	// shared.ErrNotFound if no non-Retired Location carries that code.
	FindByBarcode(ctx context.Context, code string) (location.Location, error)
	HasChildren(ctx context.Context, id string) (bool, error)
	// UpdateGrid enlarges a Box's Grid to rows x cols. Boxes only grow;
	// callers enforce that (docs/adr/0009).
	UpdateGrid(ctx context.Context, id string, rows, cols int) (location.Location, error)
	Delete(ctx context.Context, id string) error
	// FullPath returns the human-readable ancestor chain down to id (e.g.
	// "Fridge-A / Shelf-2 / Slot-4"), computed from the tree on every call.
	FullPath(ctx context.Context, id string) (string, error)
}
