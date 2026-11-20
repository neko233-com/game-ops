// CLI for querying built-in genre levers and registering custom ones.
// Usage:
//
//	go run ./cmd/genre list              # all built-in genre IDs
//	go run ./cmd/genre show gacha        # levers for one genre
//	go run ./cmd/genre metrics           # 11 universal metric definitions
package genrecmd

import (
	"flag"
	"fmt"
	"os"
	"sort"

	"github.com/neko233-com/game-ops/genre"
)

func Main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "genre — query built-in genre registries\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n  go run ./cmd/genre list\n  go run ./cmd/genre show <id>\n  go run ./cmd/genre metrics\n")
	}
	flag.Parse()

	cmd := "list"
	if flag.NArg() > 0 {
		cmd = flag.Arg(0)
	}

	switch cmd {
	case "list":
		ids := genre.List()
		fmt.Printf("Built-in genres (%d):\n\n", len(ids))
		for _, id := range ids {
			g := genre.MustGet(id)
			tags := ""
			if len(g.Tags) > 0 {
				tags = " [" + join(g.Tags, ", ") + "]"
			}
			fmt.Printf("  %-22s %s%s\n", id, g.Name, tags)
		}
	case "show":
		if flag.NArg() < 2 {
			fmt.Fprintf(os.Stderr, "usage: genre show <id>\n")
			os.Exit(1)
		}
		g, ok := genre.Get(flag.Arg(1))
		if !ok {
			fmt.Fprintf(os.Stderr, "unknown genre %q; try 'genre list'\n", flag.Arg(1))
			os.Exit(2)
		}
		fmt.Printf("Genre: %s (%s)\nTags:  %s\n\n", g.Name, g.ID, join(g.Tags, ", "))
		fmt.Printf("%-30s %-10s %-10s %-10s %-10s %-8s\n",
			"METRIC", "IDEAL", "MIN", "MAX", "WEIGHT", "STATUS")
		fmt.Println(strRepeat("─", 82))
		keys := make([]string, 0, len(g.Levers))
		for _, l := range g.Levers {
			keys = append(keys, l.Key)
		}
		sort.Strings(keys)
		for _, k := range keys {
			lv, _ := genre.LeverFor(g, k)
			fmt.Printf("%-30s %-10s %-10s %-10s %-10s %-8s\n",
				k, lv.Ideal.String(), lv.Min.String(), lv.Max.String(), lv.Weight.String(), metricStatus(lv))
		}
	case "metrics":
		meta := genre.Metrics()
		fmt.Printf("Universal metrics (%d):\n\n", len(meta))
		keys := make([]string, 0, len(meta))
		for k := range meta {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			m := meta[k]
			fmt.Printf("  %s (%s) — %s\n", k, m.Unit, m.Desc)
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown subcommand: %q (list | show | metrics)\n", cmd)
		os.Exit(1)
	}
}

func metricStatus(l genre.Lever) string {
	_ = l
	return "—" // no observed yet; status determined after Evaluate
}

func join(ss []string, sep string) string {
	if len(ss) == 0 {
		return "(none)"
	}
	out := ss[0]
	for _, s := range ss[1:] {
		out += sep + s
	}
	return out
}

func strRepeat(s string, n int) string {
	out := ""
	for i := 0; i < n; i++ {
		out += s
	}
	return out
}
