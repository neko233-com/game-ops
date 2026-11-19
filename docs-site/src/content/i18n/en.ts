/**
 * 英文译文。中文原文（src/content/*.ts）始终是权威完整版本。
 *
 * 前 3 页提供完整正文翻译；其余页面只翻译 nav/title/subtitle，
 * localizePage 会把它们的正文块自动标记为“原文为中文”。
 */

import type { PageTranslation } from "../../types";
import { callout, code, formula, h2, h3, ol, p, quote, steps, table, ul } from "../../types";

const overview: PageTranslation = {
  nav: "Home / Overview",
  title: "The One Way: A Manual for Game Numbers & Live Operations",
  subtitle:
    "This documentation accompanies a pure-Go live-ops engine and an agent skill. It covers the common foundation of all numerical design: how value is defined, how it flows, and at what pace it is consumed and replenished.",
  blocks: [
    p(
      "You may have worked on combat numbers, progression curves, bundle pricing, or event cadence — or you may be using AI to generate hundreds of thousands of config rows in one go. They look like completely different jobs, but underneath they all circle the same question: how much value does a number represent, when does that value flow into and out of the player, and do those flows match? See this clearly, and the differences between RPG, SLG, gacha, casual and sandbox games become differences in parameters, not in principles. That is 'The One Way'.",
    ),
    p(
      "This site is not an API reference, nor a recipe book for a single genre. It is a concept map: first we take the word 'numbers' apart, then we offer a reusable set of judgment tools — normalization, unit value, additive-first, the five pillars and 11 metrics, and control loops. When you finish, you should be able to describe any game's economy and cadence in the same language, and to know when to trust the numbers and when not to.",
    ),
    callout(
      "tip",
      "The three-sentence version",
      "1) Numbers are the definition of value: first pick a base unit and convert every resource into it. 2) Build curves additively first; multiplication is only a controlled final step. 3) Live operations is a control loop of 'observe — score — act — observe again', not a content calendar.",
    ),
    h2("The five-minute quick start"),
    p(
      "If you want to see this method running in a real project right away, walk through the order below. Every step has a matching documentation page you can jump to at any time.",
    ),
    steps(
      {
        title: "Read 'The One Way' (3 min)",
        text: "Establish the big picture: definition, flow and cadence of value. Focus on the honest question at the end — can numbers really predict a game's lifespan?",
      },
      {
        title: "Read 'Normalization & Unit Value' (5 min)",
        text: "Learn to fold every currency, stamina, power and time measure into one base unit, and to use exact rationals rather than floats for exchange rates.",
      },
      {
        title: "Read 'Five Pillars & 11 Metrics' (5 min)",
        text: "Get the general observation vocabulary and the health-score formula. This is what the engine actually computes.",
      },
      {
        title: "Run a report in your project",
        text: "Write a snapshot as JSON (values as strings), hand it to the engine for scoring, and see which metric is critical.",
      },
      {
        title: "Run it again after a content beat",
        text: "One audit is not operations. Re-snapshot and re-score after every content drop; watch the trend, not a single point.",
      },
    ),
    h3("Minimal runnable example"),
    p(
      "The engine's input is a snapshot: a genre id plus one exact value per lever key. Note that every value is written as a string, so 0.1 + 0.2 is still 3/10 after a serialization round-trip.",
    ),
    code(
      "json",
      `{
  "genre_id": "gacha",
  "values": {
    "retention_d1": "0.44",
    "retention_d7": "0.19",
    "retention_d30": "0.07",
    "economy_source_sink_ratio": "1.08",
    "economy_inflation": "1.01",
    "monetization_pay_rate": "0.052",
    "monetization_arpu": "1.10",
    "monetization_ltv": "1.35",
    "progression_pace": "1.21",
    "content_cadence_days": "14",
    "content_debt_ratio": "1.30"
  }
}`,
      "examples/gacha.json — values are strings; precision survives the JSON round-trip",
    ),
    code(
      "bash",
      `go run ./cmd/opsreport examples/gacha.json
# Output: score (exact num.Num), verdict, diagnostics list`,
      "Run from the companion Go repository root",
    ),
    h2("Recommended reading path"),
    table(
      ["Who you are", "Read first", "Then read", "Goal"],
      [
        [
          "New to numerical design",
          "The One Way → Normalization",
          "Additive-first → Combat / Progression / Monetization",
          "Build intuition and be able to design a non-inflating curve on your own",
        ],
        [
          "Already doing live ops",
          "Five Pillars & 11 Metrics",
          "AI-Authored Numbers → Install Across All Runtimes",
          "Turn experience into a recomputable control loop",
        ],
        [
          "AI / agent engineer",
          "Normalization → Additive-first",
          "AI config analysis → Flow analysis",
          "Let the agent design and self-check numbers within constraints",
        ],
        [
          "Just looking up a concept",
          "Jump straight to the page in the sidebar",
          "—",
          "Every page is written to be self-contained and readable alone",
        ],
      ],
    ),
    h2("A few conventions"),
    ul(
      "'Value' is not price. Price is value expressed in a particular market; value itself must first be defined.",
      "'Exact' means rational arithmetic: numerator and denominator are big integers reduced to lowest terms; configuration math never goes to float64.",
      "'Genre-agnostic' means the logic layer only reads metric keys; genre differences live entirely in lever data, not in code branches.",
      "'Project-first' means skills and tools are installed into the project directory and versioned with the repo; a global install is only a convenience.",
    ),
    callout(
      "note",
      "The scope of this site",
      "We discuss the part of design that can be written as numbers and independently checked. Fun, socialising, expression and the urge to create cannot be reduced to metrics; they decide why players stay, while numbers can only tell you what the players who stayed consumed, and how fast. This distinction is spelled out in the 'The One Way' page.",
    ),
    h3("Next step"),
    p(
      "Open 'The One Way: The Core of Numerical Design' from the sidebar. If you only have five minutes, at least read its final section.",
    ),
  ],
};

