package location

import (
	"fmt"

	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
)

// Kind is the discriminator that says which subsystem a Location tree
// belongs to - see docs/adr/0007 and CONTEXT.md "Location Kind".
//
//   - KindSampleStorage: the original tree (cabinet > shelf > slot > sub_slot),
//     leaf-only assignment, occupancy-checked.
//   - KindEquipmentStorage: a second tree (building > room > zone > cabinet >
//     shelf), no occupancy constraint, shared by Equipment and Inventory Item.
type Kind string

const (
	KindSampleStorage    Kind = "sample_storage"
	KindEquipmentStorage Kind = "equipment_storage"
)

// LevelType is the rung of the storage hierarchy a Location occupies. The
// order is fixed *within a Kind* - a parent's LevelType must be the
// immediate predecessor of its child's, levels cannot be skipped. The same
// LevelType string (e.g. "cabinet") may appear in more than one Kind's
// hierarchy at a different depth; it is always disambiguated by the
// Location's Kind. See CONTEXT.md#storage-location.
type LevelType string

const (
	// KindSampleStorage levels.
	LevelCabinet LevelType = "cabinet"
	LevelShelf   LevelType = "shelf"
	LevelSlot    LevelType = "slot"
	LevelSubSlot LevelType = "sub_slot"

	// KindEquipmentStorage levels (building > room > zone > cabinet > shelf).
	// LevelCabinet and LevelShelf are reused from the Sample hierarchy.
	LevelBuilding LevelType = "building"
	LevelRoom     LevelType = "room"
	LevelZone     LevelType = "zone"
)

// hierarchies is the fixed, ordered list of LevelTypes for each Kind, root
// first. Depth = index in the slice.
var hierarchies = map[Kind][]LevelType{
	KindSampleStorage:    {LevelCabinet, LevelShelf, LevelSlot, LevelSubSlot},
	KindEquipmentStorage: {LevelBuilding, LevelRoom, LevelZone, LevelCabinet, LevelShelf},
}

// Valid reports whether k is a known Location Kind.
func (k Kind) Valid() bool {
	_, ok := hierarchies[k]
	return ok
}

// Levels returns k's fixed level hierarchy, root first.
func (k Kind) Levels() []LevelType {
	return hierarchies[k]
}

// RootLevel returns the LevelType a root (parentless) Location of Kind k
// must have, and false if k is not a valid Kind.
func (k Kind) RootLevel() (LevelType, bool) {
	levels := hierarchies[k]
	if len(levels) == 0 {
		return "", false
	}
	return levels[0], true
}

func levelDepth(k Kind, t LevelType) (int, bool) {
	for i, lt := range hierarchies[k] {
		if lt == t {
			return i, true
		}
	}
	return 0, false
}

// LevelValidForKind reports whether t is a rung in k's hierarchy.
func LevelValidForKind(k Kind, t LevelType) bool {
	_, ok := levelDepth(k, t)
	return ok
}

// ChildLevel returns the LevelType immediately below parent in Kind k's
// hierarchy, and false if parent is already the deepest level or is not a
// rung of k.
func ChildLevel(k Kind, parent LevelType) (LevelType, bool) {
	depth, ok := levelDepth(k, parent)
	if !ok {
		return "", false
	}
	levels := hierarchies[k]
	if depth+1 >= len(levels) {
		return "", false
	}
	return levels[depth+1], true
}

// Location is a node in a physical storage hierarchy. A nil ParentID marks
// a root (a Cabinet for KindSampleStorage, a Building for
// KindEquipmentStorage). Whether a Location is a leaf - the only kind a
// Sample may be assigned to, for KindSampleStorage - is determined by
// whether it has children, not by LevelType. BarcodeCode is an
// auto-generated scan code, unique across non-Retired Locations.
type Location struct {
	ID          string
	ParentID    *string
	Name        string
	Kind        Kind
	LevelType   LevelType
	BarcodeCode string
}

func (l Location) IsRoot() bool {
	return l.ParentID == nil
}

// ValidateChild checks whether candidate is allowed to be created as a
// direct child of parent: same Kind, and candidate's LevelType is exactly
// one rung below parent's in that Kind's fixed hierarchy.
func ValidateChild(parent Location, candidate Location) error {
	if candidate.Kind != parent.Kind {
		return fmt.Errorf("%w: child Kind %q does not match parent Kind %q", shared.ErrValidation, candidate.Kind, parent.Kind)
	}
	if !LevelValidForKind(candidate.Kind, candidate.LevelType) {
		return fmt.Errorf("%w: invalid level_type %q for kind %q", shared.ErrValidation, candidate.LevelType, candidate.Kind)
	}
	want, ok := ChildLevel(parent.Kind, parent.LevelType)
	if !ok || candidate.LevelType != want {
		return fmt.Errorf("%w: %s cannot be a child of %s", shared.ErrValidation, candidate.LevelType, parent.LevelType)
	}
	return nil
}
