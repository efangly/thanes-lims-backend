package equipment_test

import (
	"context"

	domainequipment "github.com/efangly/thanes-lims-backend/internal/domain/equipment"
	domainlocation "github.com/efangly/thanes-lims-backend/internal/domain/location"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	domainvendor "github.com/efangly/thanes-lims-backend/internal/domain/vendor"
)

type stubRepo struct {
	store   map[string]domainequipment.Equipment
	created domainequipment.Equipment
	updated domainequipment.Equipment
}

func newStubRepo() *stubRepo { return &stubRepo{store: map[string]domainequipment.Equipment{}} }

func (r *stubRepo) Create(_ context.Context, e domainequipment.Equipment) (domainequipment.Equipment, error) {
	r.created = e
	r.store[e.ID] = e
	return e, nil
}
func (r *stubRepo) FindByID(_ context.Context, id string) (domainequipment.Equipment, error) {
	e, ok := r.store[id]
	if !ok {
		return domainequipment.Equipment{}, shared.ErrNotFound
	}
	return e, nil
}
func (r *stubRepo) List(_ context.Context) ([]domainequipment.Equipment, error) {
	out := make([]domainequipment.Equipment, 0, len(r.store))
	for _, e := range r.store {
		out = append(out, e)
	}
	return out, nil
}
func (r *stubRepo) Update(_ context.Context, e domainequipment.Equipment) (domainequipment.Equipment, error) {
	r.updated = e
	r.store[e.ID] = e
	return e, nil
}

type stubIDGen struct{ n int64 }

func (g *stubIDGen) Next(_ context.Context, _ string, _ *int) (int64, error) {
	g.n++
	return g.n, nil
}

type stubVendors struct{ known map[string]bool }

func (s stubVendors) FindByID(_ context.Context, id string) (domainvendor.Vendor, error) {
	if s.known[id] {
		return domainvendor.Vendor{ID: id}, nil
	}
	return domainvendor.Vendor{}, shared.ErrNotFound
}

type stubLocations struct {
	byID map[string]domainlocation.Location
}

func (s stubLocations) GetByID(_ context.Context, id string) (domainlocation.Location, error) {
	l, ok := s.byID[id]
	if !ok {
		return domainlocation.Location{}, shared.ErrNotFound
	}
	return l, nil
}