const oneness: PageTranslation = {
  nav: "The One Way: Core of Numbers",
  title: "The One Way: The Core of Numerical Design",
  subtitle:
    "The essence of numbers is 'the definition, flow and cadence of value'. Genres change, art changes, gameplay changes — these three do not.",
  blocks: [
    quote(
      "The Way begets one; one begets two; two begets three; three begets all things.",
      "Reducing all things back to a single 'one' is the first step of all numerical work.",
    ),
    h2("1. Numbers are not digits; they are relationships"),
    p(
      "The most common beginner mistake is to treat numbers as 'filling in a value for a thing': attack 100, gold drop 20, bundle ¥6. A table filled in this way falls apart after the first content update, because every number is actually answering three questions: what is it worth? Where does it come from and where does it go? How fast does it flow?",
    ),
    p(
      "Separate those three questions and you get the big picture of this manual: numbers = definition of value + flow of value + cadence of costs and rewards. 'The One Way' means folding the bewildering array of resources, currencies, power, time and purchases into one computable language. Once that folding is done, you will find that genres differ only in curve shape and the width of their tolerance bands.",
    ),
    h2("2. Defining value: unit value"),
    p(
      "Before any resource can be compared, it needs a common measure. That measure is 'unit value': choose a base unit, define what one unit of it equals, then express every other resource as a multiple or fraction of it. Without this step, every cross-system comparison is an illusion.",
    ),
    p(
      "Here is the most common illusion: a gem seems worth only ¥0.1, and gold seems almost worthless, so giving away 1,000 gold 'costs nothing'. But if one six-minute battle yields 120 gold, then 1,000 gold equals 50 minutes of standardised play — possibly an entire day of a player's progress. Without converting to unit value, you cannot tell whether a compensation is expensive.",
    ),
    h3("Three properties unit value must satisfy"),
    ol(
      "Anchorable: the base unit must be anchored to something stable and observable; time is the most common anchor because it treats every player alike.",
      "Transitive: multiplying the A→B and B→C conversion rates must equal the direct A→C rate, otherwise an arbitrage loophole exists.",
      "Auditable: conversion rates must be exact rationals stored as strings in config, so any conversion can be re-checked by hand.",
    ),
    h2("3. The flow of value: exchange rates, conservation and inflation"),
    p(
      "With unit value in hand, a game economy can be seen as a flow graph: resources enter the player from 'sources' (stages, quests, purchases) and leave through 'sinks' (upgrades, gacha, repairs, tickets). Describing that graph takes only two numbers: the production rate and the consumption rate.",
    ),
    formula(
      "source_sink_ratio = total produced per unit time ÷ total consumed per unit time",
      "Equal to 1 means balance; persistently above 1 means net inflow, with stockpiles growing in players' hands.",
    ),
    p(
      "This is the engineering approximation of 'value conservation': in a closed system, newly created currency must have a matching sink, or price levels rise. Note that a game economy is not a physical system — it is inherently open: purchases inject value out of thin air, and live events dump it in one-off. So the real goal is not absolute conservation but 'controlled, intentional injection'.",
    ),
    h3("Two sources of inflation"),
    table(
      ["Type", "Mechanism", "Typical symptom", "How to treat it"],
      [
        [
          "Stockpile inflation",
          "Production exceeds sinks over the long run; inventories pile up",
          "Veterans overflow with resources; new content loses appeal",
          "Add sinks (sunk-cost consumption) rather than simply cutting production",
        ],
        [
          "Price inflation",
          "Power / level baselines rise; the same resources buy less progress",
          "Catch-up cost explodes; veterans' assets lose value",
          "Control the speed of baseline rise with caps and segments",
        ],
      ],
      "Both kinds often happen together, but the remedies differ; diagnose which one first.",
    ),
    callout(
      "warn",
      "Do not just cut production",
      "The pain of 'having something taken away' far outweighs 'getting less'. Stockpile inflation is best solved by adding new sinks (cosmetics, collections, long-term progression). Slashing daily production usually causes a retention crash and makes the economy worse.",
    ),
    h2("4. Cadence: time as the ultimate unit"),
    p(
      "If only one base unit could be kept, I would choose time — not because it is the most precise, but because it is the fairest and least substitutable. Money can buy resources, but not more waking hours in a day; a gold farm can print coins, but not players' attention. Convert everything into 'standard effective game minutes' and you can compare combat, progression, purchases and events on one timeline.",
    ),
    p(
      "Time as a unit has a second benefit: it naturally gives you a way to measure cadence. Production rate is 'how many SU earned per minute', consumption rate is 'how many SU spent per minute', and content-consumption speed is 'how many stages a player burns per minute'. Imbalance then has a very plain definition: the time derivatives of inflow and outflow do not match.",
    ),
    formula(
      "healthy cadence ≈ content production rate ≥ content consumption rate ≈ resource production rate ≈ resource consumption rate",
      "The four rates need not be equal, but they must be of the same order; if any one is an order of magnitude off, trouble will follow.",
    ),
    p(
      "This is also why content cadence is a long-term lever: D1 may look fine, but if content consumption outruns production for long enough, D30 will silently collapse — players leave not because of a bug, but because 'there is nothing left to play'.",
    ),
    h2("5. An honest section: can numbers judge a game's lifespan?"),
    p(
      "Short answer: numbers can give leading indicators, but not a verdict. Treating numbers as fortune-telling is where this method is most easily abused, so this section must spell out its limits.",
    ),
    h3("Four leading signals numbers really can see"),
    table(
      ["Signal", "How to read it from numbers", "What it foretells"],
      [
        [
          "Retention decay",
          "D1/D7/D30 fall exponentially, with the decay rate growing each version",
          "A problem with onboarding or early-game cadence; left alone, the user pool keeps shrinking",
        ],
        [
          "Content deficit",
          "content_debt_ratio above 1 and progression_pace above 1 for a long time",
          "Players consume content faster than the team produces it; a foreseeable content drought",
        ],
        [
          "Economic entropy",
          "source_sink_ratio far from 1 over time; price index rises one-way",
          "Currency loses meaning, rewards stop motivating, long-term spending intent erodes",
        ],
        [
          "Progression saturation",
          "Marginal utility of resources approaches 0 for high-level players; progress curve caps out",
          "The core loop stops working for veterans; what remains is inertia, not motivation",
        ],
      ],
    ),
    p(
      "All four signals satisfy 'necessary': if they keep worsening, a game can hardly live long. But they do not satisfy 'sufficient': a numerically healthy project may die quickly for lack of fun, while a numerically mediocre one may live a long time on social ties, a creator ecosystem or a cultural moment.",
    ),
    h3("Where numbers cannot reach"),
    ul(
      "Fun: game feel, feedback, surprises, humour — these decide whether you want another run, and retention metrics only record that decision afterwards.",
      "Social: guilds, mentorship, rivalry and bonds create stickiness a numerical model cannot explain, and can cause cliff-like churn when the network breaks.",
      "Creation: UGC, mods and player-organised play make content-consumption speed independent of team output, at which point 'content debt' stops being a meaningful metric.",
      "Meaning: why a player comes back every day may be identity, aesthetics or companionship — none of which has a key in the snapshot.",
    ),
    callout(
      "danger",
      "A conclusion we must admit",
      "Numbers are a 'check-up report', not a 'verdict on your fate'. They tell you what is deteriorating and how fast, but not whether the game will be loved. Using the score as the sole decision input systematically cuts designs that look bad on short-term metrics but have real long-term life.",
    ),
    p(
      "So the right usage is: use numbers to find the worst deviation, then change the system driving it with a hypothesis in hand, and measure again at the next content beat. Measure the trend, not fate.",
    ),
    h2("6. Bringing the big picture into the engine"),
    p(
      "The companion Go engine compresses this thinking into a control loop. The logic layer knows no genre names; it only reads metric keys. Genre differences are expressed as lever data.",
    ),
    code(
      "go",
      `g := genre.MustGet("gacha") // or genre.Register(myGenre) first
values := map[string]num.Num{
    genre.MetricRetentionD1: num.MustFromString("0.44"),
    // ... one value per lever; missing keys are errors, never 0
}
report, err := ops.Evaluate(ops.Snapshot{GenreID: g.ID, Values: values})
// report.Score is an exact rational; report.Verdict is healthy/attention/critical`,
    ),
    p(
      "The next page unfolds 'unit value' and 'normalization' into a hands-on example: convert gold, gems, stamina, power and time into a single base unit.",
    ),
  ],
};

