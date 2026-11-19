# Genre coverage reference

Coverage is not a closed enum. The engine ships a broad built-in registry and
exposes `genre.Register(Genre)` so any current or future category can be added
from data with the same eleven metric keys.

## Built-in categories

- RPG family: `rpg`, `action-rpg`, `mmo`, `mmorpg`, `roguelike`
- Strategy: `slg`, `strategy`, `tower-defense`
- Competitive: `moba`, `fps`, `tps`, `battle-royale`, `fighting`, `sports`, `racing`
- Collection: `card`, `gacha`
- Casual: `casual`, `puzzle`, `match3`, `idle`
- World/simulation: `simulation`, `sandbox`, `survival`
- Action/other: `platformer`, `rhythm`, `visual-novel`, `adventure`, `horror`

Check `genre.List()` at runtime for the current set.

## Baseline levers (every genre starts here)

All 29 built-ins start from this exact set; a genre overrides only the rows it
cares about. Bands are deliberately broad; tune them from live data.

| Metric key | Ideal | Min | Max | Weight |
|---|---|---|---|---|
| `retention_d1` | 0.40 | 0.15 | 0.70 | 1.0 |
| `retention_d7` | 0.18 | 0.05 | 0.40 | 1.0 |
| `retention_d30` | 0.08 | 0.02 | 0.20 | 1.0 |
| `economy_source_sink_ratio` | 1 | 0.7 | 1.4 | 1.0 |
| `economy_inflation` | 1 | 0.8 | 1.3 | 1.0 |
| `monetization_pay_rate` | 0.03 | 0.01 | 0.15 | 1.0 |
| `monetization_arpu` | 1 | 0.5 | 3 | 0.8 |
| `monetization_ltv` | 5 | 1 | 20 | 0.8 |
| `progression_pace` | 1 | 0.6 | 1.6 | 0.8 |
| `content_cadence_days` | 14 | 7 | 45 | 0.8 |
| `content_debt_ratio` | 1 | 0.5 | 2 | 0.6 |

## What differs between genres

All genres start from one baseline lever set and override only what matters:

- **Gacha / card / SLG**: higher `monetization_pay_rate` and `monetization_ltv`
  ideals and weights (e.g. gacha pay-rate ideal `0.06`, weight `1.3`).
- **MMO / MMORPG**: faster `content_cadence_days` (7) and higher D30 weight —
  retention is the lifeblood of a persistent world.
- **Casual / idle**: lower pay-rate ideal, higher D1 weight; monetization is a
  long tail, not the primary early signal.
- **Sandbox**: slower cadence tolerated (30) because user-generated content
  supplies depth; lower cadence weight.

## Registering a new genre

```go
err := genre.Register(genre.Genre{
    ID:   "my-genre",
    Name: "My Genre",
    Tags: []string{"new"},
    Levers: []genre.Lever{
        {Key: genre.MetricRetentionD1, Ideal: num.MustFromString("0.35"),
         Min: num.Zero(), Max: num.One(), Weight: num.One()},
        // ... add a lever for every metric you will observe
    },
})
```

Rules:
- `ID` must be non-empty and unique; duplicates error.
- Include a lever for **every** metric you will put in the snapshot — the
  evaluator errors on a missing value rather than assuming zero.
- `Max > Min` and `Weight > 0` are required, otherwise scoring errors.
- Treat levers as data to be tuned, not code to be forked.
