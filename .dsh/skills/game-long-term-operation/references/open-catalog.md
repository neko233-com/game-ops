# 开放品类目录：支持所有游戏品类

> 核心原则：品类不是封闭枚举，而是「11 项通用指标 + 数据驱动的 lever 表」。
> 任何游戏品类——已有的、未来的、你自己发明的——都能被立即分析。

## 双层 API（向后兼容 + 全面开放）

| 层 | API | 未知 ID 行为 | 用途 |
|---|---|---|---|
| 严格 | `genre.Register` / `genre.Get` / `ops.Evaluate` / `factory.Run` | 报错 | 需确保品类已显式定义的场景 |
| 宽松 | `genre.GetAny` / `ops.EvaluateAny` / `factory.GetAny` / `factory.RunAny` | 自动派生并注册 | "支持所有"——任意新品类即刻可用 |

现有 29 内置品类、29 个生成策略、所有 CLI 与测试保持不变。重构只**拓宽入口**，不改写语义。

## 自动派生规则（`genre.Derive`）

`Derive(id, name, tags)` 从 11 项 base levers 出发，按标签子串匹配（大小写不敏感）逐条覆盖。
后匹配的规则覆盖先匹配的。标签来源：显式 `-tags` + ID 按 `-` `_` 空格拆分。
多条规则命中时按 `deriveRules` 顺序叠加。

| 标签关键词 | 覆盖要点 |
|---|---|
| `gacha` `collection` `collect` | 付费率 6%、LTV 10 |
| `casual` `hyper` `light` | 付费率 1%、D1 35% |
| `mmo` `social` `guild` | 内容节奏 7 天、D30 12% |
| `competitive` `esports` `pvp` | 内容节奏 21 天、付费率 2% |
| `shooter` `fps` `tps` | 内容节奏 21 天 |
| `strategy` `4x` `slg` | LTV 12、D30 10% |
| `card` `tcg` | 付费率 5% |
| `sandbox` `creative` `ugc` | 内容节奏 30 天 |
| `narrative` `story` `visual` | 付费率 4% |
| `idle` `incremental` `auto` | D30 6% |
| `puzzle` `match` `brain` | 内容节奏 14 天 |
| `roguelike` `procedural` `permadeath` | 内容节奏 21 天 |
| `sim` `management` `economy` | source/sink 权重 1.1 |
| `farming` `cozy` `life` | D30 10%、内容节奏 30 天 |

无标签命中也有完整的 11 项 base levers——**永远不会"不支持"**。

## JSON Spec 格式

```json
[
  {
    "id": "cozy-farming-sim",
    "name": "治愈农场模拟",
    "tags": ["cozy", "farming", "sim", "life"],
    "levers": [
      { "key": "retention_d1", "ideal": "0.45", "min": "0.15", "max": "0.7", "weight": "1" }
    ]
  }
]
```

- 数值全部用**字符串**，经 `num.Num` 精确解析；`1/3`、`0.1` 不丢精度。
- `levers` 只需写要覆盖的项？—— 否：`register` 走 `Upsert`，spec 的 levers 即完整表。
  想基于派生结果修改：先 `derive ... -save g.json` 导出完整 11 项，再编辑 JSON 后 `register`。
- `export` 导出全量目录；`register` 可回灌，round-trip 无损。

## CLI 工作流

```sh
# 1) 派生一个从未存在的品类，保存 spec
go run ./cmd/catalog derive cozy-farming-sim -name 治愈农场 -tags cozy,farming,sim -save g.json
# 2) 编辑 g.json 调整任意 lever（可选）
# 3) 注册（覆盖同名品类）
go run ./cmd/catalog register g.json
# 4) 健康评估（无 values 时用 ideal 填充，返回满分基线）
go run ./cmd/catalog eval cozy-farming-sim -values examples/gacha.json
# 5) 品类策略（无专属 Strategy → 通用兜底）
go run ./cmd/catalog strategy cozy-farming-sim -region CN
# 6) 导出/备份全量目录
go run ./cmd/catalog export -out catalog/all-genres.json
```

## 通用兜底 Strategy

`factory.RunAny` 在找不到专属 Strategy 时返回 `UniversalStrategy`：
- 用 `genre.GetAny` 的 levers 作为品类指标；
- 对 `ctx.Metrics` 中出现的 lever 做 min/max 阈值检查，越界产出 `RiskItem`；
- 输出 3 条通用建议（加法优先、精确有理数、先归一化）。

想给新品类专属判定：走 `gen-strategies.go` 生成或直接实现 `factory.Strategy` 并在 `init` 注册；
`GetAny` 会自动优先使用它，无需改任何调用方。

## 不变量（重构未触碰）

- 11 项通用指标 key 不变；品类 lever 全覆盖检查；缺失指标严格模式报错。
- 精确有理数 `num.Num`；字符串配置；加法优先曲线；开放 `genre.Register`。
- 健康评分公式、五支柱、满级时间、概率计算、报告与图表全部不变。
