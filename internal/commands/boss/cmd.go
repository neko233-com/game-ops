// The BOSS command center: one entry point to replace the daily judgment work
// of a lead designer + operations manager, in language a non-specialist can act
// on. Six subcommands cover every decision surface:
//
//	brief    <genre> [-values f.json]      # one-page verdict + top decisions
//	decide   <genre> [-values f.json]      # full sign-off packet (owner/deadline)
//	monitor  <genre> [-values f.json]      # traffic-light operations dashboard
//	duty     [-role 主策划|运营]             # full duty matrix with standards
//	simulate <genre> [-values f.json]      # 3 scenarios: now / fix / ignore
//	ask      <question>                    # natural-language routing + answer
//
// All analysis is exact (num.Num) and every category works (auto-derived).
package bosscmd

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/neko233-com/game-ops/boss"
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
	case "brief":
		cmdBrief(args)
	case "decide":
		cmdDecide(args)
	case "monitor":
		cmdMonitor(args)
	case "duty":
		cmdDuty(args)
	case "simulate":
		cmdSimulate(args)
	case "ask":
		cmdAsk(args)
	case "help", "-h", "--help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n", cmd)
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprint(os.Stderr, `boss — decision center for the boss (replaces lead-designer + ops judgment)

Usage:
  boss brief   <genre> [-values file.json]
  boss decide  <genre> [-values file.json] [-html out.html]
  boss monitor <genre> [-values file.json]
  boss duty    [-role 主策划|运营]
  boss simulate <genre> [-values file.json]
  boss ask     "<question>"            # e.g. "gacha 这个版本能上吗"

Output is conclusions first, then decisions needing sign-off, then instructions.
Unknown genres are auto-derived — any category works.
`)
}

// ---- shared ----

type bossFlags struct {
	values  *string
	base    *string
	html    *string
	region  *string
}

func newFlagSet(name string) (*flag.FlagSet, *bossFlags) {
	fs := flag.NewFlagSet(name, flag.ExitOnError)
	f := &bossFlags{
		values: fs.String("values", "", "JSON file of observed metrics"),
		base:   fs.String("base", "金币", "base value unit"),
		html:   fs.String("html", "", "write HTML cockpit to file"),
		region: fs.String("region", "GLOBAL", "CN|SEA|US|EU|GLOBAL"),
	}
	return fs, f
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
		// 也支持 {genre, values:{...}} 之外的纯 map
		var m map[string]interface{}
		if err2 := json.Unmarshal(b, &m); err2 != nil {
			return nil, err
		}
		raw.Values = m
	}
	for k, v := range raw.Values {
		switch t := v.(type) {
		case string:
			if n, err := num.FromString(t); err == nil {
				out[k] = n
			}
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

// reorderFlags lets "boss brief gacha -values f.json" work: Go's flag stops at
// the first positional arg, so move flags (and their values) in front.
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

func mustGenre(args []string, fs *flag.FlagSet) string {
	fs.Parse(reorderFlags(args))
	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "usage: boss <subcommand> <genre> [-values file.json]")
		os.Exit(1)
	}
	return fs.Arg(0)
}

func printVerdict(b boss.Brief) {
	callIcon := map[string]string{"GO": "[OK]", "HOLD": "[!!]", "STOP": "[XX]"}[b.Call]
	fmt.Printf("%s 决策：%s   健康度：%s\n", callIcon, b.Call, b.Score)
	fmt.Printf("结论：%s\n", b.Headline)
}

// ---- subcommands ----

func cmdBrief(args []string) {
	fs, f := newFlagSet("brief")
	id := mustGenre(args, fs)
	vals, err := readValues(*f.values)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	b, err := boss.Analyze(id, vals, map[string]bool{"region": *f.region != "GLOBAL"})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("========== 老板决策简报 · %s ==========\n", b.GenreName)
	printVerdict(b)
	fmt.Printf("\n需要你拍板的事（%d 项）：\n", len(b.Decisions))
	if len(b.Decisions) == 0 {
		fmt.Println("  无。数值在容差带内，无需干预。")
	}
	for _, d := range b.Decisions {
		fmt.Printf("\n  [%s] %s\n", d.Priority, d.Question)
		fmt.Printf("    证据：%s\n", d.Evidence)
		fmt.Printf("    建议：%s\n", d.Recommend)
		fmt.Printf("    不做的后果：%s\n", d.IfNot)
		fmt.Printf("    负责人：%s ｜ 期限：%s\n", d.Owner, d.Deadline)
	}
	fmt.Printf("\n最严重的数值问题：\n")
	if len(b.Findings) == 0 {
		fmt.Println("  无。")
	}
	for i, fi := range b.Findings {
		fmt.Printf("  %d. %s （%s）\n", i+1, fi.Headline, fi.Question)
	}
	fmt.Printf("\n边界提醒：%s\n", b.Confidence)
	if *f.html != "" {
		writeHTML(*f.html, b, false)
	}
}

