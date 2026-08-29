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

	// LevelBox is a terminal marker in the Sample tree, not a fixed depth: a
	// Box hangs off a Shelf, Slot, or Sub-slot, carries a Grid, holds many
	// samples by Cell position and never has child Locations (docs/adr/0009).
	// It is deliberately absent from hierarchies below.
	LevelBox LevelType = "box"
)

// maxBoxRows / maxBoxCols bound a Box Grid: rows are named A..Z, columns are
// two digits. Mirrors the CHECK constraint in migration 000036.
const (
	maxBoxRows = 26
	maxBoxCols = 99
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
	// Rows and Cols are the Box Grid dimensions - non-zero only when
	// LevelType is LevelBox (docs/adr/0009).
	Rows int
	Cols int
}

func (l Location) IsRoot() bool {
	return l.ParentID == nil
}

// IsBox reports whether this Location is a Box - a Grid holder that stores
// samples by Cell position and can never have child Locations.
func (l Location) IsBox() bool {
	return l.LevelType == LevelBox
}

// CanParentBox reports whether a Box may be created as a direct child of a
// Location of Kind k at rung parent. Only the Sample tree has Boxes, and
// only a Shelf, Slot, or Sub-slot may hold one (docs/adr/0009).
func CanParentBox(k Kind, parent LevelType) bool {
	if k != KindSampleStorage {
		return false
	}
	switch parent {
	case LevelShelf, LevelSlot, LevelSubSlot:
		return true
	default:
		return false
	}
}

// ValidateBox checks that rows x cols is a Grid a Box may carry: at least
// 1x1, at most 26 rows (A..Z) by 99 columns.
func ValidateBox(rows, cols int) error {
	if rows < 1 || rows > maxBoxRows {
		return fmt.Errorf("%w: box rows must be between 1 and %d", shared.ErrValidation, maxBoxRows)
	}
	if cols < 1 || cols > maxBoxCols {
		return fmt.Errorf("%w: box cols must be between 1 and %d", shared.ErrValidation, maxBoxCols)
	}
	return nil
}

// ParsePosition turns a Cell position string ("A1", "H12") into 1-based
// (row, col) coordinates - "A1" is (1, 1). The row is a single letter A..Z,
// the column is 1..99 with no leading zero.
func ParsePosition(p string) (row, col int, err error) {
	if len(p) < 2 || len(p) > 3 {
		return 0, 0, fmt.Errorf("%w: invalid cell position %q", shared.ErrValidation, p)
	}
	r := p[0]
	if r < 'A' || r > 'Z' {
		return 0, 0, fmt.Errorf("%w: invalid cell position %q", shared.ErrValidation, p)
	}
	digits := p[1:]
	if digits[0] == '0' {
		return 0, 0, fmt.Errorf("%w: invalid cell position %q", shared.ErrValidation, p)
	}
	n := 0
	for _, d := range digits {
		if d < '0' || d > '9' {
			return 0, 0, fmt.Errorf("%w: invalid cell position %q", shared.ErrValidation, p)
		}
		n = n*10 + int(d-'0')
	}
	return int(r-'A') + 1, n, nil
}

// PositionInGrid reports whether p is a well-formed Cell position that falls
// inside this Box's Grid. Always false for a non-Box.
func (l Location) PositionInGrid(p string) bool {
	if !l.IsBox() {
		return false
	}
	row, col, err := ParsePosition(p)
	if err != nil {
		return false
	}
	return row <= l.Rows && col <= l.Cols
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
