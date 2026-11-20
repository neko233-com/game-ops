// Package ops evaluates long-term-operation health in a genre-agnostic way.
//
// A Snapshot supplies observed values for the universal metrics; the genre's
// levers define the ideal, tolerance band and weight for each. All scoring is
// exact arbitrary-precision arithmetic via num.Num.
package ops

import (
	"errors"
	"fmt"

	"github.com/neko233-com/game-ops/genre"
	"github.com/neko233-com/game-ops/num"
)

// Snapshot is one observation of a live game's operation metrics.
type Snapshot struct {
	GenreID string
	Values  map[string]num.Num
}

// Diagnostic is the scored status of one metric.
type Diagnostic struct {
	Metric    string
	Observed  num.Num
	Ideal     num.Num
	Deviation num.Num
	Fitness   num.Num // in [0,1]
	Weight    num.Num
	Status    string // "ok", "warn" or "critical"
}

// Report is the result of an evaluation.
type Report struct {
	GenreID     string
	GenreName   string
	Score       num.Num // weighted mean of fitness, in [0,1]
	Verdict     string  // "healthy", "attention" or "critical"
	Diagnostics []Diagnostic
}

var (
	two       = num.MustFromString("2")
	zero      = num.Zero()
	one       = num.One()
	healthyAt = num.MustFromString("4/5")
	attnAt    = num.MustFromString("3/5")
)

// Evaluate scores a snapshot against its genre's levers.
func Evaluate(s Snapshot) (Report, error) {
	g, ok := genre.Get(s.GenreID)
	if !ok {
		return Report{}, fmt.Errorf("ops: unknown genre %q", s.GenreID)
	}
	if len(g.Levers) == 0 {
		return Report{}, errors.New("ops: genre has no levers")
	}

	totalWeight := num.Zero()
	weighted := num.Zero()
	diagnostics := make([]Diagnostic, 0, len(g.Levers))

	for _, lever := range g.Levers {
		observed, ok := s.Values[lever.Key]
		if !ok {
			return Report{}, fmt.Errorf("ops: missing metric %q", lever.Key)
		}
		if lever.Weight.Sign() <= 0 {
			return Report{}, fmt.Errorf("ops: non-positive weight for metric %q", lever.Key)
		}
		band, err := lever.Max.Sub(lever.Min).Div(two)
		if err != nil || band.Sign() <= 0 {
			return Report{}, fmt.Errorf("ops: invalid band for metric %q", lever.Key)
		}
		deviation := observed.Sub(lever.Ideal).Abs()
		ratio, err := deviation.Div(band)
		if err != nil {
			return Report{}, err
		}
		// fitness = 1 - deviation/band, clamped to [0,1]: 1 at the ideal,
		// 0 at the edges of the band and beyond.
		fitness := one.Sub(ratio).Clamp(zero, one)

		status := "ok"
		if observed.Cmp(lever.Min) < 0 || observed.Cmp(lever.Max) > 0 {
			status = "critical"
		} else if fitness.Cmp(attnAt) < 0 {
			status = "warn"
		}

		diagnostics = append(diagnostics, Diagnostic{
			Metric: lever.Key, Observed: observed, Ideal: lever.Ideal,
			Deviation: deviation, Fitness: fitness, Weight: lever.Weight,
			Status: status,
		})
		weighted = weighted.Add(fitness.Mul(lever.Weight))
		totalWeight = totalWeight.Add(lever.Weight)
	}

	score, err := weighted.Div(totalWeight)
	if err != nil {
		return Report{}, err
	}
	verdict := "critical"
	if score.Cmp(healthyAt) >= 0 {
		verdict = "healthy"
	} else if score.Cmp(attnAt) >= 0 {
		verdict = "attention"
	}

	return Report{
		GenreID: g.ID, GenreName: g.Name, Score: score, Verdict: verdict,
		Diagnostics: diagnostics,
	}, nil
}
