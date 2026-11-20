# Unit value & normalization (单位价值 / 归一化)

## Why normalize

"Is a 50-gem reward good?" is unanswerable until gems, gold, stamina, battle
power and time share one unit. Normalization converts every resource into a
single **base unit** using **exact** exchange rates. It is the prerequisite for
comparing reward, cost, price and progression on one axis.

## The three rules

1. **One base unit, stated explicitly.** Choose 1 minute of meaningful play or
   1 premium-currency unit. Write it in the design doc.
2. **Exact rational rates only.** Store `"1/3"`, `"10"`, `"0.25"` as strings;
   parse with `num.MustFromString`. Never `float64`.
3. **Transitive conversion via the rate graph.** If 1 diamond = 10 gold and
   1 stamina = 2 gold, then 1 diamond = 5 stamina — derived exactly, never
   entered by hand (hand-entered cross-rates drift).

## Where rates come from (don't guess them)

The rates below (`1 钻石 = 10 金币`, `1 体力 = 2 金币`) are **illustrative only**.
On a real project derive them from observed data, then validate them:

1. **Time anchor**: run `cmd/ingest` over the reward tables; the median gold (or
   base currency) earned per minute of play defines `1 分钟 = X 基准币`.
2. **Premium anchor**: from the shop table, divide the price of the smallest
   premium-currency bundle by its currency amount to get `1 付费币 = Y 法定货币`,
   then convert to the time anchor with the same method.
3. **Activity currencies**: for each activity currency, take the median ratio of
   its per-minute output to the base unit's per-minute output.
4. **Validate transitivity**: after entering direct rates, check that every
   derived cross-rate matches the hand-computed expectation. A mismatch means an
   arbitrage loop — fix the rates, never the ledger code.
5. **Re-derive every economy revision**: rates are data owned by the agent, with
   the derived value written next to each rate in the config.

## Go implementation

```go
l, err := normalize.NewLedger("金币", []normalize.Unit{
    {ID: "金币", Symbol: "G", Desc: "base currency"},
    {ID: "钻石", Symbol: "D", Desc: "premium"},
    {ID: "体力", Symbol: "S", Desc: "energy"},
}, []normalize.Rate{
    {From: "钻石", To: "金币", Value: num.MustFromString("10")},
    {From: "体力", To: "金币", Value: num.MustFromString("2")},
})
d, _ := l.Normalize("钻石", num.MustFromString("3"))    // 30 金币
dq, _ := l.Convert("钻石", "体力", num.MustFromString("1")) // 5 体力, exact
```

`normalize.NewLedger` validates unique units, positive rates and no self-rate.
`Convert` uses BFS over the rate graph and errors on unknown units or no path.

CLI spec (`go run ./cmd/normalize spec.json`):

```json
{
  "base": "金币",
  "units": [{"id": "金币", "symbol": "G"}, {"id": "钻石", "symbol": "D"}],
  "rates": [{"from": "钻石", "to": "金币", "value": "10"}],
  "conversions": [{"from": "钻石", "amount": "3"}]
}
```

## Worked example — comparing offers

Base unit: 金币. Rates: 1 钻石 = 10 金币; 1 体力 = 2 金币; 1 分钟有效游戏 ≈ 4 金币.

| Offer | Raw | Normalized | Verdict |
|---|---|---|---|
| Daily quest | 20 金币 | 20 金币 | baseline |
| Event reward | 3 钻石 | 30 金币 | 1.5× daily — reasonable |
| Shop bundle | 50 钻石 | 500 金币 | must check sink/price, not just value |
| Stamina refill | 10 体力 | 20 金币 | equals daily, fair |

Once normalized, the agent can judge **value density** (value per minute of
play, value per RMB) and detect rewards that accidentally dwarf the economy.

## Normalization in the health loop

Use normalized values for:

- `monetization_arpu` / `monetization_ltv` indices (so different SKUs compare).
- `economy_source_sink_ratio` (sum all sources and sinks **in base units**).
- Combat "power" (sum normalized stat contributions — additive, never product).
- Content value per cadence window (is the new content worth the wait?).

Pitfalls: mixing base units mid-analysis; rounding rates; hardcoding cross-rates;
comparing gross value without netting the sink. All are caught by the ledger.
