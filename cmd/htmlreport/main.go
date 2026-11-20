// Command htmlreport generates a self-contained HTML dashboard for bosses and
// non-numerical readers from an exact game-ops analysis.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/neko233-com/game-ops/i18n"
	"github.com/neko233-com/game-ops/report"
)

func main() {
	lang := flag.String("lang", "zh", "language: zh (default, primary) | en | ja")
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 || len(args) > 2 {
		fmt.Fprintln(os.Stderr, "usage: htmlreport [-lang zh|en|ja] <spec.json> [output.html]")
		os.Exit(2)
	}
	data, err := os.ReadFile(args[0])
	if err != nil {
		fmt.Fprintln(os.Stderr, "read spec:", err)
		os.Exit(1)
	}
	var spec report.Spec
	if err := json.Unmarshal(data, &spec); err != nil {
		fmt.Fprintln(os.Stderr, "parse spec:", err)
		os.Exit(1)
	}
	if spec.Lang == "" {
		spec.Lang = string(i18n.Parse(*lang))
	}
	view, err := report.Build(spec)
	if err != nil {
		fmt.Fprintln(os.Stderr, "build report:", err)
		os.Exit(1)
	}
	html, err := report.Render(view)
	if err != nil {
		fmt.Fprintln(os.Stderr, "render:", err)
		os.Exit(1)
	}
	if len(args) == 2 {
		if err := os.WriteFile(args[1], html, 0644); err != nil {
			fmt.Fprintln(os.Stderr, "write html:", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "wrote %s (%d bytes, lang=%s)\n", args[1], len(html), spec.Lang)
		return
	}
	os.Stdout.Write(html)
}
