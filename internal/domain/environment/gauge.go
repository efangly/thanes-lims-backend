package environment

type Level string

const (
	LevelOK   Level = "ok"
	LevelWarn Level = "warn"
	LevelCrit Level = "crit"
)

type Gauge struct {
	Location string
	Unit     string
	RangeMin float64
	RangeMax float64
}

// DeriveLevel classifies a reading against the gauge's configured range:
// inside range is OK, within a 10%-of-range-width margin outside it is
// Warn, further out is Crit.
func DeriveLevel(value float64, g Gauge) Level {
	if value >= g.RangeMin && value <= g.RangeMax {
		return LevelOK
	}
	margin := (g.RangeMax - g.RangeMin) * 0.1
	if value >= g.RangeMin-margin && value <= g.RangeMax+margin {
		return LevelWarn
	}
	return LevelCrit
}
