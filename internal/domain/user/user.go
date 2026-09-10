package user

import "time"

// Status is the reversible active/suspended flag on a User, separate from
// whether the User is Retired (soft-deleted) and from their Role. See
// CONTEXT.md "## User Lifecycle" and ADR 0010.
type Status string

const (
	StatusActive    Status = "active"
	StatusSuspended Status = "suspended"
)

func (s Status) Valid() bool {
	return s == StatusActive || s == StatusSuspended
}

type User struct {
	ID           int64
	Name         string
	Email        string
	PasswordHash string
	Role         Role
	Status       Status
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// IsSuspended reports whether this User is currently barred from logging in.
func (u User) IsSuspended() bool { return u.Status == StatusSuspended }
