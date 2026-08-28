// Package sample is the Oracle ADB adapter for the samples mirror table
// (scripts/oracle/001_schema.sql) used by the chatbot POC. It maps onto
// oraclesample.MirrorSample - the mirror's own row shape, which keeps
// custodian as a free-text name (the POC ADB has no Users table).
package sample

import (
	"context"
	"database/sql"
	"errors"

	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	"github.com/efangly/thanes-lims-backend/internal/ports/oraclesample"
)

type Repository struct {
	db *sql.DB
}

func New(db *sql.DB) *Repository {
	return &Repository{db: db}
}

const insertSQL = `
INSERT INTO samples (id, name, sample_type, custodian, location, status, received_at)
VALUES (:1, :2, :3, :4, :5, :6, :7)
`

func (r *Repository) Insert(ctx context.Context, s oraclesample.MirrorSample) error {
	var location string
	if s.LocationID != nil {
		location = *s.LocationID
	}
	_, err := r.db.ExecContext(ctx, insertSQL,
		s.ID, s.Name, s.Type, s.Custodian, location, s.Status, s.ReceivedAt,
	)
	return err
}

const selectByIDSQL = `
SELECT id, name, sample_type, custodian, location, status, received_at
FROM samples
WHERE id = :1
`

func (r *Repository) FindByID(ctx context.Context, id string) (oraclesample.MirrorSample, error) {
	var (
		s        oraclesample.MirrorSample
		location string
	)

	row := r.db.QueryRowContext(ctx, selectByIDSQL, id)
	err := row.Scan(&s.ID, &s.Name, &s.Type, &s.Custodian, &location, &s.Status, &s.ReceivedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return oraclesample.MirrorSample{}, shared.ErrNotFound
	}
	if err != nil {
		return oraclesample.MirrorSample{}, err
	}

	if location != "" {
		s.LocationID = &location
	}
	return s, nil
}
