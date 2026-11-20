// Package main is the single production binary: gameops.
//
// One binary carries every capability of the game long-term-operation engine.
// All subcommands are the original cmd/* packages, now importable as Main(),
// so behavior is byte-for-byte identical to `go run ./cmd/<name>` — with zero
// duplicated logic. The dispatcher rewrites os.Args so each command's own
// flag.Parse()/os.Args usage keeps working unchanged.
//
// Build:
//
//	go build -ldflags "-X main.version=v1.0.0 -X main.commit=$(git rev-parse --short HEAD) -X main.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" -o gameops ./cmd/gameops
package main

import (
	"fmt"
	"os"
	"sort"

	boss       "github.com/neko233-com/game-ops/internal/commands/boss"
	catalog    "github.com/neko233-com/game-ops/internal/commands/catalog"
	curve      "github.com/neko233-com/game-ops/internal/commands/curve"
	econcheck  "github.com/neko233-com/game-ops/internal/commands/econcheck"
	formula    "github.com/neko233-com/game-ops/internal/commands/formula"
	genre      "github.com/neko233-com/game-ops/internal/commands/genre"
	htmlreport "github.com/neko233-com/game-ops/internal/commands/htmlreport"
	ingest     "github.com/neko233-com/game-ops/internal/commands/ingest"
	judge      "github.com/neko233-com/game-ops/internal/commands/judge"
	normalize  "github.com/neko233-com/game-ops/internal/commands/normalize"
	numcalc    "github.com/neko233-com/game-ops/internal/commands/numcalc"
	numgen     "github.com/neko233-com/game-ops/internal/commands/numgen"
	opsreport  "github.com/neko233-com/game-ops/internal/commands/opsreport"
	proanalysis "github.com/neko233-com/game-ops/internal/commands/proanalysis"
	strategist "github.com/neko233-com/game-ops/internal/commands/strategist"
	"github.com/neko233-com/game-ops/internal/buildinfo"
)

// Injected at build time via -ldflags. Defaults make `go run` useful too.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

type command struct {
	name  string
	brief string
	run   func()
}

var commands = []command{
	{"boss", "老板决策中枢（GO/HOLD/STOP + 拍板项 + 执行指令）", boss.Main},
	{"catalog", "开放品类目录（derive/register/eval/strategy/export）", catalog.Main},
	{"opsreport", "五支柱健康报告（11 指标评分）", opsreport.Main},
	{"econcheck", "经济 source/sink 平衡", econcheck.Main},
	{"normalize", "单位价值归一化（精确换算）", normalize.Main},
	{"numgen", "加法优先数值曲线 + 指数守卫", numgen.Main},
	{"curve", "多段加法曲线", curve.Main},
	{"numcalc", "精确有理数计算（分数/组合/概率/统计）", numcalc.Main},
	{"formula", "精确数值表达式求值", formula.Main},
	{"genre", "品类 lever 查询", genre.Main},
	{"strategist", "品类专属策略（区域感知）", strategist.Main},
	{"judge", "AI 版本判断包", judge.Main},
	{"ingest", "流式配表摄取（百万级）", ingest.Main},
	{"proanalysis", "满级时间 + 并行敏感度 + 配表分析", proanalysis.Main},
	{"htmlreport", "Vercel 风格 HTML 报告", htmlreport.Main},
}

func main() {
	defer recoverPanic()
	buildinfo.Set(version, commit, date)

	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}
	cmd := os.Args[1]
	switch cmd {
	case "version", "-version", "--version", "-V":
		fmt.Printf("gameops %s\ncommit: %s\nbuilt:  %s\ngo:     %s\n", version, commit, date, buildinfo.GoVersion())
		return
	case "help", "-h", "--help":
		usage()
		return
	}
	for _, c := range commands {
		if c.name == cmd {
			// Rewrite os.Args to [binary, subArgs...] so each command's own
			// flag.Parse()/os.Args indexing sees exactly its arguments.
			os.Args = append([]string{os.Args[0]}, os.Args[2:]...)
			c.run()
			return
		}
	}
	fmt.Fprintf(os.Stderr, "gameops: unknown command %q\n\n", cmd)
	usage()
	os.Exit(1)
}

func recoverPanic() {
	if r := recover(); r != nil {
		fmt.Fprintf(os.Stderr, "gameops: internal error (panic recovered): %v\n", r)
		os.Exit(70) // sysexits: EX_SOFTWARE
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `gameops — game long-term-operation engine (single binary)

Usage:
  gameops <command> [args...]
  gameops version
  gameops help

Commands (%d):
`, len(commands))
	names := make([]string, 0, len(commands))
	for _, c := range commands {
		names = append(names, c.name)
	}
	sort.Strings(names)
	byName := map[string]string{}
	for _, c := range commands {
		byName[c.name] = c.brief
	}
	for _, n := range names {
		fmt.Fprintf(os.Stderr, "  %-14s %s\n", n, byName[n])
	}
	fmt.Fprintf(os.Stderr, `
Examples:
  gameops boss brief gacha -values examples/gacha-warn.json
  gameops opsreport examples/gacha.json
  gameops numcalc '1/3 + 2/7'
  gameops catalog derive cozy-farming -tags cozy,farming

All analysis uses exact rational arithmetic. Every genre is supported.
`)
}
