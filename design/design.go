// Package design implements additive-first numerical design for combat, growth
// and monetization. Curves are base + step*(level-1) with an optional hard cap;
// multiplication is deliberately avoided so marginal value stays predictable.
package design

import (
	"fmt"
	"strings"

	"github.com/neko233-com/game-ops/num"
)

var (
	zero = num.Zero()
	one  = num.One()
)

// Curve is an additive curve with an optional cap (cap == 0 means no cap).
type Curve struct {
	Name string
	Base num.Num
	Step num.Num
	Cap  num.Num
}

// AdditiveCurve builds a Curve from string parameters (config-friendly).
func AdditiveCurve(name, base, step, cap string) Curve {
	return Curve{
		Name: name, Base: num.MustFromString(base),
		Step: num.MustFromString(step), Cap: num.MustFromString(cap),
	}
}

// At returns the value at level (levels below 1 are treated as 1).
func (c Curve) At(level uint) num.Num {
	if level < 1 {
		level = 1
	}
	v := c.Base.Add(c.Step.Mul(num.FromInt64(int64(level - 1))))
	if c.Cap.Sign() > 0 && v.Cmp(c.Cap) > 0 {
		return c.Cap
	}
	return v
}

// AtMany returns values at every level in [lo, hi] inclusive.
func (c Curve) AtMany(lo, hi uint) []num.Num {
	if lo > hi {
		return nil
	}
	out := make([]num.Num, hi-lo+1)
	for i := lo; i <= hi; i++ {
		out[i-lo] = c.At(i)
	}
	return out
}

// Total returns Σ_{l=1}^{levels} c.At(l) — the accumulated value over
// a range of levels. Levels below 1 returns 0.
func (c Curve) Total(levels uint) num.Num {
	if levels == 0 {
		return zero
	}
	total := zero
	for l := uint(1); l <= levels; l++ {
		total = total.Add(c.At(l))
	}
	return total
}

// PieceSegment is one additive band in a piecewise curve.
// A piecewise curve is the ONLY legitimate replacement for a global
// exponent: each band is additive within its range, and the boundaries
// are where design intent changes.
type PieceSegment struct {
	Start uint
	End   uint
	Base  num.Num
	Step  num.Num
}

// PiecewiseCurve is a sequence of additive bands. Gaps or overlaps are
// detected at build time. No piece may be empty.
type PiecewiseCurve struct {
	Name     string
	Segments []PieceSegment
	Cap      num.Num
}

// AdditivePiecewise wraps one or more additive curves as a piecewise curve.
// Each Curve segment must not be capped (Cap==0) — the outer Cap handles that.
func AdditivePiecewise(name string, cap string, segs ...PieceSegment) PiecewiseCurve {
	return PiecewiseCurve{
		Name:     name,
		Cap:      num.MustFromString(cap),
		Segments: segs,
	}
}

// PieceBuilder builds PieceSegments from string parameters fluently.
func Piece(start, end uint, base, step string) PieceSegment {
	return PieceSegment{
		Start: start, End: end,
		Base: num.MustFromString(base), Step: num.MustFromString(step),
	}
}

// At evaluates the piecewise curve at level. Levels before the first segment
// are extrapolated linearly from the first segment's base/step. Levels after
// the last segment use the last segment's step from its final base.
func (p PiecewiseCurve) At(level uint) num.Num {
	if len(p.Segments) == 0 {
		return zero
	}
	if level < p.Segments[0].Start {
		seg := p.Segments[0]
		// Extrapolate linearly backward using seg.Base as value at seg.Start.
		v := seg.Base.Sub(seg.Step.Mul(num.FromInt64(int64(seg.Start - level))))
		return clamp(v, p.Cap)
	}
	var v num.Num
	for _, seg := range p.Segments {
		if level >= seg.Start && level <= seg.End {
			offset := num.FromInt64(int64(level - seg.Start))
			v = seg.Base.Add(seg.Step.Mul(offset))
			return clamp(v, p.Cap)
		}
	}
	// Past last segment: use last segment's final base + last segment step.
	last := p.Segments[len(p.Segments)-1]
	offset := num.FromInt64(int64(level - last.End))
	final := last.Base.Add(last.Step.Mul(num.FromInt64(int64(last.End - last.Start))))
	v = final.Add(last.Step.Mul(offset))
	return clamp(v, p.Cap)
}