const normalization: PageTranslation = {
  nav: "Normalization & Unit Value",
  title: "Normalization & Unit Value",
  subtitle:
    "Fold every resource into one base unit and record exchange rates as exact rationals. This page gives you a complete example you can copy.",
  blocks: [
    h2("1. Why you must choose a base unit first"),
    p(
      "'Is this bundle worth it?' 'Was that compensation expensive?' 'How much should stamina cost?' — without a common unit, the answer can only be a feeling. Normalization exists to destroy that feeling: define one base unit (written SU, Standard Unit, below) and give every resource a coefficient of '1 unit = how many SU'. All comparisons then become addition and subtraction on the same scale.",
    ),
    p(
      "There is a practical rule for choosing the base unit: pick the thing least vulnerable to inflation and easiest for players to understand. For the vast majority of games, that is time. We define 1 SU = one minute of 'standard effective game time', i.e. one minute of standard combat for a target player. This anchor does not drift with content updates and is unaffected by purchases.",
    ),
    callout(
      "note",
      "SU is not a currency",
      "SU is a ruler, not a resource you can hand to players. You will never see SU in an inventory, but every pricing and production decision you make should be checked in SU space.",
    ),
    h2("2. Exchange rates must be exact rationals"),
    p(
      "Config always stores strings: '1/20', '0.05', '3e-2' are all fine, but once parsed into memory the value must be an exact rational (Go's math/big.Rat, or a BigInt numerator/denominator reduced by the GCD). Never use float64 for config math — 0.1 + 0.2 is not 3/10 in binary floating point, and the thousands of accumulations in a config table amplify that error into observable economic drift.",
    ),
    ul(
      "Config layer: values are strings (decimal, fraction or scientific notation) for version control, diffs and human review.",
      "Parse layer: on load, convert to exact rationals, reduce, and do all arithmetic here.",
      "Presentation layer: only convert to float or decimal strings for logs, charts and player-facing text.",
    ),
    code(
      "go",
      `// Don't do this: f := 0.1 + 0.2 // 0.30000000000000004
// Do this instead:
goldPerBattle := num.MustFromString("120")   // 120 gold per battle
minutesPerBattle := num.MustFromString("6")   // 6 minutes per battle
suPerGold := num.MustFromString("1").Quo(
    minutesPerBattle, goldPerBattle,
) // 1 gold = 6/120 SU = 1/20 SU, exactly`,
      "Exact rationals: numerator and denominator reduced, no rounding error",
    ),
    h2("3. Transitivity: exchange rates must close the loop"),
    p(
      "The bug most often hidden in a conversion table is inconsistency: A→B is 2, B→C is 3, but A→C was written as 5. Once players find that path they can arbitrage forever. So every set of rates must be transitive, and that must be an automated validation rule, not a human eyeball check.",
    ),
    formula(
      "rate(A→C) = rate(A→B) × rate(B→C), for any A, B, C",
      "Treat resource nodes as graph vertices and rates as edge weights; the product of the weights along any path must be equal.",
    ),
    p(
      "This rule also gives the design constraint for an 'exchange shop': if an item can be bought with both gold and gems, the ratio between the two prices must equal the SU conversion ratio of gold to gems, or the price difference becomes arbitrage profit. Either align the prices, or make the premium deliberate with explicit purchase limits / cooldowns.",
    ),
    h2("4. Full worked example: five resources into SU"),
    p(
      "Assume a game with the following settings. We use them to convert gold, gems, stamina, power and RMB into SU.",
    ),
    table(
      ["Resource", "Setting", "Derivation", "1 unit = ? SU"],
      [
        ["Standard time", "1 minute of standard combat = 1 SU", "Definition", "1"],
        ["Stamina", "1 stamina buys a 6-minute battle", "6 min ÷ 1", "6"],
        ["Gold", "One battle yields 120 gold", "6 ÷ 120", "1/20"],
        ["Gem", "1 gem = 1/5 stamina (5 gems buy 1 stamina)", "6 × 1/5", "6/5"],
        ["Power", "10 power costs 100 gold + 1 stamina", "100×1/20 + 6 = 11, then ÷10", "11/10"],
        ["RMB", "¥1 = 10 gems", "10 × 6/5", "12"],
      ],
      "All coefficients are rational; power conversion only holds in the 'marginal progression cost' sense.",
    ),
    h3("Checking transitivity"),
    p("Take the gem chain: the direct conversion and the conversion via stamina must be equal."),
    code(
      "text",
      `100 gems →(1/5)→ 20 stamina →(6)→ 120 SU
100 gems →(6/5)→ 120 SU            ✓ consistent

¥1 →(12)→ 12 SU
¥1 →(10 gems)→ 10 →(6/5)→ 12 SU   ✓ consistent`,
    ),
    h3("Using the conversion for a real decision"),
    p(
      "Now answer the opening question: is giving away 1,000 gold expensive? 1000 × 1/20 = 50 SU, i.e. 50 minutes of standard play, roughly eight or nine battles. For a daily-active player that is a solid mid-tier daily reward; for a newcomer it may be two days of progress. Without normalization, all you can say is '1,000 gold doesn't seem like much'.",
    ),
    p(
      "Now purchases: a ¥6 bundle with 60 gems is worth 60 × 6/5 = 72 SU, or 72 minutes of game time. If a player averages 60 minutes a day, that bundle gives just over a day of progress — whether it is cheap or overpowered depends on its place in the price curve, but at least it can now be discussed quantitatively.",
    ),
    h2("5. A special note on power"),
    callout(
      "warn",
      "Power is not a currency",
      "Gold, gems and stamina can be exchanged directly; power cannot. The 11/10 SU above means 'at the current progression stage, raising one point of power costs this much resource value on average'. It is a marginal cost that changes with level brackets. Treating power as an exchangeable currency disconnects power pricing from real investment.",
    ),
    p(
      "So power conversion must be recorded per bracket: one set of coefficients per level range, re-measured on every content update. Comparing raw power numbers across brackets is meaningless — one million power means completely different SU at launch versus version 20.",
    ),
    h2("6. Implementation checklist"),
    ol(
      "Choose a base unit and write down its definition (recommended: 1 SU = one minute of standard effective game time).",
      "List all resources and write '1 unit = ? SU' for each, using only rationals.",
      "Put all rates into config as strings; convert to exact rationals on load.",
      "Write an automated check: for any A→B→C, the rate must equal A→C.",
      "In every review involving cross-resource pricing, convert to SU before discussing.",
      "Convert to float only in the presentation layer; keep exact values in logs and reports.",
    ),
    p(
      "Next: once unit value is understood, you can discuss another basic principle — why, on top of these conversions, curves should be built with addition rather than multiplication.",
    ),
  ],
};

