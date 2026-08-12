package environment_test

import (
	"testing"

	"github.com/efangly/thanes-lims-backend/internal/domain/environment"
	"github.com/stretchr/testify/assert"
)

func TestDeriveLevel(t *testing.T) {
	g := environment.Gauge{RangeMin: 2, RangeMax: 8} // width 6, margin 0.6

	cases := []struct {
		name  string
		value float64
		want  environment.Level
	}{
		{"middle of range", 5, environment.LevelOK},
		{"exactly at min", 2, environment.LevelOK},
		{"exactly at max", 8, environment.LevelOK},
		{"just below min, within margin", 1.5, environment.LevelWarn},
		{"exactly at margin boundary", 1.4, environment.LevelWarn},
		{"just past margin", 1.3, environment.LevelCrit},
		{"far below", 0, environment.LevelCrit},
		{"far above", 20, environment.LevelCrit},
	}

	for _, tc := range cases {
		assert.Equal(t, tc.want, environment.DeriveLevel(tc.value, g), tc.name)
	}
}
