// Package genre — data-driven catalog extensions.
//
// The built-in registry lives in genre.go. This file makes the catalog OPEN at
// runtime: any genre can be described as JSON (Spec), loaded from disk, derived
// from tags, or auto-created for an unseen ID. No code generation, no rebuild.
package genre

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/neko233-com/game-ops/num"
)

// Spec is the JSON-serializable description of one genre.
// Levers use num.Num values which marshal as exact strings, so precision
// survives a JSON round-trip.
type Spec struct {
	ID     string   `json:"id"`
	Name   string   `json:"name"`
	Tags   []string `json:"tags,omitempty"`
	Levers []Lever  `json:"levers"`
}

// Upsert registers g, replacing any existing genre with the same ID.
// Use Register when duplicates must error; use Upsert for data-driven loads.
func Upsert(g Genre) error {
	if g.ID == "" {
		return errors.New("genre: empty ID")
	}
	if len(g.Levers) == 0 {
		return errors.New("genre: no levers: " + g.ID)
	}
	registry[g.ID] = g
	return nil
}

// LoadSpecs upserts every spec. Later specs win on duplicate IDs.
// Returns the number of genres loaded.
func LoadSpecs(specs []Spec) (int, error) {
	for _, s := range specs {
		g := Genre{ID: s.ID, Name: s.Name, Tags: s.Tags, Levers: s.Levers}
		if g.Name == "" {
			g.Name = s.ID
		}
		if err := Upsert(g); err != nil {
			return 0, err
		}
	}
	return len(specs), nil
}

// LoadFile loads an array of Specs (or a single Spec object) from a JSON file.
func LoadFile(path string) (int, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	var many []Spec
	if err := json.Unmarshal(b, &many); err == nil {
		return LoadSpecs(many)
	}
	var one Spec
	if err := json.Unmarshal(b, &one); err != nil {
		return 0, err
	}
	return LoadSpecs([]Spec{one})
}

// LoadDir loads every *.json file in dir (non-recursive), sorted by filename so
// later files deterministically override earlier ones.
func LoadDir(dir string) (int, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0, err
	}
	files := make([]string, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(strings.ToLower(e.Name()), ".json") {
			files = append(files, filepath.Join(dir, e.Name()))
		}
	}
	sort.Strings(files)
	total := 0
	for _, f := range files {
		n, err := LoadFile(f)
		if err != nil {
			return total, errors.New("genre: " + f + ": " + err.Error())
		}
		total += n
	}
	return total, nil
}

// Export returns every registered genre as a sorted slice of Specs.
// Round-trip: Export -> json.Marshal -> LoadSpecs reproduces the registry.
func Export() []Spec {
	ids := List()
	out := make([]Spec, 0, len(ids))
	for _, id := range ids {
		g := registry[id]
		out = append(out, Spec{
			ID: g.ID, Name: g.Name, Tags: append([]string(nil), g.Tags...),
			Levers: append([]Lever(nil), g.Levers...),
		})
	}
	return out
}

// deriveRule maps any of tags to lever overrides. Tag matching is
// case-insensitive and substring-based so "cozy-farming" still hits "farming".
type deriveRule struct {
	tags      []string
	overrides []Lever
}