func cmdDecide(args []string) {
	fs, f := newFlagSet("decide")
	id := mustGenre(args, fs)
	vals, err := readValues(*f.values)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	b, err := boss.Analyze(id, vals, map[string]bool{"region": *f.region != "GLOBAL"})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("========== 老板完整决策包 · %s ==========\n\n", b.GenreName)
	printVerdict(b)
	fmt.Println()
	fmt.Println("--- 一、需要拍板 ---")
	for _, d := range b.Decisions {
		fmt.Printf("\n[%s] %s\n", d.Priority, d.Question)
		fmt.Printf("  背景：%s\n", d.Context)
		fmt.Printf("  证据：%s\n", d.Evidence)
		fmt.Printf("  选项：\n")
		for _, o := range d.Options {
			fmt.Printf("    - %s\n", o)
		}
		fmt.Printf("  建议：%s\n", d.Recommend)
		fmt.Printf("  不做的后果：%s\n", d.IfNot)
		fmt.Printf("  负责人：%s ｜ 期限：%s\n", d.Owner, d.Deadline)
	}
	fmt.Println("\n--- 二、执行指令 ---")
	for _, a := range b.Actions {
		fmt.Printf("\n[%s] %s\n", a.Priority, a.Task)
		fmt.Printf("  负责人：%s\n", a.Owner)
		fmt.Printf("  工具：%s\n", a.Tool)
		fmt.Printf("  验收：%s\n", a.Verify)
	}
	fmt.Printf("\n--- 三、数值无法替你判断的（必须人工复核）---\n")
	for _, l := range b.Limits {
		fmt.Printf("  • %s\n", l)
	}
	if *f.html != "" {
		writeHTML(*f.html, b, true)
	}
}

func cmdMonitor(args []string) {
	fs, f := newFlagSet("monitor")
	id := mustGenre(args, fs)
	vals, err := readValues(*f.values)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	b, err := boss.Analyze(id, vals, map[string]bool{"region": *f.region != "GLOBAL"})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("========== 运营红黄绿灯 · %s ==========\n", b.GenreName)
	printVerdict(b)
	fmt.Printf("\n%-6s %-24s %-14s %-14s %s\n", "灯", "事务", "当前", "目标", "说明")
	fmt.Println(strings.Repeat("─", 88))
	for _, d := range b.Duties {
		if d.Metric == "" || d.Status == "idle" {
			continue
		}
		icon := map[string]string{"ok": "[绿]", "warn": "[黄]", "critical": "[红]"}[d.Status]
		// Finding 形如 "含义：当前 X / 目标 Y"，直接取后半段做紧凑展示。
		short := d.Finding
		if i := strings.Index(short, "："); i >= 0 {
			short = short[i+1:]
		}
		fmt.Printf("%-6s %-24s %-20s %s\n", icon, d.Area, short, d.Duty)
	}
	fmt.Println("\n未观测事务（需补数据）：")
	for _, d := range b.Duties {
		if d.Status == "idle" {
			fmt.Printf("  · %s / %s — %s\n", d.Role, d.Area, d.Duty)
		}
	}
}

func cmdDuty(args []string) {
	fs, f := newFlagSet("duty")
	role := fs.String("role", "", "主策划 | 运营")
	_ = f
	fs.Parse(args)
	fmt.Println("========== 主策 + 运营 完整工作事务矩阵 ==========")
	fmt.Println("每项绑定引擎能力与判定标准；数值可自动判定的已接入。")
	fmt.Println()
	fmt.Printf("%-6s %-8s %-14s %-38s %-30s %s\n", "编号", "角色", "领域", "事务", "判定标准", "引擎工具")
	fmt.Println(strings.Repeat("─", 130))
	for _, d := range boss.DutyMatrix {
		if *role != "" && d.Role != *role {
			continue
		}
		fmt.Printf("%-6s %-8s %-14s %-38s %-30s %s\n", d.ID, d.Role, d.Area, d.Duty, d.Standard, d.Tool)
	}
}

