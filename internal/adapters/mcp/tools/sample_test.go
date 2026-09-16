package tools_test

import (
	"context"
	"testing"
	"time"

	"github.com/efangly/thanes-lims-backend/internal/adapters/mcp/tools"
	"github.com/efangly/thanes-lims-backend/internal/domain/sample"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	"github.com/efangly/thanes-lims-backend/internal/domain/user"
	portssample "github.com/efangly/thanes-lims-backend/internal/ports/sample"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSampleTools_GetByID_Found(t *testing.T) {
	samples := new(mockSampleRepository)
	users := new(mockUserRepository)
	st := &tools.SampleTools{Samples: samples, Users: users}

	s := sample.Sample{ID: "SMP-1", Name: "Blood A", Type: sample.TypeBlood, Status: sample.StatusPending, CustodianUserID: 7, ReceivedAt: time.Now()}
	samples.On("FindByID", mock.Anything, "SMP-1").Return(s, nil)
	users.On("FindByID", mock.Anything, int64(7)).Return(user.User{ID: 7, Name: "วิภา สายใจ"}, nil)

	_, out, err := st.GetByID(context.Background(), nil, tools.GetSampleByIDInput{ID: "SMP-1"})

	assert.NoError(t, err)
	assert.Equal(t, "SMP-1", out.ID)
	assert.Equal(t, "วิภา สายใจ", out.CustodianName)
	assert.Equal(t, "pending", out.Status)
}

func TestSampleTools_GetByID_NotFound(t *testing.T) {
	samples := new(mockSampleRepository)
	users := new(mockUserRepository)
	st := &tools.SampleTools{Samples: samples, Users: users}

	samples.On("FindByID", mock.Anything, "MISSING").Return(sample.Sample{}, shared.ErrNotFound)

	result, _, err := st.GetByID(context.Background(), nil, tools.GetSampleByIDInput{ID: "MISSING"})

	assert.NoError(t, err)
	assert.True(t, result.IsError)
}

func TestSampleTools_GetByID_EmptyID(t *testing.T) {
	st := &tools.SampleTools{}
	result, _, err := st.GetByID(context.Background(), nil, tools.GetSampleByIDInput{ID: "  "})
	assert.NoError(t, err)
	assert.True(t, result.IsError)
}

func TestSampleTools_ListByStatus_OlderThanDays(t *testing.T) {
	samples := new(mockSampleRepository)
	users := new(mockUserRepository)
	st := &tools.SampleTools{Samples: samples, Users: users}

	now := time.Now()
	old := sample.Sample{ID: "SMP-OLD", Status: sample.StatusPending, ReceivedAt: now.AddDate(0, 0, -15), CustodianUserID: 1}
	recent := sample.Sample{ID: "SMP-NEW", Status: sample.StatusPending, ReceivedAt: now.AddDate(0, 0, -2), CustodianUserID: 1}

	status := sample.StatusPending
	samples.On("List", mock.Anything, portssample.ListFilter{Status: &status}).
		Return([]sample.Sample{old, recent}, nil)
	users.On("FindByID", mock.Anything, int64(1)).Return(user.User{ID: 1, Name: "Somchai"}, nil)

	days := 7
	_, out, err := st.ListByStatus(context.Background(), nil, tools.ListSamplesByStatusInput{Status: "pending", OlderThanDays: &days})

	assert.NoError(t, err)
	assert.Equal(t, 1, out.Count)
	assert.Equal(t, "SMP-OLD", out.Samples[0].ID)
	assert.NotNil(t, out.Samples[0].PendingDays)
	assert.GreaterOrEqual(t, *out.Samples[0].PendingDays, 15)
}

func TestSampleTools_ListByStatus_InvalidStatus(t *testing.T) {
	st := &tools.SampleTools{}
	result, _, err := st.ListByStatus(context.Background(), nil, tools.ListSamplesByStatusInput{Status: "bogus"})
	assert.NoError(t, err)
	assert.True(t, result.IsError)
}

func TestSampleTools_ListByCustodianName_ResolvesAndFilters(t *testing.T) {
	samples := new(mockSampleRepository)
	users := new(mockUserRepository)
	st := &tools.SampleTools{Samples: samples, Users: users}

	users.On("List", mock.Anything).Return([]user.User{
		{ID: 1, Name: "Somchai"},
		{ID: 2, Name: "วิภา สายใจ"},
	}, nil)
	userID := int64(2)
	samples.On("List", mock.Anything, portssample.ListFilter{CustodianUserID: &userID}).
		Return([]sample.Sample{
			{ID: "SMP-3", Status: sample.StatusCompleted, CustodianUserID: 2, ReceivedAt: time.Now()},
			{ID: "SMP-4", Status: sample.StatusTesting, CustodianUserID: 2, ReceivedAt: time.Now()},
		}, nil)
	users.On("FindByID", mock.Anything, int64(2)).Return(user.User{ID: 2, Name: "วิภา สายใจ"}, nil)

	_, out, err := st.ListByCustodianName(context.Background(), nil, tools.ListSamplesByCustodianNameInput{Name: "วิภา สายใจ"})

	assert.NoError(t, err)
	assert.Equal(t, 2, out.Count)
}

func TestSampleTools_ListByCustodianName_NotFound(t *testing.T) {
	samples := new(mockSampleRepository)
	users := new(mockUserRepository)
	st := &tools.SampleTools{Samples: samples, Users: users}

	users.On("List", mock.Anything).Return([]user.User{{ID: 1, Name: "Somchai"}}, nil)

	result, _, err := st.ListByCustodianName(context.Background(), nil, tools.ListSamplesByCustodianNameInput{Name: "Nobody"})

	assert.NoError(t, err)
	assert.True(t, result.IsError)
}