/** 其余页面：只翻译 nav/title/subtitle，正文块由 localizePage 统一标记。 */
const prose = {
  "additive-first": {
    nav: "Why Additive First, Not Multiplicative",
    title: "Why Additive Beats Multiplicative",
    subtitle:
      "Multiplication is not the enemy; stacked multiplication is. Addition keeps things predictable and debuggable; multiplication appears once, in a controlled position.",
  },
  "value-flow": {
    nav: "Value Flow Diagram (Interactive)",
    title: "Value Flow: Where Money Comes From and Goes",
    subtitle:
      "A draggable, clickable diagram that explains 'sources → wallet → sinks'. Look at the picture first, then read.",
  },
  "system-map": {
    nav: "System Map (Interactive)",
    title: "System Map: One Change Moves Everything",
    subtitle:
      "Six links form one causal chain. Click a node to see what it means, flip the switch to compare 'additive-first' with 'multiplicative stacking'.",
  },
  pillars: {
    nav: "Five Pillars & 11 Metrics",
    title: "Five Pillars and 11 Metrics",
    subtitle:
      "Describing a game needs only one shared set of observations. Genres do not change the metrics, only each metric's tolerance band and weight.",
  },
  numbers: {
    nav: "Combat / Progression / Monetization Numbers",
    title: "Combat, Progression and Monetization Numbers",
    subtitle:
      "Apply the previous principles to three concrete curves: how to design combat, how to pace progression, and how to price without destroying retention.",
  },
  "ai-numbers": {
    nav: "AI-Authored Numbers & Million-Scale Configs",
    title: "AI-Authored Numbers and Million-Scale Config Analysis",
    subtitle:
      "Let the agent design all numbers inside human-given constraints, stream a million Excel/JSON/CSV files, and judge the current build's numerical health.",
  },
  flow: {
    nav: "Flow Analysis + codegraph233",
    title: "Flow Analysis and codegraph233 (Optional)",
    subtitle:
      "Turn gameplay code into a queryable graph via AST so the agent can follow the call chain and understand 'what players actually go through'. Slow, optional, and not end-to-end verified in this environment.",
  },
  install: {
    nav: "Install Across All Runtimes",
    title: "Install Across All Runtimes (Project-First)",
    subtitle:
      "Supported in opencode / codex / dsh / claude. Install the skill into the project first, then consider global; copy the source from GitHub and follow the README quickstart.",
  },
  "pro-analytics": {
    nav: "Professional Numerical Analysis (ECharts)",
    title: "Professional Numerical Analysis: Auditable Judgments from Curves to Configs",
    subtitle:
      "A dense view for lead numerical designers: dual axes, legends, threshold lines and exact values. All engine calculations use exact rationals.",
  },
} satisfies Record<string, Pick<PageTranslation, "nav" | "title" | "subtitle">>;

export const enTranslations: Record<string, PageTranslation> = {
  overview,
  oneness,
  normalization,
  ...prose,
};