func cmdSimulate(args []string) {
	fs, f := newFlagSet("simulate")
	id := mustGenre(args, fs)
	vals, err := readValues(*f.values)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	g := genre.GetAny(id)
	base := map[string]num.Num{}
	for k, v := range vals {
		base[k] = v
	}
	for _, l := range g.Levers {
		if _, ok := base[l.Key]; !ok {
			base[l.Key] = l.Ideal
		}
	}
	fmt.Printf("========== 三情景推演 · %s ==========\n\n", g.Name)

	run := func(name string, v map[string]num.Num) {
		r, err := ops.EvaluateAny(ops.Snapshot{GenreID: id, Values: v})
		if err != nil {
			fmt.Printf("  %s: 计算失败 %v\n", name, err)
			return
		}
		fmt.Printf("  %-12s 健康度 %-10s 结论 %s\n", name, r.Score.DecimalString(4), verdictCN(r.Verdict))
	}
	run("当前状态", base)

	// 方案二：按建议把偏差拉回目标（理想值）
	fixed := map[string]num.Num{}
	for k, v := range base {
		fixed[k] = v
	}
	for _, l := range g.Levers {
		fixed[l.Key] = l.Ideal
	}
	run("按建议整改", fixed)

	// 方案三：偏差扩大 50%（不处理继续恶化）
	ignore := map[string]num.Num{}
	for k, v := range base {
		ignore[k] = v
	}
	half := num.MustFromString("1/2")
	for _, l := range g.Levers {
		if v, ok := base[l.Key]; ok {
			d := v.Sub(l.Ideal)
			ignore[l.Key] = v.Add(d.Mul(half))
		}
	}
	run("不处理恶化", ignore)

	fmt.Println("\n结论：整改可回到健康基线；放任偏离会持续放大风险。建议批准 brief 中的拍板项。")
}

func verdictCN(v string) string {
	return map[string]string{"healthy": "健康", "attention": "需关注", "critical": "有风险"}[v]
}

func cmdAsk(args []string) {
	q := strings.Join(args, " ")
	if q == "" {
		fmt.Fprintln(os.Stderr, "usage: boss ask \"gacha 这个版本能上吗\"")
		os.Exit(1)
	}
	// 轻量意图路由：从问题里提取品类 id/关键词，输出对应命令。
	id := extractGenre(q)
	if id == "" {
		fmt.Printf("未识别品类。请明确品类后重试，例如：boss ask \"gacha 这个版本能上吗\"\n")
		fmt.Println("可用品类前缀示例：gacha / slg / casual / mobo / cozy-farming-sim（任意 ID 均支持）")
		os.Exit(1)
	}
	fmt.Printf("问题：%s\n", q)
	fmt.Printf("识别品类：%s\n\n", id)
	fmt.Printf("老板问「能不能上」→ 直接看简报结论：\n")
	fmt.Printf("  go run ./cmd/boss brief %s\n", id)
	fmt.Printf("\n需要完整拍板材料（含负责人/期限/验收）→\n")
	fmt.Printf("  go run ./cmd/boss decide %s\n", id)
	fmt.Printf("\n想看三种决策后果 →\n")
	fmt.Printf("  go run ./cmd/boss simulate %s\n", id)
	fmt.Printf("\n想看主策+运营该做什么 →\n")
	fmt.Printf("  go run ./cmd/boss duty\n")
}

func extractGenre(q string) string {
	lower := strings.ToLower(q)
	// 内置/示例 id 优先精确匹配
	ids := append(genre.List(), "cozy-farming-sim")
	sort.Slice(ids, func(i, j int) bool { return len(ids[i]) > len(ids[j]) })
	for _, id := range ids {
		if strings.Contains(lower, strings.ToLower(id)) {
			return id
		}
	}
	words := map[string]string{
		"抽卡": "gacha", "卡牌": "card", "策略": "slg", "休闲": "casual",
		"射击": "fps", "二次元": "gacha", "农场": "cozy-farming-sim",
	}
	for w, id := range words {
		if strings.Contains(q, w) {
			return id
		}
	}
	return ""
}

// writeHTML is defined in html.go
