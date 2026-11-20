// CLI for the OPEN genre catalog: derive, register, evaluate and export ANY
// game category at runtime — no code generation, no rebuild.
//
// Usage:
//
//	go run ./cmd/catalog list
//	go run ./cmd/catalog derive <id> [-name N] [-tags a,b,c] [-save file.json]
//	go run ./cmd/catalog register <file.json>        # array of genre Specs
//	go run ./cmd/catalog eval <id> [-values file.json]
//	go run ./cmd/catalog strategy <id> [-values file.json] [-region CN]
//	go run ./cmd/catalog export [-out catalog.json]
package catalogcmd

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/neko233-com/game-ops/factory"
	_ "github.com/neko233-com/game-ops/factory/load"
	"github.com/neko233-com/game-ops/genre"
	"github.com/neko233-com/game-ops/num"
	"github.com/neko233-com/game-ops/ops"
)

func Main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}
	cmd := os.Args[1]
	args := os.Args[2:]
	switch cmd {
	case "list":
		cmdList(args)
	case "derive":
		cmdDerive(args)
	case "register":
		cmdRegister(args)
	case "eval":
		cmdEval(args)
	case "strategy":
		cmdStrategy(args)
	case "export":
		cmdExport(args)
	case "help", "-h", "--help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n", cmd)
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprint(os.Stderr, `catalog — open, data-driven game genre catalog

Usage:
  catalog list
  catalog derive <id> [-name "名字"] [-tags gacha,casual] [-save file.json]
  catalog register <file.json>          # Spec[] or single Spec
  catalog eval <id> [-values file.json] # health report (auto-derives unknown id)
  catalog strategy <id> [-values file.json] [-region CN|SEA|US|EU|GLOBAL]
  catalog export [-out file.json]

Any category is supported: unknown IDs are auto-derived from their name/tags
using the universal 11 levers and 14 tag rules. Register a JSON spec to override.
`)
}

func cmdList(args []string) {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	fs.Parse(args)
	ids := genre.List()
	fmt.Printf("Registered genres (%d):\n\n", len(ids))
	fmt.Printf("%-24s %-34s %s\n", "ID", "NAME", "TAGS")
	fmt.Println(strings.Repeat("─", 88))
	for _, id := range ids {
		g := genre.MustGet(id)
		fmt.Printf("%-24s %-34s %s\n", id, g.Name, strings.Join(g.Tags, ", "))
	}
}

// reorderFlags moves flag tokens before positional args so Go's flag package
// (which stops at the first non-flag) still parses "catalog derive <id> -tags x".
func reorderFlags(args []string) []string {
	flags, positional := []string{}, []string{}
	for i := 0; i < len(args); i++ {
		a := args[i]
		if strings.HasPrefix(a, "-") {
			flags = append(flags, a)
			if !strings.Contains(a, "=") && i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				i++
				flags = append(flags, args[i])
			}
		} else {
			positional = append(positional, a)
		}
	}
	return append(flags, positional...)
}

func cmdDerive(rawArgs []string) {
	args := reorderFlags(rawArgs)
	fs := flag.NewFlagSet("derive", flag.ExitOnError)
	name := fs.String("name", "", "human-readable name")
	tags := fs.String("tags", "", "comma-separated tags")
	save := fs.String("save", "", "save derived spec to JSON file")
	fs.Parse(args)
	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "usage: catalog derive <id> [-name N] [-tags a,b] [-save f.json]")
		os.Exit(1)
	}
	id := fs.Arg(0)
	tagList := []string{}
	if *tags != "" {
		for _, t := range strings.Split(*tags, ",") {
			t = strings.TrimSpace(t)
			if t != "" {
				tagList = append(tagList, t)
			}
		}
	}
	g := genre.Derive(id, *name, tagList)
	if err := genre.Upsert(g); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("Derived & registered: %s (%s)\nTags: %s\n\n", g.ID, g.Name, strings.Join(g.Tags, ", "))
	printLevers(g)
	if *save != "" {
		writeSpec(*save, []genre.Spec{specOf(g)})
		fmt.Printf("\nSaved to %s\n", *save)
	}
}

func cmdRegister(args []string) {
	fs := flag.NewFlagSet("register", flag.ExitOnError)
	fs.Parse(args)
	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "usage: catalog register <file.json>")
		os.Exit(1)
	}
	path := fs.Arg(0)
	b, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	var specs []genre.Spec
	if err := json.Unmarshal(b, &specs); err != nil {
		var one genre.Spec
		if err2 := json.Unmarshal(b, &one); err2 != nil {
			fmt.Fprintf(os.Stderr, "parse: %v\n", err)
			os.Exit(1)
		}
		specs = []genre.Spec{one}
	}
	for _, s := range specs {
		if err := genre.ValidateSpec(s); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
	n, err := genre.LoadSpecs(specs)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("Registered %d genre(s) from %s\n", n, path)
	for _, s := range specs {
		fmt.Printf("  %-24s %s\n", s.ID, s.Name)
	}
}

