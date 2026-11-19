// Package normalize maps game resources to a single base unit with exact
// rational exchange rates. This is the "单位价值/归一化" foundation: rewards,
// prices and combat power become comparable only after normalization.
package normalize

import (
	"errors"
	"fmt"

	"github.com/neko233/game-ops/num"
)

// Unit is a kind of value in the game economy.
type Unit struct {
	ID     string
	Symbol string
	Desc   string
}

// Rate states that 1 From is worth Value units of To.
type Rate struct {
	From  string
	To    string
	Value num.Num
}

// Ledger converts between units using exact rates and a transitive graph.
type Ledger struct {
	base  string
	units map[string]Unit
	edges map[string]map[string]num.Num
}

// NewLedger validates inputs and builds a bidirectional conversion graph.
func NewLedger(base string, units []Unit, rates []Rate) (*Ledger, error) {
	l := &Ledger{
		base: base, units: map[string]Unit{}, edges: map[string]map[string]num.Num{},
	}
	for _, u := range units {
		if u.ID == "" {
			return nil, errors.New("normalize: empty unit id")
		}
		if _, exists := l.units[u.ID]; exists {
			return nil, errors.New("normalize: duplicate unit: " + u.ID)
		}
		l.units[u.ID] = u
		l.edges[u.ID] = map[string]num.Num{}
	}
	if _, ok := l.units[base]; !ok {
		return nil, errors.New("normalize: base unit not registered: " + base)
	}
	for _, r := range rates {
		if r.From == r.To {
			return nil, errors.New("normalize: self rate: " + r.From)
		}
		if _, ok := l.units[r.From]; !ok {
			return nil, errors.New("normalize: unknown rate source: " + r.From)
		}
		if _, ok := l.units[r.To]; !ok {
			return nil, errors.New("normalize: unknown rate target: " + r.To)
		}
		if r.Value.Sign() <= 0 {
			return nil, errors.New("normalize: non-positive rate for " + r.From + "->" + r.To)
		}
		inverse, err := num.One().Div(r.Value)
		if err != nil {
			return nil, err
		}
		l.edges[r.From][r.To] = r.Value
		l.edges[r.To][r.From] = inverse
	}
	return l, nil
}

// Base returns the base unit id.
func (l *Ledger) Base() string { return l.base }

// Convert converts amount from one unit to another using the shortest path.
// All arithmetic is exact; chained rates are multiplied, never rounded.
func (l *Ledger) Convert(from, to string, amount num.Num) (num.Num, error) {
	if _, ok := l.units[from]; !ok {
		return num.Num{}, fmt.Errorf("normalize: unknown unit %q", from)
	}
	if _, ok := l.units[to]; !ok {
		return num.Num{}, fmt.Errorf("normalize: unknown unit %q", to)
	}
	if from == to {
		return amount, nil
	}
	// BFS over the rate graph, tracking the accumulated exact multiplier.
	type node struct {
		id     string
		factor num.Num
	}
	queue := []node{{id: from, factor: num.One()}}
	seen := map[string]bool{from: true}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for next, rate := range l.edges[cur.id] {
			if seen[next] {
				continue
			}
			factor := cur.factor.Mul(rate)
			if next == to {
				return amount.Mul(factor), nil
			}
			seen[next] = true
			queue = append(queue, node{id: next, factor: factor})
		}
	}
	return num.Num{}, fmt.Errorf("normalize: no conversion path %q -> %q", from, to)
}

// Normalize converts amount into the ledger's base unit.
func (l *Ledger) Normalize(from string, amount num.Num) (num.Num, error) {
	return l.Convert(from, l.base, amount)
}
