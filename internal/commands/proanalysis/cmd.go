// Command proanalysis is the lead-designer workstation CLI: it computes exact
// time-to-cap across archetypes, runs a parallel sensitivity batch, scans
// config tables concurrently, and emits a dense professional Markdown report.
package proanalysiscmd

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/neko233-com/game-ops/design"
	"github.com/neko233-com/game-ops/i18n"
	"github.com/neko233-com/game-ops/ingest"
	"github.com/neko233-com/game-ops/num"
	"github.com/neko233-com/game-ops/progression"
)

type activityJSON struct {
	Name            string `json:"name"`
	XPPerMinute     string `json:"xpPerMinute"`
	DailyCapMinutes uint   `json:"dailyCapMinutes"`
}

type archetypeJSON struct {
	Name          string `json:"name"`
	MinutesPerDay uint   `json:"minutesPerDay"`
	Efficiency    string `json:"efficiency"`
	Weight        string `json:"weight"`
	RetentionD1   string `json:"retentionD1"`
	RetentionD7   string `json:"retentionD7"`
	RetentionD30  string `json:"retentionD30"`
}

func (a archetypeJSON) archetype() progression.Archetype {
	return progression.Archetype{
		Name: a.Name, MinutesPerDay: a.MinutesPerDay,
		Efficiency: num.MustFromString(a.Efficiency), Weight: num.MustFromString(a.Weight),
		RetentionD1: num.MustFromString(a.RetentionD1), RetentionD7: num.MustFromString(a.RetentionD7),
		RetentionD30: num.MustFromString(a.RetentionD30),
	}
}

type spec struct {
	Title    string `json:"title"`
	BaseUnit string `json:"baseUnit"`
	MaxLevel uint   `json:"maxLevel"`
	XPCurve  struct {
		Name, Base, Step, Cap string
	} `json:"xpCurve"`
	Activities  []activityJSON  `json:"activities"`
	Archetypes  []archetypeJSON `json:"archetypes"`
	Sensitivity struct {
		Steps     []string `json:"steps"`
		MaxLevels []uint   `json:"maxLevels"`
	} `json:"sensitivity"`
}

func fmtN(n num.Num) string {
	return fmt.Sprintf("%s (%.4f)", n.String(), n.Float64())
}

func findResult(s progression.Summary, archetype string) (progression.Result, bool) {
	for _, r := range s.Results {
		if r.ArchetypeName == archetype {
			return r, true
		}
	}
	return progression.Result{}, false
}

