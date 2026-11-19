// Package genre models game genres for long-term operation analysis.
//
// The operation engine is genre-agnostic: every genre is described by the same
// universal live-ops metrics, each with a genre-specific ideal, tolerance band
// and weight. A broad built-in registry covers the major market categories,
// and Register lets callers add ANY genre (including future ones), so coverage
// is open-ended rather than a closed enum.
package genre

import (
	"errors"
	"sort"

	"github.com/neko233/game-ops/num"
)

// Universal live-ops metric keys.
const (
	MetricRetentionD1     = "retention_d1"
	MetricRetentionD7     = "retention_d7"
	MetricRetentionD30    = "retention_d30"
	MetricSourceSinkRatio = "economy_source_sink_ratio"
	MetricInflation       = "economy_inflation"
	MetricPayRate         = "monetization_pay_rate"
	MetricARPU            = "monetization_arpu"
	MetricLTV             = "monetization_ltv"
	MetricProgressionPace = "progression_pace"
	MetricContentCadence  = "content_cadence_days"
	MetricContentDebt     = "content_debt_ratio"
)

// MetricMeta describes a universal metric.
type MetricMeta struct {
	Label string
	Desc  string
	Unit  string
}

var metrics = map[string]MetricMeta{
	MetricRetentionD1:     {"Day-1 retention", "Share of new users active the next day", "ratio"},
	MetricRetentionD7:     {"Day-7 retention", "Share of new users active on day seven", "ratio"},
	MetricRetentionD30:    {"Day-30 retention", "Share of new users active on day thirty", "ratio"},
	MetricSourceSinkRatio: {"Source/sink ratio", "Currency generated vs consumed; 1 = balanced", "ratio"},
	MetricInflation:       {"Inflation index", "Price-level index; 1 = stable", "index"},
	MetricPayRate:         {"Payer conversion", "Share of users who pay", "ratio"},
	MetricARPU:            {"ARPU index", "Average revenue per user, normalized", "index"},
	MetricLTV:             {"LTV index", "Lifetime value relative to a baseline", "index"},
	MetricProgressionPace: {"Progression pace", "Actual vs intended progression speed; 1 = on target", "ratio"},
	MetricContentCadence:  {"Content cadence", "Days between content drops", "days"},
	MetricContentDebt:     {"Content debt", "Content backlog relative to consumption; 1 = healthy", "ratio"},
}

// Metrics returns a copy of the universal metric metadata.
func Metrics() map[string]MetricMeta {
	out := make(map[string]MetricMeta, len(metrics))
	for k, v := range metrics {
		out[k] = v
	}
	return out
}

// Lever is a genre's target band and importance for one universal metric.
type Lever struct {
	Key    string  `json:"key"`
	Ideal  num.Num `json:"ideal"`
	Min    num.Num `json:"min"`
	Max    num.Num `json:"max"`
	Weight num.Num `json:"weight"`
}

// Genre describes how the universal levers apply to one game category.
type Genre struct {
	ID     string   `json:"id"`
	Name   string   `json:"name"`
	Tags   []string `json:"tags"`
	Levers []Lever  `json:"levers"`
}

var registry = map[string]Genre{}

// Register adds a genre. It errors on an empty ID or a duplicate ID.
func Register(g Genre) error {
	if g.ID == "" {
		return errors.New("genre: empty ID")
	}
	if _, exists := registry[g.ID]; exists {
		return errors.New("genre: duplicate ID: " + g.ID)
	}
	registry[g.ID] = g
	return nil
}

// Get returns a copy of a registered genre.
func Get(id string) (Genre, bool) {
	g, ok := registry[id]
	return g, ok
}

// MustGet is Get but panics when the genre is unknown.
func MustGet(id string) Genre {
	g, ok := Get(id)
	if !ok {
		panic("genre: unknown genre: " + id)
	}
	return g
}

