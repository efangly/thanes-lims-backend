package sample_test

import (
	"testing"

	"github.com/efangly/thanes-lims-backend/internal/domain/sample"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	"github.com/stretchr/testify/assert"
)

func TestSample_Transition_Valid(t *testing.T) {
	s := sample.Sample{ID: "SMP-2569-00001", Status: sample.StatusPending}

	step, err := s.Transition(sample.StatusTesting, "somchai")

	assert.NoError(t, err)
	assert.Equal(t, sample.StatusTesting, s.Status)
	assert.Equal(t, sample.CoCStateDone, step.State)
	assert.Equal(t, "somchai", step.Who)
}

func TestSample_Transition_Invalid(t *testing.T) {
	s := sample.Sample{ID: "SMP-2569-00001", Status: sample.StatusCompleted}

	_, err := s.Transition(sample.StatusPending, "somchai")

	assert.ErrorIs(t, err, shared.ErrValidation)
	assert.Equal(t, sample.StatusCompleted, s.Status, "status should not change on failed transition")
}

func TestSample_CanTransition(t *testing.T) {
	pending := sample.Sample{Status: sample.StatusPending}
	assert.True(t, pending.CanTransition(sample.StatusTesting))
	assert.False(t, pending.CanTransition(sample.StatusCompleted))

	completed := sample.Sample{Status: sample.StatusCompleted}
	assert.False(t, completed.CanTransition(sample.StatusPending))
	assert.False(t, completed.CanTransition(sample.StatusTesting))
}
