// Command numgen generates additive-first combat/growth/monetization tables
// from a JSON plan (all numbers as strings) and prints validation issues.
package numgencmd

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"

	"github.com/neko233-com/game-ops/design"
	"github.com/neko233-com/game-ops/num"
)

type curveSpec struct {
	Name string `json:"name"`
	Base string `json:"base"`
	Step string `json:"step"`
	Cap  string `json:"cap"`
}

func (c curveSpec) curve() design.Curve {
	name := c.Name
	if name == "" {
		name = "curve"
	}
	return design.Curve{
		Name: name, Base: num.MustFromString(c.Base),
		Step: num.MustFromString(c.Step), Cap: num.MustFromString(c.Cap),
	}
}

type planSpec struct {
	Name           string    `json:"name"`
	Growth         curveSpec `json:"growth"`
	CombatPower    curveSpec `json:"combatPower"`
	Damage         curveSpec `json:"damage"`
	HP             curveSpec `json:"hp"`
	Reward         curveSpec `json:"reward"`
	Price          curveSpec `json:"price"`
	Levels         uint      `json:"levels"`
	MaxGrowthRatio string    `json:"maxGrowthRatio"`
}

func Main() {
	formulaFlag := flag.String("formula", "", "raw growth formula to check for forbidden global exponents")
	flag.Parse()
	args := flag.Args()
	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, "usage: numgen [-formula EXPR] <plan.json>")
		os.Exit(2)
	}
	data, err := os.ReadFile(args[0])
	if err != nil {
		fmt.Fprintln(os.Stderr, "read plan:", err)
		os.Exit(1)
	}
	var s planSpec
	if err := json.Unmarshal(data, &s); err != nil {
		fmt.Fprintln(os.Stderr, "parse plan:", err)
		os.Exit(1)
	}
	plan := design.Plan{
		Name:   s.Name,
		Growth: s.Growth.curve(), CombatPower: s.CombatPower.curve(),
		Damage: s.Damage.curve(), HP: s.HP.curve(),
		Reward: s.Reward.curve(), Price: s.Price.curve(),
	}
	ratio := num.MustFromString("1.5")
	if s.MaxGrowthRatio != "" {
		ratio = num.MustFromString(s.MaxGrowthRatio)
	}
	if s.Levels == 0 {
		s.Levels = 10
	}

	fmt.Printf("Plan: %s (additive-first)\n\n", plan.Name)
	fmt.Printf("LEVEL  %s\n", "  growth  combatPower  damage  hp  reward  price")
	for level := uint(1); level <= s.Levels; level++ {
		row := plan.At(level)
		keys := make([]string, 0, len(row))
		for k := range row {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		fmt.Printf("%5d", level)
		for _, k := range keys {
			fmt.Printf("  %s=%s", k, row[k].DecimalString(4))
		}
		fmt.Println()
	}

	issues := plan.Validate(ratio)
	issues = append(issues, design.RejectGlobalExponent(*formulaFlag)...)
	fmt.Printf("\nValidation (%d issues):\n", len(issues))
	if len(issues) == 0 {
		fmt.Println("  OK - all curves additive, monotonic and bounded")
	}
	for _, issue := range issues {
		fmt.Printf("  [%s] %s: %s\n", issue.Severity, issue.Field, issue.Message)
	}
	if len(issues) > 0 {
		os.Exit(1)
	}
}
