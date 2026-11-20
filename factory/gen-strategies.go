//go:build ignore

// gen-strategies.go — batch-generate one Strategy file per genre subdirectory.
// Run: go run gen-strategies.go
// Not included in the final build.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// StrategySpec is everything one strategy needs to render.
type StrategySpec struct {
	ID       string
	EName    string
	CName    string
	Notes    string   // short design note
	Metrics  []Metric // genre-specific extra metrics
	Advices  []string // default advice bullets
	RiskExpr []RiskExpr
}

type Metric struct {
	Key   string
	Label string
	Desc  string
}

type RiskExpr struct {
	Metric string
	// Observed > Ideal → warn, Observed < Ideal → warn, etc.
	Op    string // "gt" | "lt"
	Ideal string
	Key   string
	MsgEn string
	MsgZh string
}

// spec 为每个品类定义完整策略。
var spec = []StrategySpec{
	{
		ID: "gacha", EName: "Gacha Collection", CName: "抽卡/卡池",
		Notes: "Probabilistic banner pulls. Exact rational probability for drop rates, pity, and expected SSR pulls. CN legal 80-pity floor; US no legal cap — design can set one anyway.",
		Metrics: []Metric{
			{"ssrPity", "SSR 保底抽数", "Total pulls before pity kick-in"},
			{"ssrRate", "SSR 掉率", "Probability per pull (exact rational)"},
			{"ssrExpectedPulls", "期望值", "Expected pulls per SSR including pity"},
			{"dupeRate", "重复率", "Probability a pull is already-owned"},
			{"currencyPerPull", "单抽价格", "Base-unit value of one pull"},
		},
		Advices: []string{
			"保底抽数应控制在 80–120；CN 法律要求保底，海外可自由设计",
			"SSR 期望值 ≤ 75 × 掉率（80 保底 + 0.6% = 133 期望，偏长）",
			"抽卡总价值 ÷ 产出价值 ≤ 1.5× 同类活动（避免挤压）",
			"用 numcalc geom 精确算期望值，不要用 float64",
		},
	},
	{
		ID: "slg", EName: "Strategy / 4X", CName: "策略/4X/国战",
		Notes: "Alliance-focused metagame. Pay rate ceiling must be low (CN ~1–3%). D30 retention is the leading indicator. Alliance activity density drives content burn-rate.",
		Metrics: []Metric{
			{"allianceActivity", "联盟活跃密度", "% players on alliance tasks per day"},
			{"allianceSize", "联盟规模", "Average members per active alliance"},
			{"powerInflation", "战力通胀", "Top-100 mean power ÷ top-1% mean power"},
			{"warSeasonLength", "赛季时长", "Days per full war season"},
			{"resourcePerHour", "资源产出/小时", "Alliance-wide output per real hour"},
		},
		Advices: []string{
			"CN SLG pay rate ≤ 2%；海外可到 3–5%，但绝对不要高于 6%",
			"战力通胀 top-100/top-1% ≤ 200×；高于则新玩家无法竞争",
			"活跃联盟密度 ≥ 40% 玩家/天；低于 30% 内容消耗会卡住",
			"赛季长度必须固定（14–28 天）并在公告中预告 7 天以上",
		},
	},
	{
		ID: "mmo", EName: "Massively Multiplayer Online", CName: "MMO/大型多人",
		Notes: "Server economy + social guilds. Content cadence is 7–14 days. No pity; progression is linear. Guilds must hit activity density targets.",
		Metrics: []Metric{
			{"guildActivity", "公会活跃密度", "% members online per day"},
			{"serverPop", "服务器人口", "Peak concurrent users per server"},
			{"auctionVolume", "拍卖行成交额", "Daily base-unit GMV"},
			{"lvlCapDays", "满级天数", "Days for normal player to hit cap"},
			{"raidClears", "Raid 通关率", "% raids cleared in first attempt"},
		},
		Advices: []string{
			"服务器人口 2000–8000 是稳定区间；低于 1000 开合并流程",
			"满级天数 30–90 天；用 progression.Run 做精确计算",
			"公会活跃密度 ≥ 30%；低于 20% 会长会流失",
			"拍卖行手续费率 5–10%；高于 15% 抑制玩家交易",
		},
	},
	{
		ID: "mmorpg", EName: "MMORPG", CName: "MMORPG",
		Notes: "MMO + RPG progression. LTV heavy; ARPU moderate. Content cadence 7 days. World bosses and guild raids drive retention.",
		Metrics: []Metric{
			{"guildSize", "公会规模", "Avg members"},
			{"raidWeekly", "周 raid 次数", "Raids per guild per week"},
			{"worldBoss", "世界 BOSS 参与率", "% players that join weekly"},
			{"raidClearRate", "Raid 通关率", "% raids cleared"},
			{"lvlCapDays", "满级天数", "Days for normal player"},
		},
		Advices: []string{
			"Raid 周频率 2–4 次；低于 2 次公会内容不足，高于 4 次变成工作",
			"世界 BOSS 参与率 ≥ 25% 说明活动组织正常",
			"新团本首通周期 3–6 周，不要提前宣布",
			"满级天数 60–120 天；偏长对轻玩家是 barrier",
		},
	},
	{
		ID: "moba", EName: "MOBA", CName: "MOBA/竞技",
		Notes: "Esports-driven. Content cadence 21–60 days (hero releases). Pay rate must stay low — cosmetics only. Matchmaker balance is #1.",
		Metrics: []Metric{
			{"matchTime", "对局时长", "Minutes per normal match"},
			{"queueTime", "排队时长", "Seconds in queue before match"},
			{"winratePct", "胜率分布", "% within ±50–1"},
			{"heroRelease", "新英雄间隔", "Days between hero releases"},
			{"cosmeticARPU", "皮肤 ARPU", "ARPU from cosmetics (no gameplay pay) "},
		},
		Advices: []string{
			"Pay rate ≤ 3%；核心玩法必须全免费，钱只能卖皮肤",
			"新英雄间隔 21–30 天；太短平衡不上",
			"对局时长 15–25 分钟；不要超过 35 分钟",
			"排队时长 ≥ 60s 就必须开跨服匹配或扩容服务器",
		},
	},
	{
		ID: "fps", EName: "First-Person Shooter", CName: "FPS/射击",
		Notes: "Cosmetic-only monetization, high skill variance. Content cadence 21–60 days. Anti-cheat signals are leading indicators.",
		Metrics: []Metric{
			{"matchTime", "对局时长", "Minutes per match"},
			{"queueTime", "排队时长", "Seconds before match"},
			{"cheatBanRate", "外挂封禁率", "Bans per 1000 players per day"},
			{"eventCadence", "活动间隔", "Days between major events"},
			{"battlePassAdoption", "战令购买率", "% players buy battle pass"},
		},
		Advices: []string{
			"Pay rate ≤ 2%；禁止任何战力购买",
			"外挂封禁率 > 0.5%/天 说明反作弊正常；> 2% 说明漏洞严重",
			"战令购买率 ≥ 15% 说明内容吸引力足够",
			"匹配时长 ≥ 120s 就必须开跨服",
		},
	},
	{
		ID: "tps", EName: "Third-Person Shooter", CName: "TPS/第三人称射击",
		Notes: "Similar to FPS but with tactical cover mechanics. Content cadence 14–60 days.",
		Metrics: []Metric{
			{"matchTime", "对局时长", "Minutes per match"},
			{"eventCadence", "活动间隔", "Days"},
			{"battlePassAdoption", "战令购买率", "%"},
		},
		Advices: []string{
			"和 FPS 共用战令和皮肤池；不要重复制作",
			"掩体系统平衡：掩体覆盖率 30–40% 最优",
		},
	},
	{
		ID: "battle-royale", EName: "Battle Royale", CName: "大逃杀",
		Notes: "Cosmetic-only, 60–100 player servers. Content cadence 21–60 days. Battle pass 100% of revenue.",
		Metrics: []Metric{
			{"matchTime", "对局时长", "Minutes"},
			{"dropSize", "房间人数", "Players per match"},
			{"battlePassAdoption", "战令购买率", "%"},
			{"resurrectionRate", "复活率", "% players revive teammates"},
		},
		Advices: []string{
			"房间人数 60–100；过少地图空荡，过多服务器扛不住",
			"对局时长 15–20 分钟；超过 25 分钟退潮率 > 30%",
			"战令购买率 ≥ 20% 健康",
		},
	},
	{
		ID: "card", EName: "Collectible Card Game", CName: "卡牌",
		Notes: "Gacha-backed card collection. Drop rates + pity. Rotation cadence 90–180 days.",
		Metrics: []Metric{
			{"setCycleDays", "卡包周期", "Days between new set rotation"},
			{"cardDensity", "卡池密度", "Cards per set"},
			{"wildcardRate", "万能率", "Cards that can substitute any card"},
			{"payRate", "付费率", "% players that buy packs"},
			{"expectedSSR", "期望 SSR", "Pity-adjusted expected SSR per 100 pulls"},
		},
		Advices: []string{
			"卡包周期 90–180 天；太短环境动荡，太长内容枯竭",
			"Pay rate ≤ 5%；高于 10% 玩家流失严重",
			"用 numcalc binom/geom 精确算期望值",
		},
	},
	{
		ID: "casual", EName: "Casual", CName: "休闲",
		Notes: "Low commitment (≤10 min / session). Short retention chains. Ads + IAP. CN D1 ≥ 35% target; US 30%; SEA 40%.",
		Metrics: []Metric{
			{"sessionLength", "单次时长", "Minutes per session"},
			{"adIntensity", "广告密度", "Ads per 60 minutes play"},
			{"iapMix", "IAP 占比", "IAP revenue ÷ (IAP + Ads)"},
			{"arpm", "ARPM", "Ads revenue per 1000 impressions"},
		},
		Advices: []string{
			"D1 留存：CN ≥ 35%，US ≥ 30%，SEA ≥ 40%；低于 25% 核心 loop 有问题",
			"广告密度 2–4 次/小时；高于 5 次卸载率飙升",
			"ARPM 目标 US ≥ $8、EU ≥ $6、SEA ≥ $2",
		},
	},
	{
		ID: "match3", EName: "Match-3", CName: "三消",
		Notes: "Casual + puzzle hybrid. Level-progression gating (life system, stamina, boosters). Event cadence 7–14 days.",
		Metrics: []Metric{
			{"staminaLimit", "体力上限", "Max stamina per day"},
			{"lifeCycle", "生命周期", "Minutes per life regen"},
			{"boosterPrice", "强化道具价格", "IAP cost of top booster"},
			{"clearRate", "通关率", "% players clear hard stages"},
			{"eventCadence", "活动间隔", "Days between themed events"},
		},
		Advices: []string{
			"活动间隔 ≤ 14 天；超过 21 天内容耗尽",
			"Hard stage 通关率 ≥ 5%；低于 2% 关卡数值爆了",
			"体力上限 60–120；低于 40 玩家卡关会退，高于 240 弱化广告驱动",
		},
	},
	{
		ID: "idle", EName: "Idle / Incremental", CName: "放置/增量",
		Notes: "Passive progression. IAP-driven boosters. D30 retention lower bound. CN 0.06; SEA 0.05; US 0.04.",
		Metrics: []Metric{
			{"offlineMult", "离线倍率", "Production multiplier for offline time"},
			{"resetDays", "重置周期", "Days between prestige/ascension"},
			{"boosterLifeDays", "强化道具有效期", "Days a booster persists"},
			{"passiveShare", "被动占比", "% total XP from offline"},
		},
		Advices: []string{
			"离线倍率 ≤ 2×；高于 2× 玩家只挂机不玩，广告收入崩",
			"重置周期 30–90 天；太短重置疲劳，太长内容撑不住",
			"被动 XP 占比 ≤ 70%；高于就不再是放置玩法了",
		},
	},
	{
		ID: "simulation", EName: "Simulation", CName: "模拟经营",
		Notes: "Economy-focused. Content cadence 30–90 days. Balance resource chains; avoid inflation drift.",
		Metrics: []Metric{
			{"currencyTypes", "货币种类", "Total distinct resource currencies"},
			{"tradeRate", "汇率", "Major currency ÷ minor currency (exact)"},
			{"inflationDrift", "通胀漂移", "Price index 90-day change"},
			{"contentCadence", "内容更新间隔", "Days"},
		},
		Advices: []string{
			"货币种类 ≤ 5 种；超过 6 种玩家记不住",
			"通胀漂移 ≤ 10% / 季度；高于 15% 必须强制 rebalance",
			"内容更新 45–90 天；不要超过 120 天",
		},
	},
	{
		ID: "sandbox", EName: "Sandbox", CName: "沙盒",
		Notes: "User-generated content. Content cadence 30–90 days (low). Social share is the #1 acquisition channel.",
		Metrics: []Metric{
			{"ugcShare", "UGC 占比", "% time players spend on user content"},
			{"hostPopulation", "主服人口", "Concurrent on official server"},
			{"multiplayer", "多人参与率", "% players join multiplayer sessions"},
		},
		Advices: []string{
			"UGC 占比 ≥ 50% 说明玩家自发活跃",
			"多人参与率 ≥ 30%；低于 20% 说明单人体验太重",
			"主服人口 2000–8000 稳定；低于 1000 玩家会转私服",
		},
	},
	{
		ID: "survival", EName: "Survival", CName: "生存",
		Notes: "Craft-first, roguelike-adjacent progression. Content cadence 30–90 days. Server merging leads to retention spikes.",
		Metrics: []Metric{
			{"wipeCycle", "重置周期", "Days between world wipes"},
			{"raidFrequency", "PvP raid 频率", "% players participate per week"},
			{"deathPunish", "死亡惩罚", "% lost progress on death"},
		},
		Advices: []string{
			"重置周期 30–90 天；太短玩家没积累感，太长新玩家进不去",
			"死亡惩罚 10–30%；0% 没压力，≥ 50% 卸载",
			"PvP raid 参与率 ≥ 40%",
		},
	},
	{
		ID: "roguelike", EName: "Roguelike / Roguelite", CName: "肉鸽",
		Notes: "Procedural dungeons, roguelite meta progression. Content cadence 21–60 days. Build diversity is the #1 depth metric.",
		Metrics: []Metric{
			{"buildCount", "构建数量", "Distinct meta build archetypes"},
			{"clearTime", "通关时长", "Minutes per successful run"},
			{"metaProgDays", "元进度周期", "Days to unlock every meta perk"},
			{"clearRate", "通关率", "% runs cleared (blind) vs expected"},
		},
		Advices: []string{
			"构建种类 ≥ 15；低于 8 深度不足",
			"元进度周期 30–60 天；太短 meta 疲劳，太长新玩家卡关",
			"通关率 ≈ 5–10%；高于 15% 太浅，低于 2% 太难",
		},
	},
	{
		ID: "platformer", EName: "Platformer", CName: "平台跳跃",
		Notes: "Single-player campaign + speedrun. Content cadence 60–180 days (long).",
		Metrics: []Metric{
			{"runLength", "主线时长", "Hours to 100%"},
			{"deathRate", "死亡密度", "Deaths per 100 sessions"},
			{"%100", "全通率", "% players hit 100%"},
		},
		Advices: []string{
			"主线时长 8–12 小时",
			"死亡密度 ≥ 5/100 sessions 说明关卡设计合理",
		},
	},
	{
		ID: "fighting", EName: "Fighting", CName: "格斗",
		Notes: "1v1 2D fighter. Balance patches every 30–60 days. EVO-caliber tier list changes drive scene health.",
		Metrics: []Metric{
			{"patchCadence", "平衡补丁", "Days"},
			{"tierSpread", "平衡度", "% characters in top 3 tiers"},
			{"tournamentSize", "赛事规模", "Online tournament peak players"},
		},
		Advices: []string{
			"平衡补丁 30–60 天；超过 90 天 scene 萎缩",
			"Top-3 tier 覆盖率 ≥ 50% 角色；低于 30% 太偏",
		},
	},
	{
		ID: "sports", EName: "Sports", CName: "体育",
		Notes: "Real-world league schedules sync. Content cadence tied to real-world seasons (365 days).",
		Metrics: []Metric{
			{"seasonDays", "赛季时长", "Days per in-game season"},
			{"eventSync", "赛事同步", "Align with real-world leagues"},
		},
		Advices: []string{
			"赛季时长严格同步真实世界（180–365 天）",
			"实时赛事同步误差 ≤ 3 天",
		},
	},
	{
		ID: "racing", EName: "Racing", CName: "赛车",
		Notes: "Sim or arcade. Balance patches tied to real-world car releases. Content cadence 60–180 days.",
		Metrics: []Metric{
			{"patchCadence", "平衡补丁", "Days"},
			{"carBalance", "车款平衡", "Win-rate variance across cars"},
		},
		Advices: []string{
			"车款胜率方差 ≤ ±3% 健康",
		},
	},
	{
		ID: "rhythm", EName: "Rhythm", CName: "音乐节奏",
		Notes: "Chart design is the #1 skill metric. Content cadence 14–45 days.",
		Metrics: []Metric{
			{"chartCadence", "谱面间隔", "Days between new charts"},
			{"diffSpread", "难度分布", "% charts per difficulty"},
		},
		Advices: []string{
			"谱面间隔 14–30 天；超过 45 天内容枯竭",
			"Easy/Normal/Hard 覆盖率 30/40/30 最优",
		},
	},
	{
		ID: "tower-defense", EName: "Tower Defense", CName: "塔防",
		Notes: "Wave-based, build-phase and combat-phase separation. Content cadence 21–60 days.",
		Metrics: []Metric{
			{"waveCount", "波次", "Waves per level"},
			{"clearRate", "通关率", "% clear all waves"},
			{"towerBalance", "塔种类平衡", "Win-rate variance"},
		},
		Advices: []string{
			"每关 15–30 波；太少深度不够，太长疲劳",
			"通关率 5–20% 健康",
		},
	},
	{
		ID: "visual-novel", EName: "Visual Novel", CName: "视觉小说",
		Notes: "Single-player narrative. Content cadence 90–365 days. Pay rate 3–4% (premium chapters).",
		Metrics: []Metric{
			{"sceneCount", "场景数", "Scenes per chapter"},
			{"chapterCadence", "章节间隔", "Days"},
			{"premiumPrice", "价格", "Base-unit cost of premium chapters"},
		},
		Advices: []string{
			"章节间隔 90–180 天；太长粉丝流失",
			"Premium 价格 3000–6000 base units",
		},
	},
	{
		ID: "adventure", EName: "Adventure", CName: "冒险",
		Notes: "Story + exploration + light combat. Content cadence 60–180 days.",
		Metrics: []Metric{
			{"playtimeHours", "主线时长", "Hours"},
			{"%Complete", "全通率", "% players hit 100%"},
		},
		Advices: []string{
			"主线时长 10–20 小时",
			"全通率 ≥ 15%",
		},
	},
	{
		ID: "horror", EName: "Horror", CName: "恐怖",
		Notes: "Atmospheric survival + jump scares. Content cadence 60–180 days.",
		Metrics: []Metric{
			{"playtimeHours", "时长", "Hours"},
			{"%Complete", "全通率", "%"},
		},
		Advices: []string{
			"主线时长 6–15 小时",
			"Jump scare density ≤ 3/hour 避免疲劳",
		},
	},
	{
		ID: "strategy", EName: "Strategy", CName: "策略",
		Notes: "RTS or turn-based. Balance patches 30–60 days. Unit variance drives scene health.",
		Metrics: []Metric{
			{"patchCadence", "平衡补丁", "Days"},
			{"unitVariance", "单位平衡", "Win-rate variance"},
		},
		Advices: []string{
			"单位胜率方差 ≤ ±2% 健康",
			"平衡补丁 30–60 天",
		},
	},
	{
		ID: "rpg", EName: "Role-Playing", CName: "RPG",
		Notes: "Progression-heavy. No pity; linear or piecewise curves. LTV 8 ideal.",
		Metrics: []Metric{
			{"lvlCapDays", "满级天数", "Days for normal player"},
			{"mainStoryHours", "主线时长", "Hours"},
			{"classBalance", "职业平衡", "Win-rate variance"},
			{"sideQuestCount", "支线数量", "Side quests"},
		},
		Advices: []string{
			"满级天数 90–180 天",
			"职业胜率方差 ≤ ±3%",
		},
	},
	{
		ID: "action-rpg", EName: "Action RPG", CName: "动作 RPG",
		Notes: "Hack-and-slash + progression. Similar to RPG but real-time combat. Content cadence 21–60 days.",
		Metrics: []Metric{
			{"lvlCapDays", "满级天数", "Days"},
			{"mainStoryHours", "主线时长", "Hours"},
			{"comboCeiling", "连段上限", "Max hits per combo"},
		},
		Advices: []string{
			"连段上限 ≤ 20 hits 避免 input reading 作弊",
			"新内容间隔 21–45 天",
		},
	},
	{
		ID: "puzzle", EName: "Puzzle", CName: "解谜",
		Notes: "Logic-first single-player. Content cadence 14–45 days. Ads + IAP.",
		Metrics: []Metric{
			{"puzzleCount", "关卡数", "Total levels"},
			{"eventCadence", "活动间隔", "Days"},
		},
		Advices: []string{
			"总关卡 ≥ 300；低于 200 消耗太快",
			"活动间隔 14–30 天",
		},
	},
}

