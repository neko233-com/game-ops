// Command judge assembles an AI-ready judgment packet: the current build's
// health, the worst exact deviations, config-table observations and a copyable
// prompt. The agent owns every number; this tool supplies the measurements.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"

	"github.com/neko233/game-ops/genre"
	"github.com/neko233/game-ops/ingest"
	"github.com/neko233/game-ops/num"
	"github.com/neko233/game-ops/ops"
)

type valuesFile struct {
	Genre  string            `json:"genre"`
	Values map[string]string `json:"values"`
}

type finding struct {
	Key      string
	Weighted num.Num
	Observed string
	Ideal    string
	Status   string
	Advice   string
}

func main() {
	valuesPath := flag.String("values", "", "gacha-style JSON with genre + string values")
	statsPath := flag.String("stats", "", "optional ingest stats JSON")
	baseUnit := flag.String("base", "金币", "base value unit for the prompt")
	flag.Parse()
	if *valuesPath == "" {
		fmt.Fprintln(os.Stderr, "usage: judge -values <values.json> [-stats stats.json] [-base 金币]")
		os.Exit(2)
	}

	raw, err := os.ReadFile(*valuesPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "read values:", err)
		os.Exit(1)
	}
	var vf valuesFile
	if err := json.Unmarshal(raw, &vf); err != nil {
		fmt.Fprintln(os.Stderr, "parse values:", err)
		os.Exit(1)
	}
	g, ok := genre.Get(vf.Genre)
	if !ok {
		fmt.Fprintln(os.Stderr, "unknown genre:", vf.Genre)
		os.Exit(1)
	}
	values := map[string]num.Num{}
	for k, v := range vf.Values {
		n, err := num.FromString(v)
		if err != nil {
			fmt.Fprintf(os.Stderr, "value %q: %v\n", k, err)
			os.Exit(1)
		}
		values[k] = n
	}
	report, err := ops.Evaluate(ops.Snapshot{GenreID: vf.Genre, Values: values})
	if err != nil {
		fmt.Fprintln(os.Stderr, "evaluate:", err)
		os.Exit(1)
	}

	leverWeight := map[string]num.Num{}
	for _, l := range g.Levers {
		leverWeight[l.Key] = l.Weight
	}
	var findings []finding
	for _, d := range report.Diagnostics {
		if d.Status == "ok" {
			continue
		}
		weighted := d.Deviation.Mul(leverWeight[d.Metric])
		direction := "上调"
		if d.Observed.Cmp(d.Ideal) > 0 {
			direction = "下调"
		}
		findings = append(findings, finding{
			Key: d.Metric, Weighted: weighted,
			Observed: d.Observed.String(), Ideal: d.Ideal.String(), Status: d.Status,
			Advice: "建议" + direction + "该指标向目标靠拢，并说明对价值/流动/节奏的影响",
		})
	}
	sort.Slice(findings, func(i, j int) bool {
		return findings[i].Weighted.Cmp(findings[j].Weighted) > 0
	})

	fmt.Printf("# 版本数值判断包（%s / %s）\n\n", g.Name, g.ID)
	fmt.Printf("- 健康度：**%s**，结论：**%s**\n", report.Score.DecimalString(4), report.Verdict)
	fmt.Printf("- 基准单位：**%s**（所有价值先换算到它再比较）\n", *baseUnit)
	fmt.Printf("- 告警/关注指标：%d 项\n\n", len(findings))

	fmt.Println("## 按重要性排序的偏差（精确有理数）")
	if len(findings) == 0 {
		fmt.Println("当前没有超出容差带的指标。保持节奏，按周期复盘。")
	}
	for i, f := range findings {
		fmt.Printf("%d. `%s` 状态=%s 当前=%s 目标=%s 加权偏差=%s — %s\n",
			i+1, f.Key, f.Status, f.Observed, f.Ideal, f.Weighted.String(), f.Advice)
	}
	fmt.Println()

	if *statsPath != "" {
		sb, err := os.ReadFile(*statsPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, "read stats:", err)
			os.Exit(1)
		}
		var stats ingest.Stats
		if err := json.Unmarshal(sb, &stats); err != nil {
			fmt.Fprintln(os.Stderr, "parse stats:", err)
			os.Exit(1)
		}
		fmt.Println("## 配表数据观察")
		fmt.Printf("- 扫描 %d 个文件 / %d 行，解析错误 %d\n", stats.Files, stats.Rows, stats.Errors)
		type outlier struct {
			key   string
			ratio num.Num
		}
		var outs []outlier
		for key, agg := range stats.Columns {
			if agg.Min.IsZero() {
				continue
			}
			r, err := agg.Max.Div(agg.Min)
			if err == nil && r.Cmp(num.MustFromString("10")) > 0 {
				outs = append(outs, outlier{key, r})
			}
		}
		sort.Slice(outs, func(i, j int) bool { return outs[i].ratio.Cmp(outs[j].ratio) > 0 })
		for _, o := range outs {
			fmt.Printf("- `%s` 最大值是最小值的 %s 倍，可能存在配表异常或设计上的巨大跨度，请 AI 重点检查。\n", o.key, o.ratio.DecimalString(2))
		}
		if len(outs) == 0 {
			fmt.Println("- 未发现跨度超过 10 倍的列。")
		}
		fmt.Println()
	}

	fmt.Println("## 可直接复制给 AI 的指令")
	fmt.Println("```")
	fmt.Printf("你拥有全部数值的决定权。基准单位：%s。品类：%s。\n", *baseUnit, g.ID)
	fmt.Printf("当前健康度：%s（%s）。最严重的偏差：%v。\n", report.Score.DecimalString(4), report.Verdict, findings)
	fmt.Println("请判断当前版本数值是否健康，列出最严重的 3 个价值/流动/节奏问题（用精确数字），")
	fmt.Println("并给出把配置里哪个字段的字符串值从多少改成多少。改完后重新运行 numgen、ingest、ops 验证。")
	fmt.Println("```")
}
