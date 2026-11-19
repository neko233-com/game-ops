// Command econcheck computes the source/sink balance from exact string values.
// A ratio near 1 means the economy is balanced; >1 means net injection
// (inflation risk), <1 means scarcity (progression stalls).
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/neko233/game-ops/num"
)

type spec struct {
	Sources map[string]string `json:"sources"`
	Sinks   map[string]string `json:"sinks"`
}

func total(m map[string]string) num.Num {
	t := num.Zero()
	for _, v := range m {
		t = t.Add(num.MustFromString(v))
	}
	return t
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: econcheck <spec.json>")
		os.Exit(2)
	}
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, "read spec:", err)
		os.Exit(1)
	}
	var s spec
	if err := json.Unmarshal(data, &s); err != nil {
		fmt.Fprintln(os.Stderr, "parse spec:", err)
		os.Exit(1)
	}
	sources, sinks := total(s.Sources), total(s.Sinks)
	fmt.Printf("Sources: %s (%d entries)\n", sources.String(), len(s.Sources))
	fmt.Printf("Sinks:   %s (%d entries)\n", sinks.String(), len(s.Sinks))
	if sinks.IsZero() {
		fmt.Fprintln(os.Stderr, "no sinks: every source inflates the economy")
		os.Exit(1)
	}
	ratio, err := sources.Div(sinks)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ratio:", err)
		os.Exit(1)
	}
	fmt.Printf("Source/sink ratio: %s (%.4f)\n", ratio.String(), ratio.Float64())
	low, high := num.MustFromString("9/10"), num.MustFromString("11/10")
	switch {
	case ratio.Cmp(high) > 0:
		fmt.Println("Verdict: INFLATION - net value injection; add sinks or reduce sources")
	case ratio.Cmp(low) < 0:
		fmt.Println("Verdict: SCARCITY - progression will stall; add sources or reduce sinks")
	default:
		fmt.Println("Verdict: BALANCED")
	}
}