// renderStrategy outputs the complete .go file for one strategy spec.
// pkgName is the sanitized package name (hyphens → underscores).
func renderStrategy(s StrategySpec, pkgName string) string {
	var b stringsBuilder
	fmt.Fprintf(&b, `// Package %s implements the strategy for genre %s (%s).
// Auto-generated by gen-strategies.go — edit the spec if you need more.
package %s

import (
	"github.com/neko233-com/game-ops/factory"
)

func init() { factory.Register(strategy{}) }

type strategy struct{}

func (strategy) ID() string   { return %q }
func (strategy) Name() string { return %q }

func (strategy) GenreMetrics() []factory.MetricLabel {
	return []factory.MetricLabel{
`, pkgName, s.EName, s.CName, pkgName, s.ID, s.EName+" / "+s.CName)

	for _, m := range s.Metrics {
		fmt.Fprintf(&b, "\t\t{Key: %q, Label: %q, Desc: %q},\n", m.Key, m.Label, m.Desc)
	}
	fmt.Fprintf(&b, "\t}\n}\n\nfunc (strategy) Advice() []string {\n\treturn []string{\n")
	for _, a := range s.Advices {
		fmt.Fprintf(&b, "\t\t%q,\n", a)
	}
	fmt.Fprintf(&b, "\t}\n}\n\nfunc (strategy) Evaluate(ctx factory.StrategyContext) (factory.StrategyResult, error) {\n\tres := factory.StrategyResult{\n\t\tGenre:  %q,\n\t\tRegion: ctx.Region,\n\t\tNotes: %s,\n\t}\n\treturn res, nil\n}\n", s.ID, "`"+s.Notes+"`")

	return b.String()
}

