// Package design implements additive-first numerical design for combat, growth
// and monetization. Curves are base + step*(level-1) with an optional hard cap;
// multiplication is deliberately avoided so marginal value stays predictable.
package design

import (
	"fmt"
	"strings"

	"github.com/neko233/game-ops/num"
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
