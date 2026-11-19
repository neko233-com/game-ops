# game-ops — 游戏长线运营与数值 Skill + 引擎

一个纯 Go（**Go 1.27，无 cgo，仅标准库**）的数值/长线运营引擎，外加一个跨运行时的
Agent Skill。**复制本仓库 → 读本 README → 安装即用。**

## 30 秒上手

```powershell
# 1) 安装 skill（项目级优先，自动覆盖 opencode / codex / dsh / claude）
scripts/install-skill.ps1          # macOS/Linux: bash scripts/install-skill.sh

# 2) 直接对 Agent 说
#    "用 game-long-term-operation skill 审查我的战斗数值曲线"
```

或者只用引擎：

```powershell
$env:CGO_ENABLED = "0"
go run ./cmd/opsreport examples/gacha.json        # 长线运营健康报告
go run ./cmd/normalize examples/normalize.json    # 单位价值归一化
go run ./cmd/numgen examples/plan.json            # 生成战斗/成长/商业化数值
go run ./cmd/numgen -formula 'base*1.2^level' examples/plan.json  # 指数公式守卫
go run ./cmd/econcheck examples/econ.json         # source/sink 平衡 (~1)
go run ./cmd/judge -values examples/gacha.json    # 给 AI 的版本判断包
go run ./cmd/htmlreport examples/report.json dashboard.html  # 老板版 HTML 仪表盘
go run ./cmd/ingest -roots ./Config -max 1000000 -workers 16 > stats.json
```

## 大道归一：数值的核心

> 数值的本质不是公式，而是**价值的定义、流动与节奏**。

所有数值任务（战斗投放、成长、商业化）都归约为三问：**定义**（基准单位是什么、
其他资源值多少）、**流动**（source/sink 是否平衡）、**节奏**（价值产出/消耗的速度
如何随时间变化）。详见 `skills/game-long-term-operation/references/unified-theory.md`，
以及纯 Vite+ React 文档站 `docs-site/`（含全领域踩坑表、加法优先、寿命判断等）。

**游戏寿命能否用数值判断？** 能给出领先指标（留存衰减、内容消耗/产出比、经济通胀、
成长饱和、付费集中度），但不是最终判决——趣味与社交无法被数值替代。报告指标时必须
同时报告其局限。

## 关键设计决策

| 决策 | 原因 |
|---|---|
| **加法优先** `base + step*(n-1)` + cap | 边际价值恒定、可预测、可封顶；乘法叠加导致数值膨胀/边界爆炸/复利崩溃 |
| **精确有理数** `math/big.Rat` | `0.1+0.2 == 3/10`；无 float 舍入；支持任意大小与 `1/3` |
| **字符串配置** | 精度穿过 JSON；Agent 改的是字符串值 |
| **单位归一化** | 所有资源换算到一个基准单位后才能比较奖励/价格/战力 |
| **开放注册表** | 11 个通用指标 + 每品类 levers，支持任意新品类，不是封闭枚举 |
| **流式摄取** | 有界 worker 池，100 万张配表不爆内存 |

## 包与命令

| 路径 | 作用 |
|---|---|
| `num` | 精确有理数（任意数值、分数、科学计数法、JSON round-trip） |
| `genre` | 29 个内置品类 + `Register` 开放注册；11 个通用指标 |
| `ops` | `Evaluate(Snapshot) -> Report`：加权健康评分 |
| `formula` | 精确数值表达式求值（`arpu * pay_rate`、`2^3^2`） |
| `normalize` | 单位价值账本；精确换算与传递换算 |
| `ingest` | 流式扫描 100 万 `.xlsx/.json/.jsonl/.csv/.tsv`（仅标准库） |
| `design` | 加法优先的战斗/成长/商业化曲线 + 爆炸校验 |
| `report` / `cmd/htmlreport` | 面向老板的**自包含 HTML 仪表盘**（内联 CSS+SVG，可离线/邮件） |
| `cmd/judge` | 给 AI 的版本判断包：精确偏差 + 配表异常 + 可复制 prompt |
| `cmd/opsreport` / `cmd/normalize` / `cmd/numgen` / `cmd/econcheck` / `cmd/ingest` | 对应 CLI（健康报告 / 归一化 / 数值生成 / 经济平衡 / 配表摄取） |
| `design.RejectGlobalExponent` | 拒绝任何全局指数公式（即使加了 cap） |
| `docs-site/` 的 React Flow 页 | 价值流动图 + 系统关系图（可拖拽/缩放，给非数值读者） |

## 让 Agent 设计全部数值

Skill 明确授权 Agent 设计**所有**数值：战斗投放、成长曲线、商业化定价。流程：
归一化 → 加法曲线 → 校验 → 摄取 100 万配表 → 用 `ops` 报告评判当前版本 →
Agent 给出具体配置改动 → 重新测量。Prompt 模板见 `references/ai-design-workflow.md`。

## 心流分析（可选 / 较慢 / 本项目不测试）

搭配 [codegraph233](https://github.com/neko233-com/codegraph233)（Rust AST GraphRAG MCP，
覆盖代码、prefab/scene/shader、配表），按需在项目级自动下载：

```powershell
scripts/codegraph-setup.ps1        # macOS/Linux: bash scripts/codegraph-setup.sh
codegraph233 init; codegraph233 index; codegraph233 status
# MCP: cg233_design_graph / cg233_config_tables / cg233_call_graph / cg233_asset_tree
```

心流分析比数值评分慢得多，是独立可选的一遍；脚本未在本仓库测试中验证。

## Skill 安装与多运行时

项目级优先，全局兜底。目标目录：

| 运行时 | 项目级 | 全局 |
|---|---|---|
| OpenCode | `.opencode/skills/` | `~/.config/opencode/skills/` |
| Claude Code | `.claude/skills/` | `~/.claude/skills/` |
| Codex | `.agents/skills/` | `~/.agents/skills/` |
| dsh | `.dsh/skills/` | `~/.dsh/skills/`（或 `$DSH_SKILLS_DIR`） |

`scripts/install-skill.ps1 [-Scope project|global|all]`（或 `.sh`）。详见
`skills/game-long-term-operation/references/installation.md`。

## 测试内容不会被其他人拉取

`.gitignore` 忽略 `*_test.go`、`/tests/`、`**/testdata/`。本地用
`go test ./...` 跑黑盒测试；这些文件永不提交/推送。

## 验证

```powershell
$env:CGO_ENABLED = "0"
go build ./...
go vet ./...
go test -count=1 ./...
# docs-site: cd docs-site; npm install; npm run build
```

## 文档站

`docs-site/` 是用 **Vite+ (vite-plus) + React + TypeScript** 编写的配套文档，
覆盖全部基础概念（大道归一、归一化、加法优先、踩坑表、五大支柱、战斗/成长/商业化、
百万配表分析、心流、安装）。`cd docs-site && npm install && npm run dev`。
