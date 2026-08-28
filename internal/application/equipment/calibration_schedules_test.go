package equipment_test

import (
	"context"
	"testing"
	"time"

	applicationequipment "github.com/efangly/thanes-lims-backend/internal/application/equipment"
	domainequipment "github.com/efangly/thanes-lims-backend/internal/domain/equipment"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	portequipment "github.com/efangly/thanes-lims-backend/internal/ports/equipment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubSchedules struct {
	store  map[int64]domainequipment.CalibrationSchedule
	nextID int64
}

func newStubSchedules() *stubSchedules {
	return &stubSchedules{store: map[int64]domainequipment.CalibrationSchedule{}}
}

func (s *stubSchedules) Create(_ context.Context, sc domainequipment.CalibrationSchedule) (domainequipment.CalibrationSchedule, error) {
	s.nextID++
	sc.ID = s.nextID
	s.store[sc.ID] = sc
	return sc, nil
}
func (s *stubSchedules) FindByID(_ context.Context, id int64) (domainequipment.CalibrationSchedule, error) {
	sc, ok := s.store[id]
	if !ok {
		return domainequipment.CalibrationSchedule{}, shared.ErrNotFound
	}
	return sc, nil
}
func (s *stubSchedules) ListByEquipment(_ context.Context, equipmentID string) ([]domainequipment.CalibrationSchedule, error) {
	var out []domainequipment.CalibrationSchedule
	for _, sc := range s.store {
		if sc.EquipmentID == equipmentID {
			out = append(out, sc)
		}
	}
	return out, nil
}
func (s *stubSchedules) Update(_ context.Context, sc domainequipment.CalibrationSchedule) (domainequipment.CalibrationSchedule, error) {
	s.store[sc.ID] = sc
	return sc, nil
}
func (s *stubSchedules) Delete(_ context.Context, id int64) error {
	delete(s.store, id)
	return nil
}

type stubCalibration struct {
	appended []domainequipment.CalibrationEvent
}

func (s *stubCalibration) Append(_ context.Context, ev domainequipment.CalibrationEvent) (domainequipment.CalibrationEvent, error) {
	ev.ID = int64(len(s.appended) + 1)
	s.appended = append(s.appended, ev)
	return ev, nil
}
func (s *stubCalibration) ListByEquipment(context.Context, string) ([]domainequipment.CalibrationEvent, error) {
	return s.appended, nil
}
func (s *stubCalibration) FindByID(_ context.Context, id int64) (domainequipment.CalibrationEvent, error) {
	for _, ev := range s.appended {
		if ev.ID == id {
			return ev, nil
		}
	}
	return domainequipment.CalibrationEvent{}, shared.ErrNotFound
}
func (s *stubCalibration) Search(context.Context, portequipment.CalibrationSearchFilter) ([]domainequipment.CalibrationEvent, error) {
	return s.appended, nil
}

func TestCalibrationSchedule_CreateAndUpdate(t *testing.T) {
	repo := newStubRepo()
	repo.store["EQ-X-001"] = domainequipment.Equipment{ID: "EQ-X-001"}
	scheds := newStubSchedules()
	uc := applicationequipment.NewCalibrationScheduleUseCase(repo, scheds)

	due := time.Now().AddDate(0, 1, 0)
	sc, err := uc.Create(context.Background(), applicationequipment.CreateCalibrationScheduleInput{
		EquipmentID: "EQ-X-001", Label: " สอบเทียบภายใน ", NextDueDate: due,
	})
	require.NoError(t, err)
	assert.Equal(t, "สอบเทียบภายใน", sc.Label)

	_, err = uc.Create(context.Background(), applicationequipment.CreateCalibrationScheduleInput{
		EquipmentID: "EQ-X-001", Label: "x", NextDueDate: due, IntervalMonths: intPtr(0),
	})
	assert.ErrorIs(t, err, shared.ErrValidation)

	// wrong equipment on update -> not found
	_, err = uc.Update(context.Background(), applicationequipment.UpdateCalibrationScheduleInput{
		EquipmentID: "EQ-OTHER", ID: sc.ID,
	})
	assert.ErrorIs(t, err, shared.ErrNotFound)
}

func TestRecordCalibration_AutoAdvancesMatchingSchedule(t *testing.T) {
	repo := newStubRepo()
	repo.store["EQ-X-001"] = domainequipment.Equipment{
		ID: "EQ-X-001", NextCalibrationDue: time.Now().AddDate(0, 0, 3),
	}
	scheds := newStubSchedules()
	interval := 12
	oldDue := time.Now().AddDate(0, 0, 3)
	internal, _ := scheds.Create(context.Background(), domainequipment.CalibrationSchedule{
		EquipmentID: "EQ-X-001", Label: "สอบเทียบภายใน", NextDueDate: oldDue, IntervalMonths: &interval,
	})
	external, _ := scheds.Create(context.Background(), domainequipment.CalibrationSchedule{
		EquipmentID: "EQ-X-001", Label: "สอบเทียบภายนอก", NextDueDate: oldDue,
	})

	cal := &stubCalibration{}
	uc := applicationequipment.NewRecordCalibrationUseCase(repo, cal, scheds)

	_, err := uc.Execute(context.Background(), applicationequipment.RecordCalibrationInput{
		ID:                 "EQ-X-001",
		NextCalibrationDue: time.Now().AddDate(1, 0, 0),
		CalibrationType:    "สอบเทียบภายใน",
		Result:             domainequipment.CalibrationResultPass,
	})
	require.NoError(t, err)

	got, _ := scheds.FindByID(context.Background(), internal.ID)
	assert.True(t, got.NextDueDate.After(oldDue.AddDate(0, 11, 0)), "matching schedule advanced ~12 months")

	unchanged, _ := scheds.FindByID(context.Background(), external.ID)
	assert.Equal(t, oldDue, unchanged.NextDueDate, "non-matching schedule untouched")

	require.Len(t, cal.appended, 1)
	assert.Equal(t, domainequipment.CalibrationResultPass, cal.appended[0].Result)
}

func intPtr(n int) *int { return &n }