// ValidatePiece checks that segment ranges partition [1,cap] cleanly and
// that each segment's base/step is non-negative.
func (p PiecewiseCurve) ValidatePiece() []Issue {
	var issues []Issue
	if len(p.Segments) == 0 {
		return []Issue{{"error", p.Name, "piecewise curve has no segments"}}
	}
	for i, seg := range p.Segments {
		if seg.End < seg.Start {
			issues = append(issues, Issue{"error", p.Name,
				fmt.Sprintf("segment %d end (%d) < start (%d)", i, seg.End, seg.Start)})
		}
		if seg.Base.Sign() <= 0 {
			issues = append(issues, Issue{"error", p.Name,
				fmt.Sprintf("segment %d base must be positive", i)})
		}
		if seg.Step.Sign() < 0 {
			issues = append(issues, Issue{"error", p.Name,
				fmt.Sprintf("segment %d step is negative", i)})
		}
	}
	// Check adjacency.
	for i := 1; i < len(p.Segments); i++ {
		prev, cur := p.Segments[i-1], p.Segments[i]
		if cur.Start != prev.End+1 {
			issues = append(issues, Issue{"warn", p.Name,
				fmt.Sprintf("gap or overlap between segment %d end %d and segment %d start %d",
					i-1, prev.End, i, cur.Start)})
		}
	}
	return issues
}

func clamp(v, cap num.Num) num.Num {
	if cap.Sign() > 0 && v.Cmp(cap) > 0 {
		return cap
	}
	return v
}

// Plan bundles the standard curves used across combat, growth and monetization.
type Plan struct {
	Name        string
	Growth      Curve
	CombatPower Curve
	Damage      Curve
	HP          Curve
	Reward      Curve
	Price       Curve
}

func (p Plan) curves() map[string]Curve {
	return map[string]Curve{
		"growth": p.Growth, "combatPower": p.CombatPower, "damage": p.Damage,
		"hp": p.HP, "reward": p.Reward, "price": p.Price,
	}
}

// At returns every curve's value at a level.
func (p Plan) At(level uint) map[string]num.Num {
	out := make(map[string]num.Num)
	for name, c := range p.curves() {
		out[name] = c.At(level)
	}
	return out
}

// Issue is a validation finding. Severity is "error" or "info".
type Issue struct {
	Severity string
	Field    string
	Message  string
}

// RejectGlobalExponent flags a raw formula that uses a global exponent.
// Even a capped exponential is forbidden: the compounding happens long before
// the cap, so the player-visible curve is still explosive. Use additive curves
// with piecewise bands and hard caps instead.
func RejectGlobalExponent(expr string) []Issue {
	lower := strings.ToLower(expr)
	for _, token := range []string{"^", "**", "pow(", "math.pow", "expm1"} {
		if strings.Contains(lower, token) {
			return []Issue{{
				Severity: "error", Field: "formula",
				Message: "global exponent is forbidden even with a cap (" + token +
					"); use an additive curve base + step*(n-1) with a hard cap and piecewise bands",
			}}
		}
	}
	return nil
}

// Validate enforces additive-first discipline. maxGrowthRatio bounds the
// adjacent-level ratio; ratios above it signal multiplicative explosion.
func (p Plan) Validate(maxGrowthRatio num.Num) []Issue {
	var issues []Issue
	if maxGrowthRatio.Sign() <= 0 {
		maxGrowthRatio = num.MustFromString("3/2")
	}
	for name, c := range p.curves() {
		if c.Name == "" {
			continue // unset curve in a partial plan
		}
		if c.Base.Sign() <= 0 {
			issues = append(issues, Issue{"error", name, "base must be positive"})
		}
		if c.Step.Sign() < 0 {
			issues = append(issues, Issue{"error", name, "negative step breaks monotonicity"})
		}
		if c.Cap.Sign() > 0 && c.Cap.Cmp(c.Base) < 0 {
			issues = append(issues, Issue{"error", name, "cap below base"})
		}
		if c.Base.Sign() > 0 && c.Step.Sign() > 0 {
			// Worst adjacent ratio is at level 1: (base+step)/base.
			first := c.At(1)
			second := c.At(2)
			ratio, err := second.Div(first)
			if err == nil && ratio.Cmp(maxGrowthRatio) > 0 {
				issues = append(issues, Issue{
					"error", name,
					fmt.Sprintf("adjacent-level ratio %s exceeds cap %s (additive step too large)", ratio.String(), maxGrowthRatio.String()),
				})
			}
		}
	}
	return issues
}