// List returns all registered genre IDs, sorted.
func List() []string {
	ids := make([]string, 0, len(registry))
	for id := range registry {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

// LeverFor returns a genre's lever for a metric key.
func LeverFor(g Genre, key string) (Lever, bool) {
	for _, l := range g.Levers {
		if l.Key == key {
			return l, true
		}
	}
	return Lever{}, false
}

func lv(key, ideal, min, max, weight string) Lever {
	return Lever{
		Key: key, Ideal: num.MustFromString(ideal), Min: num.MustFromString(min),
		Max: num.MustFromString(max), Weight: num.MustFromString(weight),
	}
}

func baseLevers() []Lever {
	return []Lever{
		lv(MetricRetentionD1, "0.40", "0.15", "0.70", "1.0"),
		lv(MetricRetentionD7, "0.18", "0.05", "0.40", "1.0"),
		lv(MetricRetentionD30, "0.08", "0.02", "0.20", "1.0"),
		lv(MetricSourceSinkRatio, "1", "0.7", "1.4", "1.0"),
		lv(MetricInflation, "1", "0.8", "1.3", "1.0"),
		lv(MetricPayRate, "0.03", "0.01", "0.15", "1.0"),
		lv(MetricARPU, "1", "0.5", "3", "0.8"),
		lv(MetricLTV, "5", "1", "20", "0.8"),
		lv(MetricProgressionPace, "1", "0.6", "1.6", "0.8"),
		lv(MetricContentCadence, "14", "7", "45", "0.8"),
		lv(MetricContentDebt, "1", "0.5", "2", "0.6"),
	}
}

func mk(id, name string, tags []string, overrides ...Lever) Genre {
	levers := baseLevers()
	for _, o := range overrides {
		replaced := false
		for i := range levers {
			if levers[i].Key == o.Key {
				levers[i] = o
				replaced = true
				break
			}
		}
		if !replaced {
			levers = append(levers, o)
		}
	}
	return Genre{ID: id, Name: name, Tags: tags, Levers: levers}
}

func init() {
	builtins := []Genre{
		mk("rpg", "Role-Playing", []string{"rpg", "progression"},
			lv(MetricLTV, "8", "2", "30", "0.9")),
		mk("action-rpg", "Action RPG", []string{"rpg", "action"}),
		mk("mmo", "Massively Multiplayer Online", []string{"mmo", "social"},
			lv(MetricContentCadence, "7", "3", "21", "1.0"),
			lv(MetricRetentionD30, "0.12", "0.03", "0.25", "1.2")),
		mk("mmorpg", "MMORPG", []string{"mmo", "rpg"},
			lv(MetricContentCadence, "7", "3", "21", "1.0")),
		mk("slg", "Strategy / 4X", []string{"strategy", "competitive"},
			lv(MetricLTV, "12", "2", "40", "1.0"),
			lv(MetricRetentionD30, "0.10", "0.02", "0.20", "1.1")),
		mk("moba", "MOBA", []string{"competitive", "esports"},
			lv(MetricContentCadence, "21", "7", "60", "0.7"),
			lv(MetricPayRate, "0.02", "0.01", "0.10", "0.8")),
		mk("fps", "First-Person Shooter", []string{"shooter", "action"},
			lv(MetricContentCadence, "21", "7", "60", "0.8")),
		mk("tps", "Third-Person Shooter", []string{"shooter", "action"}),
		mk("battle-royale", "Battle Royale", []string{"shooter", "competitive"},
			lv(MetricContentCadence, "21", "7", "60", "0.9")),
		mk("card", "Collectible Card Game", []string{"card", "strategy"},
			lv(MetricPayRate, "0.05", "0.02", "0.20", "1.2")),
		mk("gacha", "Gacha Collection", []string{"gacha", "collection"},
			lv(MetricPayRate, "0.06", "0.02", "0.25", "1.3"),
			lv(MetricLTV, "10", "2", "35", "1.1")),
		mk("casual", "Casual", []string{"casual", "hyper-casual"},
			lv(MetricPayRate, "0.01", "0.003", "0.05", "0.8"),
			lv(MetricRetentionD1, "0.35", "0.10", "0.60", "1.1")),
		mk("puzzle", "Puzzle", []string{"puzzle", "casual"}),
		mk("match3", "Match-3", []string{"puzzle", "casual"},
			lv(MetricContentCadence, "14", "7", "30", "1.0")),
		mk("idle", "Idle / Incremental", []string{"idle", "casual"},
			lv(MetricRetentionD30, "0.06", "0.01", "0.15", "1.1")),
		mk("simulation", "Simulation", []string{"simulation"}),
		mk("sandbox", "Sandbox", []string{"sandbox", "creative"},
			lv(MetricContentCadence, "30", "14", "90", "0.7")),
		mk("survival", "Survival", []string{"survival", "crafting"}),
		mk("roguelike", "Roguelike / Roguelite", []string{"roguelike", "procedural"},
			lv(MetricContentCadence, "21", "7", "60", "0.9")),
		mk("platformer", "Platformer", []string{"action", "platformer"}),
		mk("fighting", "Fighting", []string{"competitive", "fighting"}),
		mk("sports", "Sports", []string{"sports", "competitive"}),
		mk("racing", "Racing", []string{"racing", "competitive"}),
		mk("rhythm", "Rhythm", []string{"rhythm", "music"}),
		mk("tower-defense", "Tower Defense", []string{"strategy", "td"}),
		mk("visual-novel", "Visual Novel", []string{"narrative", "casual"},
			lv(MetricPayRate, "0.04", "0.01", "0.15", "1.0")),
		mk("adventure", "Adventure", []string{"adventure", "narrative"}),
		mk("horror", "Horror", []string{"horror", "narrative"}),
		mk("strategy", "Strategy", []string{"strategy"}),
	}
	for _, g := range builtins {
		_ = Register(g)
	}
}
