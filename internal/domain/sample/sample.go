package sample

import (
	"fmt"
	"time"

	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
)

type Type string

const (
	TypeBlood  Type = "blood"
	TypeUrine  Type = "urine"
	TypeWater  Type = "water"
	TypeTissue Type = "tissue"
	TypeFood   Type = "food"
	TypeSerum  Type = "serum"
)

func (t Type) Valid() bool {
	switch t {
	case TypeBlood, TypeUrine, TypeWater, TypeTissue, TypeFood, TypeSerum:
		return true
	default:
		return false
	}
}

type Status string

const (
	StatusPending     Status = "pending"
	StatusTesting     Status = "testing"
	StatusCompleted   Status = "completed"
	StatusTransferred Status = "transferred"
)

// OccupyingStatuses are the Statuses under which a Sample still occupies its
// assigned leaf Location. Transferred frees the Location back up (see
// CONTEXT.md#storage-location).
var OccupyingStatuses = []Status{StatusPending, StatusTesting, StatusCompleted}

// validTransitions encodes the sample lifecycle state machine. Completed is
// terminal - a sample is never reopened, matching real LIMS practice (a new
// sample/re-test is created instead).
var validTransitions = map[Status][]Status{
	StatusPending:     {StatusTesting, StatusTransferred},
	StatusTesting:     {StatusCompleted, StatusTransferred},
	StatusTransferred: {StatusPending, StatusTesting},
	StatusCompleted:   {},
}

type Sample struct {
	ID   string
	Name string
	Type Type
	// CustodianUserID is the FK to the User responsible for this Sample
	// (see CONTEXT.md "Custodian"). Replaced the old free-text Custodian
	// string in Phase 3.
	CustodianUserID int64
	LocationID      *string
	Status          Status
	ReceivedAt      time.Time
	// BarcodeID is an optional physical/scan identifier, separate from ID
	// (see CONTEXT.md "Barcode ID"). Either user-supplied or generated as
	// SMP-BC-{seq5}; unique across non-Retired Samples when set.
	BarcodeID *string
	// Description is free-text notes captured on the create-sample popup.
	Description string
	// Position is the Cell ("A1", "H12") this Sample occupies when its
	// LocationID points at a Box; nil for a Sample on a plain leaf or with
	// no Location at all (docs/adr/0009).
	Position *string
}

func (s Sample) CanTransition(to Status) bool {
	for _, allowed := range validTransitions[s.Status] {
		if allowed == to {
			return true
		}
	}
	return false
}

// Transition moves the sample to a new status and produces the
// corresponding Chain-of-Custody step to append. It mutates s in place and
// returns the CoC step for the caller to persist alongside it.
func (s *Sample) Transition(to Status, actor string) (CoCStep, error) {
	if !s.CanTransition(to) {
		return CoCStep{}, fmt.Errorf("%w: cannot transition sample from %s to %s", shared.ErrValidation, s.Status, to)
	}
	s.Status = to

	return CoCStep{
		SampleID:   s.ID,
		State:      CoCStateDone,
		Icon:       iconForStatus(to),
		Title:      titleForStatus(to),
		Who:        actor,
		OccurredAt: time.Now(),
	}, nil
}

func iconForStatus(status Status) CoCIcon {
	switch status {
	case StatusTesting:
		return IconTest
	case StatusCompleted:
		return IconCheck
	case StatusTransferred:
		return IconArrow
	default:
		return IconLoc
	}
}

func titleForStatus(status Status) string {
	switch status {
	case StatusTesting:
		return "เริ่มการทดสอบ"
	case StatusCompleted:
		return "การทดสอบเสร็จสิ้น"
	case StatusTransferred:
		return "ส่งต่อแผนก"
	default:
		return "รับตัวอย่างเข้าระบบ"
	}
}
