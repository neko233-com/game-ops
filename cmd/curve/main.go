// CLI for piecewise additive curves — the ONLY legitimate replacement for
// a global exponent. Piecewise curves keep every segment additive, so
// marginal value stays predictable while the total shape adapts to design.
//
// Usage:
//
//	# from JSON (see examples/piecewise.json)
//	go run ./cmd/curve examples/piecewise.json
//
//	# inline (overrides JSON):
//	go run ./cmd/curve -name "xp" -cap 10000 -maxlevel 60 \
//	    -seg "1,10,100,10" -seg "11,30,200,15" -seg "31,60,500,20"
//
//	# built-in "example" mode — prints a recommended default piecewise curve
//	go run ./cmd/curve example
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/neko233-com/game-ops/design"
	"github.com/neko233-com/game-ops/num"
)

type piecewiseInput struct {
	Name     string         `json:"name"`
	Cap      string         `json:"cap"`
	MaxLevel uint           `json:"maxLevel"`
	Segments []segmentInput `json:"segments"`
	Formulas []string       `json:"formulas,omitempty"` // guard formulas to RejectGlobalExponent
}

type segmentInput struct {
	Start uint   `json:"start"`
	End   uint   `json:"end"`
	Base  string `json:"base"`
	Step  string `json:"step"`
}

func main() {
	var (
		name     = flag.String("name", "", "piecewise name")
		cap      = flag.String("cap", "", "hard cap (0=no cap)")
		maxLevel = flag.Uint("maxlevel", 60, "max level to print")
		segs     = &segList{}
	)
	flag.Var(segs, "seg", "segment start,end,base,step (repeatable)")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "curve — piecewise additive curve generator\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n  go run ./cmd/curve <piecewise.json>\n  go run ./cmd/curve example\n  go run ./cmd/curve -name xp -cap 10000 -maxlevel 60 -seg 1,10,100,10 -seg 11,30,200,15\n")
	}
	flag.Parse()

	input := piecewiseInput{}

	// Mode 1: positional arg is file or "example".
	arg := ""
	if flag.NArg() > 0 {
		arg = flag.Arg(0)
	}

	switch {
	case arg == "example":
		input = piecewiseInput{
			Name:     "xp",
			Cap:      "10000",
			MaxLevel: 60,
			Segments: []segmentInput{
				{1, 10, "100", "10"},
				{11, 30, "200", "15"},
				{31, 60, "500", "20"},
			},
		}
	case arg != "":
		b, err := os.ReadFile(arg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "read: %v\n", err)
			os.Exit(2)
		}
		if err := json.Unmarshal(b, &input); err != nil {
			fmt.Fprintf(os.Stderr, "parse: %v\n", err)
			os.Exit(2)
		}
	default:
		// Flags mode.
		if *name == "" || *cap == "" || len(segs.vals) == 0 {
			fmt.Fprintf(os.Stderr, "need either a JSON file, 'example', or -name/-cap/-seg flags\n")
			flag.Usage()
			os.Exit(1)
		}
		input.Name = *name
		input.Cap = *cap
		input.MaxLevel = *maxLevel
		for _, s := range segs.vals {
			parts := strings.Split(s, ",")
			if len(parts) != 4 {
				fmt.Fprintf(os.Stderr, "-seg expects start,end,base,step — got %q\n", s)
				os.Exit(2)
			}
			var seg segmentInput
			fmt.Sscanf(parts[0], "%d", &seg.Start)
			fmt.Sscanf(parts[1], "%d", &seg.End)
			seg.Base = parts[2]
			seg.Step = parts[3]
			input.Segments = append(input.Segments, seg)
		}
	}

	if input.MaxLevel == 0 {
		input.MaxLevel = 60
	}
	if input.Cap == "" {
		input.Cap = "0"
	}
	if input.Name == "" {
		input.Name = "piecewise"
	}

	// Guard formulas first, if any.
	for _, f := range input.Formulas {
		for _, issue := range design.RejectGlobalExponent(f) {
			fmt.Fprintf(os.Stderr, "[rejected] %s: %s\n  formula: %s\n", issue.Severity, issue.Message, f)
			os.Exit(2)
		}
	}

	// Build curve.
	segsGo := make([]design.PieceSegment, 0, len(input.Segments))
	for _, s := range input.Segments {
		segsGo = append(segsGo, design.Piece(s.Start, s.End, s.Base, s.Step))
	}
	curve := design.AdditivePiecewise(input.Name, input.Cap, segsGo...)

	// Validate.
	if issues := curve.ValidatePiece(); len(issues) > 0 {
		for _, issue := range issues {
			fmt.Fprintf(os.Stderr, "[%s] %s: %s\n", issue.Severity, issue.Field, issue.Message)
		}
		for _, issue := range issues {
			if issue.Severity == "error" {
				os.Exit(3)
			}
		}
	}

	// Print.
	fmt.Printf("Piecewise curve: %s (cap=%s, max=%d)\n", curve.Name, curve.Cap.String(), input.MaxLevel)
	fmt.Printf("  %d segments\n", len(curve.Segments))
	for i, seg := range curve.Segments {
		fmt.Printf("  [%d] levels %d–%d : base=%s step=%s\n",
			i, seg.Start, seg.End, seg.Base.String(), seg.Step.String())
	}
	fmt.Println()
	fmt.Printf("%-7s %-12s %-16s\n", "LEVEL", "VALUE", "CUMULATIVE")
	fmt.Println(strRepeat("─", 36))
	cum := num.Zero()
	for l := uint(1); l <= input.MaxLevel; l++ {
		v := curve.At(l)
		cum = cum.Add(v)
		fmt.Printf("%-7d %-12s %-16s\n", l, v.String(), cum.String())
	}

	fmt.Println()
	fmt.Printf("Total (levels 1–%d): %s (≈ %s)\n", input.MaxLevel, cum.String(), cum.DecimalString(6))
	fmt.Println()
	fmt.Println("Why piecewise instead of global exponent?")
	fmt.Println("  A global exponent base·1.1ⁿ compounds to ~142× by level 50.")
	fmt.Println("  Each additive band keeps marginal gain constant within that band,")
	fmt.Println("  while segment boundaries reflect real design intent (milestones,")
	fmt.Println("  paywalls, content bursts). Every boundary is explicit, verifiable,")
	fmt.Println("  and rebalanceable without touching any other level.")
}

func strRepeat(s string, n int) string {
	out := ""
	for i := 0; i < n; i++ {
		out += s
	}
	return out
}

// segList implements flag.Value so -seg can repeat.
type segList struct{ vals []string }

func (s *segList) String() string     { return strings.Join(s.vals, ",") }
func (s *segList) Set(v string) error { s.vals = append(s.vals, v); return nil }
