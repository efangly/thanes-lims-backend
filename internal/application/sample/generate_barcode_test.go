package sample_test

import (
	"context"
	"testing"

	applicationsample "github.com/efangly/thanes-lims-backend/internal/application/sample"
	"github.com/efangly/thanes-lims-backend/internal/domain/sample"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGenerateBarcodeUseCase_GeneratesWhenMissing(t *testing.T) {
	samples := new(mockSampleRepo)
	idgen := new(mockIDGen)

	samples.On("FindByID", mock.Anything, "SMP-1").Return(sample.Sample{ID: "SMP-1"}, nil)
	idgen.On("Next", mock.Anything, "sample_barcode", (*int)(nil)).Return(int64(3), nil)
	samples.On("UpdateBarcodeID", mock.Anything, "SMP-1", mock.MatchedBy(func(code *string) bool {
		return code != nil && *code == "SMP-BC-00003"
	})).Return(sample.Sample{ID: "SMP-1", BarcodeID: strPtr("SMP-BC-00003")}, nil)

	uc := applicationsample.NewGenerateBarcodeUseCase(samples, idgen)
	got, err := uc.Execute(context.Background(), "SMP-1")

	assert.NoError(t, err)
	assert.Equal(t, "SMP-BC-00003", *got.BarcodeID)
}

func TestGenerateBarcodeUseCase_IdempotentWhenPresent(t *testing.T) {
	samples := new(mockSampleRepo)
	idgen := new(mockIDGen)

	samples.On("FindByID", mock.Anything, "SMP-2").Return(sample.Sample{ID: "SMP-2", BarcodeID: strPtr("SMP-BC-00001")}, nil)

	uc := applicationsample.NewGenerateBarcodeUseCase(samples, idgen)
	got, err := uc.Execute(context.Background(), "SMP-2")

	assert.NoError(t, err)
	assert.Equal(t, "SMP-BC-00001", *got.BarcodeID)
	idgen.AssertNotCalled(t, "Next", mock.Anything, mock.Anything, mock.Anything)
	samples.AssertNotCalled(t, "UpdateBarcodeID", mock.Anything, mock.Anything, mock.Anything)
}