// sanitizePkgName converts a genre ID into a valid Go package name.
// hyphens → underscores; everything else is already valid (alphanumeric).
func sanitizePkgName(id string) string {
	out := make([]byte, len(id))
	for i := 0; i < len(id); i++ {
		c := id[i]
		if c == '-' {
			out[i] = '_'
		} else {
			out[i] = c
		}
	}
	return string(out)
}

// stringsBuilder is minimal — avoid importing bytes just for this.
type stringsBuilder struct {
	buf []byte
}

func (b *stringsBuilder) Write(p []byte) (int, error) {
	b.buf = append(b.buf, p...)
	return len(p), nil
}
func (b *stringsBuilder) String() string { return string(b.buf) }

func main() {
	outDir := "factory"
	if err := os.MkdirAll(outDir, 0755); err != nil {
		panic(err)
	}

	// Sort for deterministic output.
	sort.Slice(spec, func(i, j int) bool { return spec[i].ID < spec[j].ID })

	for _, s := range spec {
		pkg := sanitizePkgName(s.ID)
		sub := filepath.Join(outDir, pkg)
		if err := os.MkdirAll(sub, 0755); err != nil {
			panic(err)
		}
		filename := filepath.Join(sub, "strategy.go")
		content := renderStrategy(s, pkg)
		if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
			panic(err)
		}
		fmt.Printf("generated %s\n", filename)
	}
	fmt.Printf("\nDone: %d strategies generated\n", len(spec))
}
