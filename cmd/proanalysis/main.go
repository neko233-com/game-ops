// Command proanalysis is the lead-designer workstation CLI: it computes exact
// time-to-cap across archetypes, runs a parallel sensitivity batch, scans
// config tables concurrently, and emits a dense professional Markdown report.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/neko233/game-ops/design"
	"github.com/neko233/game-ops/ingest"
	"github.com/neko233/game-ops/num"
	"github.com/neko233/game-ops/progression"
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

func main() {
	specPath := flag.String("spec", "examples/proanalysis.json", "analysis spec JSON")
	workers := flag.Int("workers", 16, "parallel workers for scenarios and ingest")
	repeat := flag.Int("repeat", 1, "repeat the sensitivity grid N times (stress test)")
	maxFiles := flag.Int64("max-files", 1000000, "config file ceiling")
	roots := flag.String("roots", "examples", "comma-separated config roots")
	out := flag.String("out", "", "output Markdown file (default stdout)")
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

	base := summaries[0]
	buf := &bufWriter{}
	w := buf.WriteString

	w("# %s\n\n", s.Title)
	w("- 生成时间：%s ｜ 精确有理数计算（`num.Num`），无浮点误差\n", time.Now().Format("2006-01-02 15:04:05"))
	w("- 基准单位：%s ｜ 并行度：%d workers ｜ 场景数：%d ｜ 配表：%d 文件 / %d 行 ｜ 耗时：%s\n\n",
		s.BaseUnit, *workers, len(scenarios), stats.Files, stats.Rows, elapsed.Round(time.Millisecond))

	w("## 1. 方法与输入\n\n")
	w("- 总经验 = Σ(l=1..满级-1) [base + step·(l−1)]，精确求和。\n")
	w("- 每日经验 = Σ(活动分钟 × 每分钟经验) × 玩家效率；活动按日上限顺序分配时间。\n")
	w("- 满级天数 = 总经验 ÷ 每日经验（精确分数）；留存触达率 = D1/D7/D30 锚点线性插值外推。\n")
	w("- 经验曲线：base=%s，基准 step=%s，cap=%s，基准满级=%d。\n\n",
		s.XPCurve.Base, s.XPCurve.Step, s.XPCurve.Cap, s.MaxLevel)

	w("## 2. 常规/轻度/重度玩家满级时间（基准场景 %s）\n\n", base.ScenarioName)
	w("| 玩家类型 | 总经验 | 每日经验 | 满级天数（精确） | 满级天数（约） | 满级时留存 | 人口权重 |\n")
	w("|---|---|---|---|---|---|---|\n")
	for _, r := range base.Results {
		w("| %s | %s | %s | %s | %.2f | %s | %s |\n",
			r.ArchetypeName, r.TotalXP.String(), r.DailyXP.String(), r.Days.String(),
			r.Days.Float64(), r.SurvivalAtCap.DecimalString(4), r.Weight.String())
	}
	w("\n加权分位数（按人口权重）：\n\n")
	w("| P10 | P25 | P50 | P75 | P90 |\n|---|---|---|---|---|\n")
	w("| %s | %s | %s | %s | %s |\n\n",
		base.Percentiles["p10"].DecimalString(2), base.Percentiles["p25"].DecimalString(2),
		base.Percentiles["p50"].DecimalString(2), base.Percentiles["p75"].DecimalString(2),
		base.Percentiles["p90"].DecimalString(2))

	w("## 3. 敏感度扫描（并行 %d 场景）\n\n", len(summaries))
	w("| 场景 | step | 满级 | 常规天数 | 轻度天数 | 重度天数 | P50 天数 |\n")
	w("|---|---|---|---|---|---|---|\n")
	for _, sum := range summaries {
		normal, _ := findResult(sum, "常规")
		light, _ := findResult(sum, "轻度")
		heavy, _ := findResult(sum, "重度")
		w("| %s | %s | %s | %s | %s | %s | %s |\n",
			sum.ScenarioName, stepOf(sum.ScenarioName), levelOf(sum.ScenarioName),
			normal.Days.DecimalString(2), light.Days.DecimalString(2),
			heavy.Days.DecimalString(2), sum.Percentiles["p50"].DecimalString(2))
	}
	w("\n")

	w("## 4. 配表异常（并行扫描）\n\n")
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
	w("- 解析错误：%d；列数：%d\n", stats.Errors, len(stats.Columns))
	if len(spreads) == 0 {
		w("- 无可分析列。\n")
	}
	for i, sp := range spreads {
		if i >= 10 {
			break
		}
		flagMark := ""
		if sp.ratio.Cmp(num.MustFromString("10")) > 0 {
			flagMark = " ⚠ 超过 10×，需人工确认"
		}
		w("- `%s`：max/min = %s%s\n", sp.key, sp.ratio.DecimalString(2), flagMark)
	}
	w("\n")

	w("## 5. 结论与建议\n\n")
	p50 := base.Percentiles["p50"]
	if p50.Cmp(num.MustFromString("120")) > 0 {
		w("- **满级周期偏长**：P50 超过 120 天。建议上调经验曲线 step、提高活动经验密度，或降低满级门槛；用第 3 节敏感度表选择目标档位。\n")
	} else if p50.Cmp(num.MustFromString("60")) < 0 {
		w("- **满级周期偏短**：P50 不足 60 天，长期内容消耗可能过快。建议下调 step 或补充满级后追求目标。\n")
	} else {
		w("- **满级周期在合理区间**：P50 处于 60–120 天。\n")
	}
	normal, _ := findResult(base, "常规")
	if normal.SurvivalAtCap.Cmp(num.MustFromString("1/10")) < 0 {
		w("- **满级触达率过低**：常规玩家满级时留存仅 %s，多数玩家在满级前流失；建议把关键奖励前移，或降低满级所需总经验。\n",
			normal.SurvivalAtCap.DecimalString(4))
	}
	for _, sp := range spreads {
		if sp.ratio.Cmp(num.MustFromString("10")) > 0 {
			w("- **配表异常**：`%s` 的 max/min 达 %s，检查是否误填单位或指数公式。\n", sp.key, sp.ratio.DecimalString(2))
			break
		}
	}
	w("- 所有建议需用 `cmd/numgen`、`cmd/ingest`、`cmd/ops` 重新验证后再进入版本。\n")

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
