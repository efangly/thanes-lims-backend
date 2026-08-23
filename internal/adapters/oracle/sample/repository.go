// Package sample is the Oracle ADB adapter for the samples mirror table
// (scripts/oracle/001_schema.sql) used by the chatbot POC. It reuses
// internal/domain/sample.Sample directly since the Oracle table's columns
// map 1:1 onto that struct - no separate Oracle-side domain type is needed.
package sample

import (
	"context"
	"database/sql"
	"errors"

	"github.com/efangly/thanes-lims-backend/internal/domain/sample"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
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

func (r *Repository) Insert(ctx context.Context, s sample.Sample) error {
	_, err := r.db.ExecContext(ctx, insertSQL,
		s.ID, s.Name, string(s.Type), s.Custodian, s.Location, string(s.Status), s.ReceivedAt,
	)
	return err
}

const selectByIDSQL = `
SELECT id, name, sample_type, custodian, location, status, received_at
FROM samples
WHERE id = :1
`

func (r *Repository) FindByID(ctx context.Context, id string) (sample.Sample, error) {
	var (
		s          sample.Sample
		sampleType string
		status     string
	)

	row := r.db.QueryRowContext(ctx, selectByIDSQL, id)
	err := row.Scan(&s.ID, &s.Name, &sampleType, &s.Custodian, &s.Location, &status, &s.ReceivedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return sample.Sample{}, shared.ErrNotFound
	}
	if err != nil {
		return sample.Sample{}, err
	}

	s.Type = sample.Type(sampleType)
	s.Status = sample.Status(status)
	return s, nil
}
