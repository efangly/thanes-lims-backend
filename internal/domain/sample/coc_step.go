package sample

import "time"

type CoCState string

const (
	CoCStateDone    CoCState = "done"
	CoCStateNow     CoCState = "now"
	CoCStatePending CoCState = "pending"
)

type CoCIcon string

const (
	IconPlus  CoCIcon = "Plus"
	IconLoc   CoCIcon = "Loc"
	IconArrow CoCIcon = "Arrow"
	IconTest  CoCIcon = "Test"
	IconCheck CoCIcon = "Check"
)

// CoCStep is a Chain-of-Custody entry: append-only, never mutated or
// deleted, since it's the audit trail of what happened to a Sample.
type CoCStep struct {
	ID         int64
	SampleID   string
	State      CoCState
	Icon       CoCIcon
	Title      string
	Meta       string
	Who        string
	OccurredAt time.Time
}