func cmdEval(rawArgs []string) {
	args := reorderFlags(rawArgs)
	fs := flag.NewFlagSet("eval", flag.ExitOnError)
	values := fs.String("values", "", "JSON file of observed metric values")
	fs.Parse(args)
	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "usage: catalog eval <id> [-values file.json]")
		os.Exit(1)
	}
	id := fs.Arg(0)
	vals, err := readValues(*values)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	r, err := ops.EvaluateAny(ops.Snapshot{GenreID: id, Values: vals})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("Genre:    %s (%s)\n", r.GenreName, r.GenreID)
	fmt.Printf("Score:    %s\n", r.Score.DecimalString(6))
	fmt.Printf("Verdict:  %s\n\n", r.Verdict)
	fmt.Printf("%-30s %-12s %-12s %-10s %-10s %s\n", "METRIC", "OBSERVED", "IDEAL", "FITNESS", "WEIGHT", "STATUS")
	fmt.Println(strings.Repeat("─", 92))
	for _, d := range r.Diagnostics {
		fmt.Printf("%-30s %-12s %-12s %-10s %-10s %s\n",
			d.Metric, d.Observed.DecimalString(6), d.Ideal.DecimalString(6),
			d.Fitness.DecimalString(4), d.Weight.String(), d.Status)
	}
}

func cmdStrategy(rawArgs []string) {
	args := reorderFlags(rawArgs)
	fs := flag.NewFlagSet("strategy", flag.ExitOnError)
	values := fs.String("values", "", "JSON file of observed metric values")
	region := fs.String("region", "GLOBAL", "CN | SEA | US | EU | GLOBAL")
	fs.Parse(args)
	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "usage: catalog strategy <id> [-values f.json] [-region CN]")
		os.Exit(1)
	}
	id := fs.Arg(0)
	vals, err := readValues(*values)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	res, err := factory.RunAny(factory.StrategyContext{
		GenreID: id, Region: factory.Region(*region), Metrics: vals,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	s := factory.GetAny(id)
	fmt.Printf("Genre:    %s\n", res.Genre)
	fmt.Printf("Region:   %s\n", res.Region)
	fmt.Printf("Strategy: %s\n\n", s.Name())
	if len(res.GenreMetrics) > 0 {
		fmt.Printf("Genre metrics (%d):\n", len(res.GenreMetrics))
		for _, m := range res.GenreMetrics {
			fmt.Printf("  %-28s %s\n", m.Key, m.Value.DecimalString(6))
		}
	}
	if len(res.Risks) > 0 {
		fmt.Printf("\nRisks (%d):\n", len(res.Risks))
		for _, r := range res.Risks {
			fmt.Printf("  [%s] %s: %s\n", r.Severity, r.Metric, r.Message)
		}
	}
	if res.Notes != "" {
		fmt.Printf("\nNotes: %s\n", res.Notes)
	}
}

func cmdExport(args []string) {
	fs := flag.NewFlagSet("export", flag.ExitOnError)
	out := fs.String("out", "", "write to file instead of stdout")
	fs.Parse(args)
	b, err := json.MarshalIndent(genre.Export(), "", "  ")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if *out == "" {
		fmt.Println(string(b))
		return
	}
	if err := os.WriteFile(*out, b, 0644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("Exported %d genres to %s\n", len(genre.Export()), *out)
}

func printLevers(g genre.Genre) {
	fmt.Printf("%-30s %-12s %-12s %-12s %s\n", "METRIC", "IDEAL", "MIN", "MAX", "WEIGHT")
	fmt.Println(strings.Repeat("─", 76))
	for _, l := range g.Levers {
		fmt.Printf("%-30s %-12s %-12s %-12s %s\n",
			l.Key, l.Ideal.String(), l.Min.String(), l.Max.String(), l.Weight.String())
	}
}

func specOf(g genre.Genre) genre.Spec {
	return genre.Spec{ID: g.ID, Name: g.Name, Tags: g.Tags, Levers: g.Levers}
}

func writeSpec(path string, specs []genre.Spec) {
	b, err := json.MarshalIndent(specs, "", "  ")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := os.WriteFile(path, b, 0644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func readValues(path string) (map[string]num.Num, error) {
	out := map[string]num.Num{}
	if path == "" {
		return out, nil
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var raw struct {
		Values map[string]interface{} `json:"values"`
	}
	if err := json.Unmarshal(b, &raw); err != nil {
		return nil, err
	}
	for k, v := range raw.Values {
		switch t := v.(type) {
		case string:
			n, err := num.FromString(t)
			if err != nil {
				continue
			}
			out[k] = n
		case float64:
			out[k] = num.MustFromString(fmt.Sprintf("%g", t))
		case json.Number:
			if n, err := num.FromString(t.String()); err == nil {
				out[k] = n
			}
		}
	}
	return out, nil
}
