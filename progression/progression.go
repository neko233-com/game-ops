// Package progression answers the lead designer's question: "how long does a
// normal player take to reach max level?" It works entirely with exact rational
// arithmetic, supports player archetypes with different time budgets and
// efficiency, adjusts for retention, and computes weighted percentiles.
package progression

import (
	"errors"
	"fmt"
	"sort"
	"sync"

	"github.com/neko233-com/game-ops/design"
	"github.com/neko233-com/game-ops/num"
)

var (
	zero = num.Zero()
	one  = num.One()
)

// Activity is an XP source with a hard daily time cap.
type Activity struct {
	Name            string
	XPPerMinute     num.Num
	DailyCapMinutes uint
}

// Archetype is a player cohort with a time budget, efficiency, population
// weight and observed retention anchors (D1/D7/D30).
type Archetype struct {
	Name          string
	MinutesPerDay uint
	Efficiency    num.Num
	Weight        num.Num
	RetentionD1   num.Num
	RetentionD7   num.Num
	RetentionD30  num.Num
}

// Scenario is one time-to-cap experiment.
type Scenario struct {
	Name       string
	MaxLevel   uint
	XPCurve    design.Curve
	Activities []Activity
	Archetypes []Archetype
}

// Result is an archetype's exact time-to-cap result.
type Result struct {
	ScenarioName  string
	ArchetypeName string
	TotalXP       num.Num
	DailyXP       num.Num
	Days          num.Num
	SurvivalAtCap num.Num
	Weight        num.Num
}

// Summary is one scenario's results plus population percentiles.
type Summary struct {
	ScenarioName     string
	Results          []Result
	Percentiles      map[string]num.Num
	PopulationWeight num.Num
}

func totalXP(s Scenario) num.Num {
	total := num.Zero()
	for level := uint(1); level < s.MaxLevel; level++ {
		total = total.Add(s.XPCurve.At(level))
	}
	return total
}

func dailyXP(a Archetype, activities []Activity) num.Num {
	remaining := num.FromInt64(int64(a.MinutesPerDay))
	base := num.Zero()
	for _, act := range activities {
		if remaining.Sign() <= 0 {
			break
		}
		cap := num.FromInt64(int64(act.DailyCapMinutes))
		minutes := remaining.Min(cap)
		base = base.Add(minutes.Mul(act.XPPerMinute))
		remaining = remaining.Sub(minutes)
	}
	return base.Mul(a.Efficiency)
}

// SurvivalAt returns the share of a cohort still playing on a given day.
// It uses exact piecewise-linear interpolation between the D1/D7/D30 anchors
// and a linear continuation of the D7->D30 slope after day 30.
func SurvivalAt(a Archetype, day num.Num) (num.Num, bool) {
	if day.Sign() < 0 {
		return zero, false
	}
	type anchor struct{ d, v num.Num }
	anchors := []anchor{
		{num.Zero(), one}, {num.One(), a.RetentionD1},
		{num.MustFromString("7"), a.RetentionD7},
		{num.MustFromString("30"), a.RetentionD30},
	}
	if day.Cmp(anchors[len(anchors)-1].d) > 0 {
		last := anchors[len(anchors)-1]
		prev := anchors[len(anchors)-2]
		slope, err := last.v.Sub(prev.v).Div(last.d.Sub(prev.d))
		if err != nil {
			return zero, false
		}
		survival := last.v.Add(day.Sub(last.d).Mul(slope))
		return survival.Clamp(zero, one), true
	}
	for i := 1; i < len(anchors); i++ {
		if day.Cmp(anchors[i].d) <= 0 {
			lo, hi := anchors[i-1], anchors[i]
			fraction, err := day.Sub(lo.d).Div(hi.d.Sub(lo.d))
			if err != nil {
				return zero, false
			}
			survival := lo.v.Add(hi.v.Sub(lo.v).Mul(fraction))
			return survival.Clamp(zero, one), true
		}
	}
	return one, true
}

// Run computes exact time-to-cap for every archetype in a scenario.
func Run(s Scenario) (Summary, error) {
	if s.MaxLevel < 2 {
		return Summary{}, errors.New("progression: max level must be >= 2")
	}
	if len(s.Activities) == 0 {
		return Summary{}, errors.New("progression: no activities")
	}
	if len(s.Archetypes) == 0 {
		return Summary{}, errors.New("progression: no archetypes")
	}
	for i, act := range s.Activities {
		if act.XPPerMinute.Sign() <= 0 {
			return Summary{}, fmt.Errorf("progression: activity %d has non-positive XP rate", i)
		}
	}
	xp := totalXP(s)
	summary := Summary{ScenarioName: s.Name, Percentiles: map[string]num.Num{}}
	for _, a := range s.Archetypes {
		if a.MinutesPerDay == 0 {
			return Summary{}, fmt.Errorf("progression: archetype %q has no play time", a.Name)
		}
		if a.Efficiency.Sign() <= 0 {
			return Summary{}, fmt.Errorf("progression: archetype %q has non-positive efficiency", a.Name)
		}
		perDay := dailyXP(a, s.Activities)
		if perDay.IsZero() {
			return Summary{}, fmt.Errorf("progression: archetype %q has zero daily XP", a.Name)
		}
		days, err := xp.Div(perDay)
		if err != nil {
			return Summary{}, err
		}
		survival, ok := SurvivalAt(a, days)
		if !ok {
			return Summary{}, fmt.Errorf("progression: cannot compute survival for %q", a.Name)
		}
		weight := a.Weight
		if weight.Sign() <= 0 {
			weight = one
		}
		summary.Results = append(summary.Results, Result{
			ScenarioName: s.Name, ArchetypeName: a.Name, TotalXP: xp,
			DailyXP: perDay, Days: days, SurvivalAtCap: survival, Weight: weight,
		})
		summary.PopulationWeight = summary.PopulationWeight.Add(weight)
	}

	sorted := append([]Result(nil), summary.Results...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Days.Cmp(sorted[j].Days) < 0 })
	for label, p := range map[string]num.Num{
		"p10": num.MustFromString("1/10"), "p25": num.MustFromString("1/4"),
		"p50": num.MustFromString("1/2"), "p75": num.MustFromString("3/4"),
		"p90": num.MustFromString("9/10"),
	} {
		cumulative := num.Zero()
		threshold := summary.PopulationWeight.Mul(p)
		for _, r := range sorted {
			cumulative = cumulative.Add(r.Weight)
			if cumulative.Cmp(threshold) >= 0 {
				summary.Percentiles[label] = r.Days
				break
			}
		}
	}
	return summary, nil
}

// Batch runs many scenarios in parallel with bounded concurrency. Results are
// returned in input order, so analysis is deterministic regardless of workers.
func Batch(scenarios []Scenario, workers int) ([]Summary, error) {
	if workers < 1 {
		workers = 1
	}
	results := make([]Summary, len(scenarios))
	errs := make([]error, len(scenarios))
	jobs := make(chan int, workers*2)
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for idx := range jobs {
				results[idx], errs[idx] = Run(scenarios[idx])
			}
		}()
	}
	for i := range scenarios {
		jobs <- i
	}
	close(jobs)
	wg.Wait()
	for _, err := range errs {
		if err != nil {
			return nil, err
		}
	}
	return results, nil
}
