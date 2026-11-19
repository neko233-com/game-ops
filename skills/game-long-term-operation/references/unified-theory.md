# 大道归一 — The Core of Numerical Design

## The thesis

Every number in a game is an answer to one of three questions:

1. **价值定义 (definition)** — what is one unit of this thing worth, in the base unit?
2. **价值流动 (flow)** — where does it come from (source), where does it go (sink)?
3. **价值节奏 (pacing)** — how fast does it move, and how does that rate change over time?

Combat power, a level curve, a gacha price, a daily reward, a drop rate — all are
instances of these three. That is the "归一" (return to one): one method, many
surfaces. A designer who internalizes this can switch genres without relearning
the craft, because genres differ only in their **levers**, not in the method.

### The base unit

Pick one unit and state it explicitly. Two common choices:

- **时间单位** — 1 minute of meaningful play. Good for retention/content planning.
- **付费单位** — 1 unit of premium currency (or 1 RMB of value). Good for monetization.

Everything else gets an **exact exchange rate** into the base unit. Once you have
a ledger of rates, you can compare a dungeon reward, a battle-pass tier and a
skin price on one axis. Without it, designers argue in different units and no
amount of spreadsheet work resolves the argument. See `normalization.md`.

### Flow: sources and sinks

A healthy economy has sources (income) and sinks (consumption) in balance:
`source/sink ratio ≈ 1`. Sustained ratio > 1 means inflation — currency accumulates,
prices drift, rewards lose meaning. Sustained < 1 means scarcity — progression
stalls and players feel punished. Every new system must declare both its source
and its sink; a source without a sink is a time bomb.

### Pacing: the derivative matters more than the value

Players perceive **change**, not absolute values. A reward that grows smoothly
feels fair; a cliff feels broken; a flat curve feels dead. Design the *rate of
change* first (the step), then the base. This is why additive curves work: the
derivative is constant and obvious.

## Why additive before multiplicative (加法优先)

`value(n) = base + step * (n - 1)`, capped. The increment is constant, so:

- The marginal value of the next level is known and tunable.
- Bounding is trivial: cap the total.
- Extrapolation to level 1000 is linear and safe.
- A designer can reason about it in their head during a live incident.

Multiplication compounds. Even a benign-looking `value(n) = base * (1.1)^n`
becomes ~1.7× at level 6, ~6× at level 20, ~45× at level 40. In a live game that
is power creep, runaway economy, and unpayable pay gaps — all from "just a small
multiplier".

### Industry pitfalls (real failure modes)

| Pitfall | Mechanism | Prevention |
|---|---|---|
| 数值膨胀 / power creep | Multipliers stack across systems | Additive buckets; cap total; audit growth ratio |
| 边界爆炸 | `base * growth^n` with no cap | Hard caps + piecewise bands |
| 复利经济崩溃 | Income earns income repeatedly | Additive income; sink scales with source |
| 概率期望偏移 | Multiplying independent drop modifiers | Sum % additively, then one roll; verify expected values |
| 付费差距失控 | Whale advantage compounds | Cap additive advantage; normalize to base unit |
| 调试不可预测 | Hidden multiplicative order | One documented order of operations; additive first |
| 战力虚高 | Power score uses products of stats | Sum normalized contributions |
| 面板与目标脱节 | Reward grows faster than sink | Validate reward ≤ sink growth in the same ledger |

### When multiplication is legitimate

Multiplication is allowed, but only under discipline:

1. **One controlled multiplier**, applied once, with a hard cap.
2. **Percentage buckets**: sum all additive % modifiers first (e.g. +10% + 20% =
   +30%), then apply a single `×1.3`. Never `×1.1 × 1.2` chained blindly.
3. **Piecewise bands** instead of a global exponent: level 1–20 step A, 21–50
   step B, etc. Each band is still additive.
4. **Probability** is the classic legitimate case, but expected value must be
   computed and validated, not assumed.

Rule of thumb: if you cannot state the marginal value of the next unit in one
sentence, the formula is too complex.

## Can game lifespan be judged from numbers?

**Honest answer: numbers give strong leading indicators, not a verdict.**

What numbers *can* tell you:

- **Retention decay slope** — D1/D7/D30 trending down predicts shrinking DAU.
- **Content burn vs production** — if `content_debt_ratio` rises and cadence is
  slower than consumption, the game will run out of content on a known date.
- **Economy entropy** — persistent inflation drift means the reward loop is
  losing meaning; players will disengage even if the content is good.
- **Progression saturation** — when most players hit the cap and have nothing to
  chase, lifetime shortens unless a new sink/goal is added.
- **Monetization concentration** — extreme LTV concentration signals a fragile
  revenue base and high churn risk if whales leave.

What numbers *cannot* tell you: whether the game is fun, whether its social
graph is sticky, whether a creative update will reignite it. Those are exactly
the things codegraph233 flow analysis and playtesting are for. Report the
indicator **and** its limit — never claim "this game will die in 6 months"
purely from a spreadsheet.

The practical stance: use numerical health as an **early-warning system** and a
**design constraint**, then let humans/agents judge the qualitative layer.
