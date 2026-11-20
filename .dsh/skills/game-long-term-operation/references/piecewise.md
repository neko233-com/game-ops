# 多段加法曲线（Piecewise Additive Curves）

## 为什么是唯一正道

Skill 明确禁止全局指数（`base·1.1ⁿ`）——因为复利效应在封顶前很久就会爆炸。但纯加法 `base + step·(n-1)` 在 80 级以上会显得"平"——边际增量恒定，后期乏力。

**多段加法**解决这个矛盾：每一段区间用一个独立的加法曲线，边界处是设计意图的断点。

### 示例：三段式 XP 曲线

| 段位 | 区间 | base | step | 设计意图 |
|---|---|---|---|---|
| 新手期 | 1–10 | 100 | 10 | 快速上手，每日即可涨一级 |
| 成长期 | 11–30 | 200 | 15 | 加入更多活动类型，边际 step 上升 |
| 成熟期 | 31–60 | 500 | 20 | 边际 step 再加一档，配合内容消耗节奏 |
| 满级 cap | — | — | — | 总上限 10000 |

数学上它仍然是**分段线性**函数，不引入任何隐藏的复利。每个边界都是显式、可验证、可回退的。

## 生成与验证

```sh
go run ./cmd/curve examples/piecewise.json
go run ./cmd/curve example                   # 打印推荐默认值
go run ./cmd/curve \
  -name xp -cap 10000 -maxlevel 60 \
  -seg 1,10,100,10 -seg 11,30,200,15 -seg 31,60,500,20
```

曲线输出每一级 value、累计值和总价值，并附带为什么不推荐全局指数的说明。

如果配置中出现全局指数公式，`curve` 会拒绝并报 exit=2：

```
[rejected] error: global exponent is forbidden even with a cap (^)
  formula: base*1.15^level
exit status 2
```

## 段间约束

`design.PiecewiseCurve.ValidatePiece()` 检查：

1. **连续性**：段 N 的 end + 1 = 段 N+1 的 start（warn 级 gap 会提示）
2. **非负 step**：任何 step < 0 会破坏单调性（error）
3. **正 base**：任何 base ≤ 0 会让曲线无效（error）

Agent 设计 piecewise 曲线时应：

- 每段的 step 只在前一段设计意图变化时才变化；不要每层变
- 在边界处打印曲线连续性，确认过渡自然
- 用 `progression.Run` 把 piecewise 曲线喂进去，计算精确的满级时间和留存触达率

## 对比乘法 vs 多段加法

| 方案 | L10 值 | L30 值 | L50 值 | 爆炸点 |
|---|---|---|---|---|
| 乘法 `100·1.1ⁿ` | 259 | 1745 | 11739 | L40 约 45× 起点 |
| 多段加法 | 190 | 485 | 1600 | 无爆炸 |
| 纯加法 `100+10n` | 190 | 390 | 590 | 太平 |

乘法 50 级时的 XP 是多段加法的 7.3×——复利没有封顶能阻止它。Agent 永远不要选乘法。

## JSON 格式

```jsonc
{
  "name": "xp",
  "cap": "10000",
  "maxLevel": 60,
  "segments": [
    {"start": 1,  "end": 10, "base": "100", "step": "10"},
    {"start": 11, "end": 30, "base": "200", "step": "15"}
  ],
  "formulas": ["base*1.15^level"]  // 可选：guard，出现任何全局指数 → exit 2
}
```

`formulas` 字段是防御性检查——如果你同时有原始公式和 piecewise 分段，CLI 会拒绝任何含指数的原始公式。这是防止"有人顺手加了指数就 commit"的第二道锁。
