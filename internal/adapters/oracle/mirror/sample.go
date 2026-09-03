package mirror

import (
	"context"
	"database/sql"

	dsample "github.com/efangly/thanes-lims-backend/internal/domain/sample"
	duser "github.com/efangly/thanes-lims-backend/internal/domain/user"
	portsample "github.com/efangly/thanes-lims-backend/internal/ports/sample"
)

// UserDirectory resolves a custodian User id to its record - the mirror's
// samples.custodian column is a free-text name. *postgres/user.Repository
// satisfies this.
type UserDirectory interface {
	FindByID(ctx context.Context, id int64) (duser.User, error)
}

// LocationPathResolver renders a Location id as its human-readable full path
// for the mirror's free-text samples.location column. The cached location
// repository satisfies this via FullPath.
type LocationPathResolver interface {
	FullPath(ctx context.Context, id string) (string, error)
}

// unresolved is the placeholder written when a custodian or location cannot
// be resolved, matching the "-" fallback in sticker_data / generate_report.
const unresolved = "-"

const mergeSampleSQL = `
MERGE INTO samples t
USING (SELECT :id AS id FROM dual) s ON (t.id = s.id)
WHEN MATCHED THEN UPDATE SET
    name = :name, sample_type = :stype, custodian = :cust,
    location = :loc, status = :status, received_at = :recv
WHEN NOT MATCHED THEN INSERT
    (id, name, sample_type, custodian, location, status, received_at)
VALUES (:id, :name, :stype, :cust, :loc, :status, :recv)
`

// UpsertSample MERGEs one Sample into the mirror. custodian and location are
// the already-resolved free-text values.
func (m *Mirror) UpsertSample(ctx context.Context, s dsample.Sample, custodian, location string) error {
	_, err := m.db.ExecContext(ctx, mergeSampleSQL,
		sql.Named("id", s.ID),
		sql.Named("name", clip(s.Name, 200)),
		sql.Named("stype", string(s.Type)),
		sql.Named("cust", clip(custodian, 100)),
		sql.Named("loc", clip(location, 100)),
		sql.Named("status", string(s.Status)),
		sql.Named("recv", s.ReceivedAt),
	)
	return err
}

// sampleRepo decorates a SampleRepository, mirroring every successful write
// that changes a column the mirror table carries (id/name/type/custodian/
// location/status/received_at). MoveWithinBox (position only) and
// UpdateBarcodeID (no mirror column) are left to the embedded repo untouched.
type sampleRepo struct {
	portsample.SampleRepository
	mirror    *Mirror
	users     UserDirectory
	locations LocationPathResolver
}

// WrapSample returns next wrapped so its writes also land in the Oracle mirror.
func WrapSample(next portsample.SampleRepository, m *Mirror, users UserDirectory, locations LocationPathResolver) portsample.SampleRepository {
	return &sampleRepo{SampleRepository: next, mirror: m, users: users, locations: locations}
}

func (r *sampleRepo) Create(ctx context.Context, s dsample.Sample) (dsample.Sample, error) {
	created, err := r.SampleRepository.Create(ctx, s)
	if err != nil {
		return created, err
	}
	r.mirrorSample(ctx, created)
	return created, nil
}

func (r *sampleRepo) UpdateStatus(ctx context.Context, s dsample.Sample) (dsample.Sample, error) {
	updated, err := r.SampleRepository.UpdateStatus(ctx, s)
	if err != nil {
		return updated, err
	}
	r.mirrorSample(ctx, updated)
	return updated, nil
}

func (r *sampleRepo) UpdateLocation(ctx context.Context, sampleID string, locationID, position *string) (dsample.Sample, error) {
	updated, err := r.SampleRepository.UpdateLocation(ctx, sampleID, locationID, position)
	if err != nil {
		return updated, err
	}
	r.mirrorSample(ctx, updated)
	return updated, nil
}

// mirrorSample resolves the custodian name and location path, then MERGEs.
func (r *sampleRepo) mirrorSample(ctx context.Context, s dsample.Sample) {
	run(ctx, "sample", s.ID, func(ctx context.Context) error {
		custodian := unresolved
		if u, err := r.users.FindByID(ctx, s.CustodianUserID); err == nil && u.Name != "" {
			custodian = u.Name
		}
		location := unresolved
		if s.LocationID != nil {
			if p, err := r.locations.FullPath(ctx, *s.LocationID); err == nil && p != "" {
				location = p
			}
		}
		return r.mirror.UpsertSample(ctx, s, custodian, location)
	})
}
