package idgen

import (
	"context"

	"gorm.io/gorm"
)

type Adapter struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Adapter {
	return &Adapter{db: db}
}

// Next atomically increments and returns the counter for (scope, year) using
// a single upsert statement - no explicit locking/transaction needed since
// the INSERT ... ON CONFLICT DO UPDATE is atomic at the Postgres level.
func (a *Adapter) Next(ctx context.Context, scope string, year *int) (int64, error) {
	y := 0
	if year != nil {
		y = *year
	}

	var next int64
	err := a.db.WithContext(ctx).Raw(`
		INSERT INTO id_sequences (scope, year, current)
		VALUES (?, ?, 1)
		ON CONFLICT (scope, year)
		DO UPDATE SET current = id_sequences.current + 1
		RETURNING current
	`, scope, y).Scan(&next).Error

	return next, err
}
