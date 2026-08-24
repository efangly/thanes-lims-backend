package location

import (
	"fmt"

	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
)

// LevelType is the rung of the storage hierarchy a Location occupies. The
// order is fixed - a parent's LevelType must be the immediate predecessor of
// its child's, levels cannot be skipped. See CONTEXT.md#storage-location.
type LevelType string

const (
	LevelCabinet LevelType = "cabinet"
	LevelShelf   LevelType = "shelf"
	LevelSlot    LevelType = "slot"
	LevelSubSlot LevelType = "sub_slot"
)

// levelOrder maps each LevelType to its depth, used to validate that a child
// is exactly one rung below its parent.
var levelOrder = map[LevelType]int{
	LevelCabinet: 0,
	LevelShelf:   1,
	LevelSlot:    2,
	LevelSubSlot: 3,
}

func (t LevelType) Valid() bool {
	_, ok := levelOrder[t]
	return ok
}

var levelsByDepth = []LevelType{LevelCabinet, LevelShelf, LevelSlot, LevelSubSlot}

// Next returns the LevelType immediately below t in the hierarchy, and
// false if t is already the deepest level (sub_slot cannot be subdivided
// further).
func (t LevelType) Next() (LevelType, bool) {
	depth, ok := levelOrder[t]
	if !ok || depth+1 >= len(levelsByDepth) {
		return "", false
	}
	return levelsByDepth[depth+1], true
}

// CanBeChildOf reports whether a Location of LevelType t may be created as a
// direct child of a Location whose LevelType is parent - i.e. t is exactly
// one rung below parent in the fixed hierarchy.
func (t LevelType) CanBeChildOf(parent LevelType) bool {
	parentDepth, ok := levelOrder[parent]
	if !ok {
		return false
	}
	childDepth, ok := levelOrder[t]
	if !ok {
		return false
	}
	return childDepth == parentDepth+1
}

// Location is a node in the physical storage hierarchy where Samples are
// kept (Cabinet > Shelf > Slot > Sub-slot). A nil ParentID marks a Cabinet
// (root). Whether a Location is a leaf - the only kind a Sample may be
// assigned to - is determined by whether it has children, not by LevelType:
// any level can be a leaf if the operator doesn't subdivide it further.
type Location struct {
	ID        string
	ParentID  *string
	Name      string
	LevelType LevelType
}

func (l Location) IsRoot() bool {
	return l.ParentID == nil
}

// ValidateChild checks whether candidate is allowed to be created as a
// direct child of l: candidate's LevelType must be exactly one rung below
// l's, and only a Cabinet (root) may have no ParentID.
func ValidateChild(parent Location, candidate Location) error {
	if !candidate.LevelType.Valid() {
		return fmt.Errorf("%w: invalid level_type %q", shared.ErrValidation, candidate.LevelType)
	}
	if !candidate.LevelType.CanBeChildOf(parent.LevelType) {
		return fmt.Errorf("%w: %s cannot be a child of %s", shared.ErrValidation, candidate.LevelType, parent.LevelType)
	}
	return nil
}
