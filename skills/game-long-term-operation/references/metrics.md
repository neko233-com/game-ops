# Universal metrics reference

These eleven keys are the vocabulary of the engine. A genre never adds new
metrics — it only changes a metric's `Ideal`, `Min`, `Max` and `Weight`.

Every lever defines a tolerance band `[Min, Max]` centered near `Ideal`.
`fitness = clamp(1 - |observed-ideal| / ((max-min)/2), 0, 1)`: 1 at ideal, 0 at
the band edges and beyond. `Status` is `critical` if outside `[Min,Max]`, `warn`
if fitness `< 0.6`, otherwise `ok`.

## Retention

| Key | Meaning | Typical ideal band (broad) |
|---|---|---|
| `retention_d1` | Share of new users active next day | 0.15 – 0.70 |
| `retention_d7` | Share active on day seven | 0.05 – 0.40 |
| `retention_d30` | Share active on day thirty | 0.02 – 0.20 |

Low retention is usually a first-session/onboarding or content-pacing problem,
not a monetization problem. Fix before optimizing revenue.

## Economy

| Key | Meaning | Healthy value |
|---|---|---|
| `economy_source_sink_ratio` | Currency generated ÷ consumed | 1 (balanced) |
| `economy_inflation` | Price-level index | 1 (stable); > 1 means prices rising |

Ratio persistently above 1 means too much currency is entering the economy —
add sinks or reduce sources. Below 1 means scarcity that stalls progression.

## Monetization

| Key | Meaning | Notes |
|---|---|---|
| `monetization_pay_rate` | Share of users who pay | Genre strongly affects the ideal |
| `monetization_arpu` | Average revenue per user (normalized index) | Use as index, not raw currency |
| `monetization_ltv` | Lifetime value relative to a baseline | High-LTV genres (SLG/gacha) tune higher |

Never raise pay rate by damaging retention; the score's weighting exists so a
monetization gain cannot silently offset a retention collapse.

## Progression

| Key | Meaning | Healthy value |
|---|---|---|
| `progression_pace` | Actual ÷ intended consumption speed | 1 (on target) |

Above 1 means players burn content too fast (raise content cadence or add depth);
below 1 means progression friction.

## Content cadence

| Key | Meaning | Healthy value |
|---|---|---|
| `content_cadence_days` | Days between content drops | Genre-specific (7 for MMO, 30+ for sandbox) |
| `content_debt_ratio` | Backlog ÷ consumption | 1 (healthy) |

Cadence and debt are the long-term levers: chaotic cadence is a leading cause of
D30 decay even when D1 looks fine.

## Acting on a report

Sort diagnostics by `critical` first, then by `deviation * weight` descending.
Change only the underlying system that drives that metric, re-snapshot after one
full content cadence, and re-evaluate. Never edit the report by hand.
