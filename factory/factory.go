// Package factory is the strategy engine: every game genre is one subdirectory
// implementing Strategy. The engine dispatches by genre ID, composes every
// strategy's genre-specific extensions on top of the universal 11 metrics,
// and returns actionable design advice.
//
// Registry model: each subdirectory calls factory.Register in its init().
// Loading a strategy is a single map lookup.
//
// Chinese-mainland vs. overseas considerations are explicit Strategy context:
// StrategyContext.Region = "CN" | "SEA" | "US" | "EU" | "GLOBAL". Every strategy
// may return region-aware recommendations (e.g. Gacha CN vs. US legal caps,
// SLG pay rate ceilings, Casual D1 targets by market).
package factory

import (
	"errors"
	"fmt"
	"sort"

	"github.com/neko233-com/game-ops/num"
)

// Region — target market for the game. Affects lever tuning and recommendations.
// "CN" mainland, "SEA" southeast asia, "US" north america, "EU" europe, "GLOBAL" no region filter.
type Region string

const (
	RegionCN     Region = "CN"
	RegionSEA    Region = "SEA"
	RegionUS     Region = "US"
	RegionEU     Region = "EU"
	RegionGlobal Region = "GLOBAL"
)

// StrategyContext is what every strategy needs at runtime.
type StrategyContext struct {
	GenreID  string
	Region   Region
	Metrics  map[string]num.Num // observed values (universal 11 + any genre-specific)
	BaseUnit string
}

// StrategyResult is what every strategy returns.
type StrategyResult struct {
	Genre        string
	Region       Region
	Universal    map[string]StrategyMetric `json:"universal"`    // universal 11 metrics (copied, normalized view)
	GenreMetrics []StrategyMetric          `json:"genreMetrics"` // genre-specific extra metrics
	Risks        []RiskItem                `json:"risks"`
	Advice       []AdviceItem              `json:"advice"`
	Notes        string                    `json:"notes"` // genre-specific design notes
}

// StrategyMetric is one metric view in a StrategyResult.
type StrategyMetric struct {
	Key      string
	Label    string
	Value    num.Num // exact
	Note     string
	Critical bool
}

// RiskItem — a concrete number that is wrong.
type RiskItem struct {
	Metric   string
	Observed num.Num
	Ideal    num.Num
	Severity string // "warn" | "critical"
	Message  string
}

// AdviceItem — a concrete config edit suggestion.
type AdviceItem struct {
	Field     string
	FromValue string
	ToValue   string
	Reason    string
}

// Strategy is the interface every genre subdirectory must implement.
type Strategy interface {
	// ID returns the genre id (must match genre.Get(id)).
	ID() string
	// Name is the human-readable name in English and Chinese, e.g. "Gacha Collection / 抽卡".
	Name() string
	// GenreMetrics returns the keys of every genre-specific metric this strategy knows about.
	// Keys are arbitrary strings; labels render them in reports.
	GenreMetrics() []MetricLabel
	// Advice returns initial advice strings — free-form defaults, no context needed.
	Advice() []string
	// Evaluate runs genre-specific analysis on top of the universal health report.
	// The strategy MUST NOT mutate ctx.Metrics; it copies values into result.
	Evaluate(ctx StrategyContext) (StrategyResult, error)
}

// MetricLabel is a label for a genre-specific metric.
type MetricLabel struct {
	Key   string
	Label string
	Desc  string
}

var registry = map[string]Strategy{}

// Register adds a strategy. Safe to call from init.
func Register(s Strategy) {
	if _, exists := registry[s.ID()]; exists {
		panic(fmt.Sprintf("factory: duplicate strategy id %q", s.ID()))
	}
	registry[s.ID()] = s
}

// Get returns a strategy by id (or false).
func Get(id string) (Strategy, bool) {
	s, ok := registry[id]
	return s, ok
}

// List returns all registered strategy ids, sorted.
func List() []string {
	ids := make([]string, 0, len(registry))
	for id := range registry {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

// Count returns how many strategies are registered.
func Count() int { return len(registry) }

// Run dispatches: finds the strategy, calls Evaluate, returns the result.
// Unknown genre → error.
func Run(ctx StrategyContext) (StrategyResult, error) {
	s, ok := registry[ctx.GenreID]
	if !ok {
		return StrategyResult{}, fmt.Errorf("factory: no strategy registered for genre %q (have: %s)",
			ctx.GenreID, List())
	}
	return s.Evaluate(ctx)
}

// ValidateMinimum checks the bare minimum a strategy must produce.
// Exported for tests.
func ValidateMinimum(r StrategyResult) error {
	if r.Genre == "" {
		return errors.New("empty genre in result")
	}
	return nil
}