var deriveRules = []deriveRule{
	{[]string{"gacha", "collection", "collect"}, []Lever{
		lv(MetricPayRate, "0.06", "0.02", "0.25", "1.3"),
		lv(MetricLTV, "10", "2", "35", "1.1"),
	}},
	{[]string{"casual", "hyper", "light"}, []Lever{
		lv(MetricPayRate, "0.01", "0.003", "0.05", "0.8"),
		lv(MetricRetentionD1, "0.35", "0.10", "0.60", "1.1"),
	}},
	{[]string{"mmo", "social", "guild"}, []Lever{
		lv(MetricContentCadence, "7", "3", "21", "1.0"),
		lv(MetricRetentionD30, "0.12", "0.03", "0.25", "1.2"),
	}},
	{[]string{"competitive", "esports", "pvp"}, []Lever{
		lv(MetricContentCadence, "21", "7", "60", "0.9"),
		lv(MetricPayRate, "0.02", "0.01", "0.10", "0.8"),
	}},
	{[]string{"shooter", "fps", "tps"}, []Lever{
		lv(MetricContentCadence, "21", "7", "60", "0.8"),
	}},
	{[]string{"strategy", "4x", "slg"}, []Lever{
		lv(MetricLTV, "12", "2", "40", "1.0"),
		lv(MetricRetentionD30, "0.10", "0.02", "0.20", "1.1"),
	}},
	{[]string{"card", "tcg"}, []Lever{
		lv(MetricPayRate, "0.05", "0.02", "0.20", "1.2"),
	}},
	{[]string{"sandbox", "creative", "ugc"}, []Lever{
		lv(MetricContentCadence, "30", "14", "90", "0.7"),
	}},
	{[]string{"narrative", "story", "visual"}, []Lever{
		lv(MetricPayRate, "0.04", "0.01", "0.15", "1.0"),
	}},
	{[]string{"idle", "incremental", "auto"}, []Lever{
		lv(MetricRetentionD30, "0.06", "0.01", "0.15", "1.1"),
	}},
	{[]string{"puzzle", "match", "brain"}, []Lever{
		lv(MetricContentCadence, "14", "7", "30", "1.0"),
	}},
	{[]string{"roguelike", "procedural", "permadeath"}, []Lever{
		lv(MetricContentCadence, "21", "7", "60", "0.9"),
	}},
	{[]string{"sim", "management", "economy"}, []Lever{
		lv(MetricSourceSinkRatio, "1", "4/5", "13/10", "1.1"),
	}},
	{[]string{"farming", "cozy", "life"}, []Lever{
		lv(MetricRetentionD30, "0.10", "0.02", "0.22", "1.1"),
		lv(MetricContentCadence, "30", "14", "90", "0.7"),
	}},
}

// Derive builds a genre from an ID and tags without registering it.
// Unknown tags still yield a valid genre: the universal base levers are the
// fallback, so ANY new category is immediately analysable.
func Derive(id, name string, tags []string) Genre {
	if name == "" {
		name = id
	}
	levers := baseLevers()
	joined := strings.ToLower(id + " " + strings.Join(tags, " "))
	for _, rule := range deriveRules {
		hit := false
		for _, t := range rule.tags {
			if strings.Contains(joined, strings.ToLower(t)) {
				hit = true
				break
			}
		}
		if !hit {
			continue
		}
		for _, o := range rule.overrides {
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
	}
	return Genre{ID: id, Name: name, Tags: append([]string(nil), tags...), Levers: levers}
}

// GetAny returns a registered genre, or derives one on the fly for an unknown
// ID (registering it so subsequent calls are stable). Use this when the catalog
// must accept ANY category, including ones never seen before.
func GetAny(id string) Genre {
	if g, ok := Get(id); ok {
		return g
	}
	g := Derive(id, id, splitIDTags(id))
	_ = Upsert(g)
	return g
}

// splitIDTags turns "cozy-farming-sim" into ["cozy","farming","sim"], letting
// an unseen hyphenated ID carry derivable meaning even without explicit tags.
func splitIDTags(id string) []string {
	parts := strings.FieldsFunc(id, func(r rune) bool {
		return r == '-' || r == '_' || r == ' '
	})
	return parts
}

// ValidateSpec checks a Spec can be loaded.
func ValidateSpec(s Spec) error {
	if s.ID == "" {
		return errors.New("genre: empty spec ID")
	}
	if len(s.Levers) == 0 {
		return errors.New("genre: spec has no levers: " + s.ID)
	}
	seen := map[string]bool{}
	for _, l := range s.Levers {
		if l.Key == "" {
			return errors.New("genre: empty lever key in " + s.ID)
		}
		if seen[l.Key] {
			return errors.New("genre: duplicate lever " + l.Key + " in " + s.ID)
		}
		seen[l.Key] = true
		if l.Weight.Sign() <= 0 {
			return errors.New("genre: non-positive weight " + l.Key + " in " + s.ID)
		}
		if l.Min.Cmp(l.Max) >= 0 {
			return errors.New("genre: min >= max " + l.Key + " in " + s.ID)
		}
	}
	return nil
}

// MustNum is a helper for callers building Specs programmatically.
func MustNum(s string) num.Num { return num.MustFromString(s) }
