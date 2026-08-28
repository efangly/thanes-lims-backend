// Package oraclesample is the port for the Oracle ADB samples mirror used by
// the chatbot POC (Select AI). It is intentionally separate from
// internal/ports/sample, which is the Postgres system-of-record contract -
// conflating the two would misrepresent which store is authoritative.
package oraclesample

import (
	"context"
	"time"
)

// MirrorSample is the Oracle mirror's own row shape. Since Phase 3 the
// Postgres Sample carries a Custodian User FK, but the POC ADB has no Users
// table - the mirror keeps custodian as a free-text name for Select AI's
// NL->SQL demo questions - so it no longer maps 1:1 onto domain
// sample.Sample and gets its own struct here.
type MirrorSample struct {
	ID         string
	Name       string
	Type       string
	Custodian  string
	LocationID *string
	Status     string
	ReceivedAt time.Time
}

type Repository interface {
	Insert(ctx context.Context, s MirrorSample) error
	FindByID(ctx context.Context, id string) (MirrorSample, error)
}
