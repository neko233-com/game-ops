// Command opsreport reads a JSON config (numeric values as strings, so they
// keep arbitrary precision) and prints a long-term-operation health report.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/neko233-com/game-ops/num"
	"github.com/neko233-com/game-ops/ops"
)

type config struct {
	Genre  string            `json:"genre"`
	Values map[string]string `json:"values"`
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: opsreport <config.json>")
		os.Exit(2)
	}
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, "read config:", err)
		os.Exit(1)
	}
	var cfg config
	if err := json.Unmarshal(data, &cfg); err != nil {
		fmt.Fprintln(os.Stderr, "parse config:", err)
		os.Exit(1)
	}

	values := make(map[string]num.Num, len(cfg.Values))
	for key, raw := range cfg.Values {
		n, err := num.FromString(raw)
		if err != nil {
			fmt.Fprintf(os.Stderr, "value %q: %v\n", key, err)
			os.Exit(1)
		}
		values[key] = n
	}

	report, err := ops.Evaluate(ops.Snapshot{GenreID: cfg.Genre, Values: values})
	if err != nil {
		fmt.Fprintln(os.Stderr, "evaluate:", err)
		os.Exit(1)
	}

	sort.Slice(report.Diagnostics, func(i, j int) bool {
		return report.Diagnostics[i].Metric < report.Diagnostics[j].Metric
	})

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "Genre:\t%s (%s)\n", report.GenreName, report.GenreID)
	fmt.Fprintf(w, "Score:\t%s\n", report.Score.DecimalString(6))
	fmt.Fprintf(w, "Verdict:\t%s\n", report.Verdict)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "METRIC\tOBSERVED\tIDEAL\tFITNESS\tWEIGHT\tSTATUS")
	for _, d := range report.Diagnostics {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			d.Metric, d.Observed.DecimalString(6), d.Ideal.DecimalString(6),
			d.Fitness.DecimalString(4), d.Weight.DecimalString(2), d.Status)
	}
	if err := w.Flush(); err != nil {
		fmt.Fprintln(os.Stderr, "write report:", err)
		os.Exit(1)
	}
}