func Main() {
	specPath := flag.String("spec", "examples/proanalysis.json", "analysis spec JSON")
	workers := flag.Int("workers", 16, "parallel workers for scenarios and ingest")
	repeat := flag.Int("repeat", 1, "repeat the sensitivity grid N times (stress test)")
	maxFiles := flag.Int64("max-files", 1000000, "config file ceiling")
	roots := flag.String("roots", "examples", "comma-separated config roots")
	out := flag.String("out", "", "output Markdown file (default stdout)")
	langFlag := flag.String("lang", "zh", "language: zh (default, primary) | en | ja")
	flag.Parse()

	raw, err := os.ReadFile(*specPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "read spec:", err)
		os.Exit(1)
	}
	var s spec
	if err := json.Unmarshal(raw, &s); err != nil {
		fmt.Fprintln(os.Stderr, "parse spec:", err)
		os.Exit(1)
	}

	activities := make([]progression.Activity, 0, len(s.Activities))
	for _, a := range s.Activities {
		activities = append(activities, progression.Activity{
			Name: a.Name, XPPerMinute: num.MustFromString(a.XPPerMinute),
			DailyCapMinutes: a.DailyCapMinutes,
		})
	}
	archetypes := make([]progression.Archetype, 0, len(s.Archetypes))
	for _, a := range s.Archetypes {
		archetypes = append(archetypes, a.archetype())
	}
	curve := design.Curve{
		Name: s.XPCurve.Name, Base: num.MustFromString(s.XPCurve.Base),
		Step: num.MustFromString(s.XPCurve.Step), Cap: num.MustFromString(s.XPCurve.Cap),
	}

	// Build the sensitivity grid; repeat it for parallel stress tests.
	steps := s.Sensitivity.Steps
	if len(steps) == 0 {
		steps = []string{s.XPCurve.Step}
	}
	levels := s.Sensitivity.MaxLevels
	if len(levels) == 0 {
		levels = []uint{s.MaxLevel}
	}
	var scenarios []progression.Scenario
	scenarios = append(scenarios, progression.Scenario{
		Name:     fmt.Sprintf("step=%s,cap=%d (基准)", s.XPCurve.Step, s.MaxLevel),
		MaxLevel: s.MaxLevel, XPCurve: curve,
		Activities: activities, Archetypes: archetypes,
	})
	for repeatIdx := 0; repeatIdx < *repeat; repeatIdx++ {
		for _, step := range steps {
			for _, level := range levels {
				c := curve
				c.Step = num.MustFromString(step)
				name := fmt.Sprintf("step=%s,cap=%d", step, level)
				if *repeat > 1 {
					name = fmt.Sprintf("%s#%d", name, repeatIdx)
				}
				scenarios = append(scenarios, progression.Scenario{
					Name: name, MaxLevel: level, XPCurve: c,
					Activities: activities, Archetypes: archetypes,
				})
			}
		}
	}

	// Run scenario batch and config ingest concurrently.
	start := time.Now()
	var summaries []progression.Summary
	var stats ingest.Stats
	var batchErr, ingestErr error
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		summaries, batchErr = progression.Batch(scenarios, *workers)
	}()
	go func() {
		defer wg.Done()
		stats, ingestErr = ingest.Scan(context.Background(), ingest.Options{
			Roots: []string{*roots}, MaxFiles: *maxFiles, Workers: *workers,
		})
	}()
	wg.Wait()
	if batchErr != nil {
		fmt.Fprintln(os.Stderr, "batch:", batchErr)
		os.Exit(1)
	}
	if ingestErr != nil {
		fmt.Fprintln(os.Stderr, "ingest:", ingestErr)
		os.Exit(1)
	}
	elapsed := time.Since(start)

	lang := i18n.Parse(*langFlag)
	tr := func(key string) string { return i18n.T(lang, key) }
	base := summaries[0]
	buf := &bufWriter{}
	w := buf.WriteString

	w("# %s\n\n", s.Title)
	w("- %s：%s ｜ %s ｜ %s：%d ｜ %s：%d ｜ %s：%d / %s：%d ｜ %s：%s\n\n",
		tr("meta.generated"), time.Now().Format("2006-01-02 15:04:05"), tr("label.exact"),
		tr("label.parallel"), *workers, tr("label.scenarios"), len(scenarios),
		tr("label.config"), stats.Files, tr("label.rows"), stats.Rows,
		tr("label.elapsed"), elapsed.Round(time.Millisecond))

	w("## 1. %s\n\n", tr("section.method.title"))
	w("- %s\n", tr("method.totalXP"))
	w("- %s\n", tr("method.dailyXP"))
	w("- %s\n", tr("method.days"))
	w("- %s：base=%s，step=%s，cap=%s，%s=%d。\n\n",
		tr("meta.curve"), s.XPCurve.Base, s.XPCurve.Step, s.XPCurve.Cap, tr("meta.baseMaxLevel"), s.MaxLevel)

	w("## 2. %s（%s %s）\n\n", tr("section.capacity.title"), tr("column.scenario"), base.ScenarioName)
	w("| %s | %s | %s | %s | %s | %s | %s |\n",
		tr("column.archetype"), tr("column.totalXP"), tr("column.dailyXP"),
		tr("column.daysExact"), tr("column.daysApprox"), tr("column.survival"), tr("column.weight"))
	w("|---|---|---|---|---|---|---|\n")
	for _, r := range base.Results {
		w("| %s | %s | %s | %s | %.2f | %s | %s |\n",
			r.ArchetypeName, r.TotalXP.String(), r.DailyXP.String(), r.Days.String(),
			r.Days.Float64(), r.SurvivalAtCap.DecimalString(4), r.Weight.String())
	}
	w("\n%s：\n\n", tr("label.weightedPercentiles"))
	w("| P10 | P25 | P50 | P75 | P90 |\n|---|---|---|---|---|\n")
	w("| %s | %s | %s | %s | %s |\n\n",
		base.Percentiles["p10"].DecimalString(2), base.Percentiles["p25"].DecimalString(2),
		base.Percentiles["p50"].DecimalString(2), base.Percentiles["p75"].DecimalString(2),
		base.Percentiles["p90"].DecimalString(2))

	w("## 3. %s（%s %d）\n\n", tr("section.sensitivity.title"), tr("label.scenarios"), len(summaries))
	w("| %s | step | %s | %s | %s | %s | %s |\n",
		tr("column.scenario"), tr("column.cap"), tr("column.normalDays"),
		tr("column.lightDays"), tr("column.heavyDays"), tr("column.p50Days"))
	w("|---|---|---|---|---|---|---|\n")
	for _, sum := range summaries {
		normal := sum.Results[0]
		light := sum.Results[1]
		heavy := sum.Results[2]
		w("| %s | %s | %s | %s | %s | %s | %s |\n",
			sum.ScenarioName, stepOf(sum.ScenarioName), levelOf(sum.ScenarioName),
			normal.Days.DecimalString(2), light.Days.DecimalString(2),
			heavy.Days.DecimalString(2), sum.Percentiles["p50"].DecimalString(2))
	}
	w("\n")

	w("## 4. %s\n\n", tr("section.qa.title"))
	type spread struct {
		key   string
		ratio num.Num
	}
	var spreads []spread
	for key, agg := range stats.Columns {
		if agg.Min.IsZero() {
			continue
		}
		r, err := agg.Max.Div(agg.Min)
		if err == nil {
			spreads = append(spreads, spread{key, r})
		}
	}
	sort.Slice(spreads, func(i, j int) bool {
		c := spreads[i].ratio.Cmp(spreads[j].ratio)
		if c != 0 {
			return c > 0
		}
		return spreads[i].key < spreads[j].key // deterministic tie-break for map order
	})
	w("- %s：%d；%s：%d\n", tr("label.parseErrors"), stats.Errors, tr("label.columns"), len(stats.Columns))
	if len(spreads) == 0 {
		w("- %s\n", tr("note.noColumns"))
	}
	for i, sp := range spreads {
		if i >= 10 {
			break
		}
		flagMark := ""
		if sp.ratio.Cmp(num.MustFromString("10")) > 0 {
			flagMark = " ⚠ " + tr("flag.outlier")
		}
		w("- `%s`：max/min = %s%s\n", sp.key, sp.ratio.DecimalString(2), flagMark)
	}
	w("\n")

	w("## 5. %s\n\n", tr("section.recommendations.title"))
	p50 := base.Percentiles["p50"]
	switch {
	case p50.Cmp(num.MustFromString("120")) > 0:
		w("- **%s**\n", tr("conclusion.long"))
	case p50.Cmp(num.MustFromString("60")) < 0:
		w("- **%s**\n", tr("conclusion.short"))
	default:
		w("- **%s**\n", tr("conclusion.ok"))
	}
	normal := base.Results[0]
	if normal.SurvivalAtCap.Cmp(num.MustFromString("1/10")) < 0 {
		w("- **%s**\n", fmt.Sprintf(tr("conclusion.survival"), normal.SurvivalAtCap.DecimalString(4)))
	}
	for _, sp := range spreads {
		if sp.ratio.Cmp(num.MustFromString("10")) > 0 {
			w("- **%s**\n", fmt.Sprintf(tr("conclusion.config"), sp.key, sp.ratio.DecimalString(2)))
			break
		}
	}
	w("- %s\n", tr("advice.verify"))

	text := buf.String()
	if *out != "" {
		if err := os.WriteFile(*out, []byte(text), 0644); err != nil {
			fmt.Fprintln(os.Stderr, "write out:", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "wrote %s\n", *out)
		return
	}
	fmt.Print(text)
}

func stepOf(name string) string {
	for i := 0; i+5 <= len(name); i++ {
		if name[i:i+5] == "step=" {
			rest := name[i+5:]
			for j, c := range rest {
				if c == ',' {
					return rest[:j]
				}
			}
			return rest
		}
	}
	return "-"
}

func levelOf(name string) string {
	for i := 0; i+4 <= len(name); i++ {
		if name[i:i+4] == "cap=" {
			rest := name[i+4:]
			for j, c := range rest {
				if c == ',' || c == ' ' || c == '(' {
					return rest[:j]
				}
			}
			return rest
		}
	}
	return "-"
}

type bufWriter struct{ b []byte }

func (w *bufWriter) WriteString(format string, args ...interface{}) {
	w.b = append(w.b, []byte(fmt.Sprintf(format, args...))...)
}

func (w *bufWriter) String() string { return string(w.b) }
