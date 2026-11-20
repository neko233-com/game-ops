// CLI for dispatching to genre-specific strategies via the factory.
//
// Usage:
//
//	go run ./cmd/strategist list                     # all 29 registered genres
//	go run ./cmd/strategist show gacha               # metrics + advice only (no context)
//	go run ./cmd/strategist run gacha -region CN -values examples/gacha.json
//	go run ./cmd/strategist run slg -region CN -values examples/gacha-warn.json
//	go run ./cmd/strategist run card -region US -values examples/gacha.json
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/neko233-com/game-ops/factory"
	_ "github.com/neko233-com/game-ops/factory/load" // auto-register all 29 strategies
	"github.com/neko233-com/game-ops/num"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `strategist — dispatch to genre-specific strategy via factory

Usage:
  go run ./cmd/strategist list
  go run ./cmd/strategist show <genre-id>
  go run ./cmd/strategist run <genre-id> [-region CN|SEA|US|EU|GLOBAL] [-values file.json] [-unit 金币]

Registered genres:
`)
		for _, id := range factory.List() {
			fmt.Fprintf(os.Stderr, "  %s\n", id)
		}
	}
	// Top-level flag.Parse below handles the first arg ("run"|"show"|"list").
	// The "run" subcommand uses its own flag.NewFlagSet for region/values/unit.
	flag.Parse()

	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(1)
	}

	cmd := flag.Arg(0)

	switch cmd {
	case "list":
		ids := factory.List()
		fmt.Printf("Registered strategies: %d\n\n", factory.Count())
		fmt.Printf("%-22s %-10s\n", "GENRE ID", "NAME")
		fmt.Println("──────────────────────────────")
		for _, id := range ids {
			s, _ := factory.Get(id)
			fmt.Printf("%-22s %-10s\n", id, s.Name())
		}

	case "show":
		if flag.NArg() < 2 {
			fmt.Fprintf(os.Stderr, "usage: strategist show <genre>\n")
			os.Exit(1)
		}
		id := flag.Arg(1)
		s, ok := factory.Get(id)
		if !ok {
			fmt.Fprintf(os.Stderr, "unknown genre %q — see 'strategist list'\n", id)
			os.Exit(2)
		}
		fmt.Printf("Genre: %s (%s)\n\n", s.Name(), s.ID())
		metrics := s.GenreMetrics()
		if len(metrics) > 0 {
			fmt.Printf("Genre-specific metrics (%d):\n", len(metrics))
			for _, m := range metrics {
				fmt.Printf("  %-24s %s — %s\n", m.Key, m.Label, m.Desc)
			}
		} else {
			fmt.Println("(no genre-specific metrics)")
		}
		fmt.Println()
		advices := s.Advice()
		fmt.Printf("Default advice (%d):\n", len(advices))
		for _, a := range advices {
			fmt.Printf("  • %s\n", a)
		}
		// Sample Evaluate with empty context to show baseline output.
		ctx := factory.StrategyContext{GenreID: id, Region: factory.RegionGlobal}
		res, err := s.Evaluate(ctx)
		if err == nil {
			fmt.Println()
			fmt.Printf("Notes: %s\n", res.Notes)
		}

	case "run":
		if flag.NArg() < 2 {
			fmt.Fprintf(os.Stderr, "usage: strategist run <genre> [-region X] [-values file] [-unit U]\n")
			os.Exit(1)
		}
		id := flag.Arg(1)

		// Sub-flags after "run <genre>" — not parsed by the top-level flag.
		subArgs := flag.Args()[2:]
		subFS := flag.NewFlagSet("run", flag.ExitOnError)
		subFS.SetOutput(os.Stderr)
		var (
			subRegion = subFS.String("region", "GLOBAL", "CN | SEA | US | EU | GLOBAL")
			subValues = subFS.String("values", "", "JSON file of observed metric values")
			subUnit   = subFS.String("unit", "金币", "base unit label")
		)
		if err := subFS.Parse(subArgs); err != nil {
			os.Exit(2)
		}

		s, ok := factory.Get(id)
		if !ok {
			fmt.Fprintf(os.Stderr, "unknown genre %q — see 'strategist list'\n", id)
			os.Exit(2)
		}

		ctx := factory.StrategyContext{
			GenreID:  id,
			Region:   factory.Region(*subRegion),
			BaseUnit: *subUnit,
			Metrics:  map[string]num.Num{},
		}

		if *subValues != "" {
			b, err := os.ReadFile(*subValues)
			if err != nil {
				fmt.Fprintf(os.Stderr, "read: %v\n", err)
				os.Exit(2)
			}
			var raw struct {
				Values map[string]interface{} `json:"values"`
			}
			if err := json.Unmarshal(b, &raw); err != nil {
				fmt.Fprintf(os.Stderr, "parse: %v\n", err)
				os.Exit(2)
			}
			for k, v := range raw.Values {
				switch t := v.(type) {
				case string:
					n, err := num.FromString(t)
					if err != nil {
						fmt.Fprintf(os.Stderr, "bad value %s=%q: %v\n", k, t, err)
						continue
					}
					ctx.Metrics[k] = n
				case float64:
					ctx.Metrics[k] = num.MustFromString(fmt.Sprintf("%g", t))
				case json.Number:
					n, err := num.FromString(t.String())
					if err != nil {
						continue
					}
					ctx.Metrics[k] = n
				}
			}
		}

		res, err := s.Evaluate(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "evaluate: %v\n", err)
			os.Exit(3)
		}

		fmt.Printf("Genre:    %s\n", res.Genre)
		fmt.Printf("Region:   %s\n", res.Region)
		fmt.Printf("Strategy: %s\n\n", s.Name())

		// Print genre metrics.
		if len(res.GenreMetrics) > 0 {
			fmt.Printf("Genre-specific outputs (%d):\n", len(res.GenreMetrics))
			for _, m := range res.GenreMetrics {
				note := m.Note
				if note != "" {
					note = "  (" + note + ")"
				}
				fmt.Printf("  %-24s %-10s%s\n", m.Key, m.Value.DecimalString(6), note)
			}
		}

		// Risks.
		if len(res.Risks) > 0 {
			fmt.Printf("\nRisks (%d):\n", len(res.Risks))
			for _, r := range res.Risks {
				fmt.Printf("  [%s] %s: observed %s vs ideal %s — %s\n",
					r.Severity, r.Metric, r.Observed.DecimalString(6), r.Ideal.DecimalString(6), r.Message)
			}
		}

		// Advice.
		advices := s.Advice()
		if len(advices) > 0 {
			fmt.Printf("\nAdvice (%d):\n", len(advices))
			for _, a := range advices {
				fmt.Printf("  • %s\n", a)
			}
		}

		if res.Notes != "" {
			fmt.Printf("\nNotes:\n  %s\n", res.Notes)
		}

	default:
		fmt.Fprintf(os.Stderr, "unknown command %q — list | show | run\n", cmd)
		os.Exit(1)
	}
}
