// Package oraclesample is the port for the Oracle ADB samples mirror used by
// the chatbot POC (Select AI). It is intentionally separate from
// internal/ports/sample, which is the Postgres system-of-record contract -
// conflating the two would misrepresent which store is authoritative.
package oraclesample

import (
	"context"

	"github.com/efangly/thanes-lims-backend/internal/domain/sample"
)

type Repository interface {
	Insert(ctx context.Context, s sample.Sample) error
	FindByID(ctx context.Context, id string) (sample.Sample, error)
}
