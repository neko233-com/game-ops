//go:build ignore
// +build ignore

// gen-strategies.go — batch-generate one Strategy file per genre subdirectory.
// Run: go run factory/gen-strategies.go
//
// Spec drives everything. Each StrategySpec defines:
//   - Metrics            genre-specific metric labels
//   - Advices            default advice bullets (Region-agnostic)
//   - RegionAdvices      additional advice bullets filtered by ctx.Region
//   - Thresholds         threshold table — Region-aware min/max/ideal per metric
//   - Notes              rendered as StrategyResult.Notes
//
// ThresholdTable: map[MetricKey]map[Region]Threshold{Min, Max, Ideal, MsgLow, MsgHigh}
// Any metric not in ctx.Metrics is skipped silently; only observed values get checked.
// Region "GLOBAL" in the table is the fallback.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ---- Data structures ----

type StrategySpec struct {
	ID            string
	EName         string
	CName         string
	Notes         string
	Metrics       []Metric
	Advices       []string
	RegionAdvices []RegionAdvice                  // filtered by ctx.Region at runtime
	Thresholds    map[string]map[string]Threshold // metricKey → region → Threshold
}

type Metric struct {
	Key   string
	Label string
	Desc  string
}

type RegionAdvice struct {
	Region string // "CN" | "SEA" | "US" | "EU" | "GLOBAL" (GLOBAL = always)
	Advice string
}

type Threshold struct {
	MinLow  string // below this → critical
	Min     string // below this → warn
	IdealLo string // ideal lower bound (min acceptable)
	IdealHi string // ideal upper bound (max acceptable)
	Max     string // above this → warn
	MaxHigh string // above this → critical
	MsgLow  string // warn/critical below
	MsgHigh string // warn/critical above
	Unit    string // displayed unit, empty = generic
}

// Short aliases for spec readability.
const (
	CN   = "CN"
	SEA  = "SEA"
	US   = "US"
	EU   = "EU"
	GLOB = "GLOBAL"
)

// mkTh builds a Threshold compactly (unit defaults to "").
func mkTh(idealLo, idealHi, minLo, min, max, maxHi, msgLow, msgHigh string) Threshold {
	return Threshold{minLo, min, idealLo, idealHi, max, maxHi, msgLow, msgHigh, ""}
}

// ---- Spec: 29 genres, each with Region-aware thresholds and advice ----

var spec = []StrategySpec{
	// ============ Gacha ============
	{
		ID: "gacha", EName: "Gacha Collection", CName: "抽卡/卡池",
		Notes: "抽卡是概率门。核心变量：保底抽数、掉率、期望值、重复率。CN 法律强制保底 80；海外可自由设计。",
		Metrics: []Metric{
			{"ssrPity", "SSR 保底抽数", "Total pulls before pity kick-in"},
			{"ssrRate", "SSR 掉率", "Probability per pull (exact rational)"},
			{"ssrExpectedPulls", "期望值", "Expected pulls per SSR including pity"},
			{"dupeRate", "重复率", "Probability a pull is already-owned"},
			{"currencyPerPull", "单抽价格", "Base-unit value of one pull"},
		},
		Advices: []string{
			"保底抽数应控制在 80–120；保底越高期望值越差，付费转化率越低",
			"SSR 期望值 = 几何分布 E[k] + 保底修正。用 numcalc geom 精确算",
			"单抽价值 ÷ 产出价值 ≤ 1.5× 同类活动，避免挤压其它付费点",
			"重复率设计 ≤ 25%；高于则玩家抽新卡包的动力崩",
		},
		RegionAdvices: []RegionAdvice{
			{CN, "CN 法律要求必须公开掉率和保底抽数；不得隐瞒。建议 80 保底"},
			{US, "US 法律无保底要求，但建议设 120 以内保底以提高口碑"},
			{SEA, "SEA 玩家保底期望 ≥ 90，期望值敏感；需用精确概率算抽数"},
			{EU, "EU 与 US 类似，建议 ≤ 120 保底"},
		},
		Thresholds: map[string]map[string]Threshold{
			"ssrPity": {
				CN:   mkTh("80", "80", "", "70", "100", "120", "保底低于法律要求 80", "保底高于 120 期望值过差"),
				GLOB: mkTh("60", "120", "", "50", "150", "180", "保底太短无安全感", "保底太长期望值过差"),
			},
			"ssrRate": {
				CN:   mkTh("0.004", "0.010", "0.002", "0.003", "0.015", "0.025", "掉率太低期望值超过 200 抽", "掉率太高 SSR 变白菜价"),
				GLOB: mkTh("0.004", "0.012", "0.002", "0.003", "0.020", "0.040", "掉率太低期望值超过 200 抽", "掉率太高 SSR 变白菜价"),
			},
			"dupeRate": {
				GLOB: mkTh("0.05", "0.20", "0", "0", "0.30", "0.45", "重复率过低缺少收藏动力", "重复率过高玩家对新卡包失去兴趣"),
			},
		},
	},

	// ============ SLG ============
	{
		ID: "slg", EName: "Strategy / 4X", CName: "策略/4X/国战",
		Notes: "SLG 是联盟驱动的长线对抗。Pay rate 必须低；联盟活跃密度是内容消耗的核心指标。",
		Metrics: []Metric{
			{"allianceActivity", "联盟活跃密度", "% players on alliance tasks per day"},
			{"allianceSize", "联盟规模", "Average members per active alliance"},
			{"powerInflation", "战力通胀", "Top-100 mean power ÷ top-1% mean power"},
			{"warSeasonLength", "赛季时长", "Days per full war season"},
			{"resourcePerHour", "资源产出/小时", "Alliance-wide output per real hour"},
		},
		Advices: []string{
			"战力通胀 top-100/top-1% ≤ 200×；高于则新玩家无法竞争",
			"联盟活跃密度 ≥ 40%；低于 30% 内容消耗会卡住",
			"赛季时长必须固定（14–28 天），公告提前 7 天以上预告",
			"资源产出曲线必须 piecewise 加法；禁止全局指数",
		},
		RegionAdvices: []RegionAdvice{
			{CN, "CN SLG Pay rate ≤ 2%；绝对天花板 6%。联盟密度 ≥ 40%"},
			{US, "US SLG Pay rate 3–5%；绝对天花板 8%。联盟密度 ≥ 35%"},
			{SEA, "SEA SLG Pay rate 2–4%；绝对天花板 7%。联盟密度 ≥ 38%"},
		},
		Thresholds: map[string]map[string]Threshold{
			"allianceActivity": {
				CN:   mkTh("0.40", "0.80", "0", "0.35", "0.95", "1.0", "联盟密度太低，内容消耗卡住", "联盟密度过高，变成打工失去乐趣"),
				GLOB: mkTh("0.30", "0.80", "0", "0.25", "0.95", "1.0", "联盟密度太低，内容消耗卡住", "联盟密度过高，变成打工失去乐趣"),
			},
			"powerInflation": {
				GLOB: mkTh("1.5", "3.0", "", "1.0", "5.0", "10.0", "战力太平没有目标感", "战力通胀太严重，新玩家无法追赶"),
			},
			"warSeasonLength": {
				GLOB: mkTh("14", "28", "", "7", "35", "60", "赛季太短联盟积累不足", "赛季太长新玩家错过窗口"),
			},
		},
	},

	// ============ Casual ============
	{
		ID: "casual", EName: "Casual", CName: "休闲",
		Notes: "休闲的核心是 D1 留存。CN/US/SEA 基准不同。广告密度不能超过 4次/小时。",
		Metrics: []Metric{
			{"sessionLength", "单次时长", "Minutes per session"},
			{"adIntensity", "广告密度", "Ads per 60 minutes play"},
			{"iapMix", "IAP 占比", "IAP revenue ÷ (IAP + Ads)"},
			{"arpm", "ARPM", "Ads revenue per 1000 impressions"},
			{"retention_d1", "D1 留存", "% new users active day 1"},
		},
		Advices: []string{
			"广告密度 2–4 次/小时；高于 5 次卸载率飙升",
			"IAP 占比 30–60% 健康；纯广告或纯 IAP 都有天花板",
			"Session 时长 5–15 分钟；太长不符合休闲定位",
		},
		RegionAdvices: []RegionAdvice{
			{CN, "CN D1 留存基准 ≥ 35%；低于 30% 核心 loop 有问题"},
			{US, "US D1 留存基准 ≥ 30%；低于 25% 核心 loop 有问题"},
			{SEA, "SEA D1 留存基准 ≥ 40%；低于 35% 核心 loop 有问题"},
			{EU, "EU D1 留存基准 ≥ 32%；低于 27% 核心 loop 有问题"},
		},
		Thresholds: map[string]map[string]Threshold{
			"retention_d1": {
				CN:   mkTh("0.35", "0.55", "0.25", "0.30", "0.65", "0.80", "D1 留存低于 CN 基准 35%", "D1 留存高于 55% 需检查数据污染"),
				US:   mkTh("0.30", "0.50", "0.20", "0.25", "0.60", "0.75", "D1 留存低于 US 基准 30%", "D1 留存高于 50% 需检查数据污染"),
				SEA:  mkTh("0.40", "0.60", "0.30", "0.35", "0.70", "0.85", "D1 留存低于 SEA 基准 40%", "D1 留存高于 60% 需检查数据污染"),
				GLOB: mkTh("0.28", "0.50", "0.18", "0.23", "0.60", "0.75", "D1 留存太低，核心 loop 有问题", "D1 留存高于 50% 需检查数据污染"),
			},
			"adIntensity": {
				GLOB: mkTh("2", "4", "0", "1", "5", "6", "广告密度太低广告收入不足", "广告密度太高卸载率飙升"),
			},
			"iapMix": {
				GLOB: mkTh("0.30", "0.60", "0.10", "0.20", "0.80", "0.95", "IAP 占比太低 ARPU 天花板低", "IAP 占比太高会伤害 ad 用户"),
			},
			"arpm": {
				US:  mkTh("8", "15", "4", "6", "20", "25", "ARPM 太低广告单价偏低", "ARPM 太高说明地区精准投放过度"),
				EU:  mkTh("6", "12", "3", "5", "18", "22", "ARPM 太低广告单价偏低", "ARPM 太高说明地区精准投放过度"),
				SEA: mkTh("2", "5", "1", "1.5", "8", "10", "ARPM 太低广告单价偏低", "ARPM 太高说明地区精准投放过度"),
				CN:  mkTh("3", "7", "1", "2", "10", "15", "ARPM 太低广告单价偏低", "ARPM 太高说明地区精准投放过度"),
			},
		},
	},

	// ============ MOBA ============
	{
		ID: "moba", EName: "MOBA", CName: "MOBA/竞技",
		Notes: "MOBA 纯皮肤付费，核心玩法不能卖。排队时长和新英雄节奏是核心。",
		Metrics: []Metric{
			{"matchTime", "对局时长", "Minutes per normal match"},
			{"queueTime", "排队时长", "Seconds in queue before match"},
			{"heroRelease", "新英雄间隔", "Days between hero releases"},
			{"cosmeticARPU", "皮肤 ARPU", "ARPU from cosmetics (no gameplay pay)"},
			{"payRate", "付费率", "% players pay for skins"},
		},
		Advices: []string{
			"Pay rate ≤ 3%；核心玩法必须全免费，钱只能卖皮肤",
			"对局时长 15–25 分钟；不要超过 35 分钟",
			"新英雄间隔 21–30 天；太短平衡不上",
		},
		RegionAdvices: []RegionAdvice{
			{CN, "CN MOBA Pay rate 2–4%；青少年防沉迷，皮肤价格不能超过 168 元"},
			{US, "US MOBA Pay rate 3–5%；无防沉迷限制"},
			{SEA, "SEA MOBA Pay rate 2–4%；低 ARPU 靠量补"},
		},
		Thresholds: map[string]map[string]Threshold{
			"matchTime":   {GLOB: mkTh("15", "25", "5", "10", "30", "40", "对局太短阵容选择无意义", "对局太长退潮率 > 30%")},
			"queueTime":   {GLOB: mkTh("20", "60", "0", "10", "90", "120", "排队太短匹配池不足", "排队太长必须开跨服匹配")},
			"heroRelease": {GLOB: mkTh("21", "35", "7", "14", "60", "90", "英雄更新太慢内容枯竭", "英雄更新太快平衡不上")},
			"payRate":     {GLOB: mkTh("0.02", "0.04", "0", "0.01", "0.06", "0.10", "付费率太低皮肤吸引力不足", "付费率太高说明可能有玩法付费")},
		},
	},

	// ============ FPS ============
	{
		ID: "fps", EName: "First-Person Shooter", CName: "FPS/射击",
		Notes: "FPS 反作弊是核心外挂封禁率是 leading indicator。",
		Metrics: []Metric{
			{"matchTime", "对局时长", "Minutes per match"},
			{"queueTime", "排队时长", "Seconds before match"},
			{"cheatBanRate", "外挂封禁率", "Bans per 1000 players per day"},
			{"battlePassAdoption", "战令购买率", "% players buy battle pass"},
			{"payRate", "付费率", "% players pay"},
		},
		Advices: []string{
			"Pay rate ≤ 2%；禁止任何战力购买",
			"反作弊封禁率 > 0.5‰/天 说明反作弊正常；> 2‰ 说明漏洞严重",
			"战令购买率 ≥ 15% 说明内容吸引力足够",
		},
		RegionAdvices: []RegionAdvice{
			{CN, "CN FPS Pay rate 1.5–3%；战令必须全价格透明"},
			{US, "US FPS Pay rate 2–4%"},
			{SEA, "SEA FPS Pay rate 1–3%；ARPU 低靠量补"},
		},
		Thresholds: map[string]map[string]Threshold{
			"cheatBanRate":       {GLOB: mkTh("0.0005", "0.002", "0", "0.0001", "0.005", "0.020", "反作弊封禁率太低说明漏洞严重", "反作弊封禁率太高可能误封玩家")},
			"battlePassAdoption": {GLOB: mkTh("0.15", "0.30", "0", "0.10", "0.40", "0.60", "战令购买率太低吸引力不足", "战令购买率太高核心玩法过于依赖战令")},
			"queueTime":          {GLOB: mkTh("30", "90", "0", "10", "120", "180", "排队太短匹配池不足", "排队太长必须开跨服")},
		},
	},

	// ============ MMO ============
	{
		ID: "mmo", EName: "Massively Multiplayer Online", CName: "MMO/大型多人",
		Notes: "MMO 服务器人口 2000–8000 稳定。公会活跃密度是 retention 的 leading indicator。",
		Metrics: []Metric{
			{"guildActivity", "公会活跃密度", "% members online per day"},
			{"serverPop", "服务器人口", "Peak concurrent users per server"},
			{"auctionVolume", "拍卖行成交额", "Daily base-unit GMV"},
			{"lvlCapDays", "满级天数", "Days for normal player to hit cap"},
			{"raidClears", "Raid 通关率", "% raids cleared in first attempt"},
		},
		Advices: []string{
			"服务器人口 2000–8000 是稳定区间；低于 1000 开合并流程",
			"公会活跃密度 ≥ 30%；低于 20% 会长会流失",
			"满级天数 30–90 天；用 progression.Run 精确算",
		},
		RegionAdvices: []RegionAdvice{
			{CN, "CN MMO Pay rate 2–5%；满级天数 45–90 天"},
			{US, "US MMO Pay rate 3–6%；满级天数 60–120 天"},
		},
		Thresholds: map[string]map[string]Threshold{
			"guildActivity": {GLOB: mkTh("0.30", "0.70", "0", "0.20", "0.90", "1.0", "公会活跃密度太低会长流失", "公会活跃密度过高玩家疲于奔命")},
			"serverPop":     {GLOB: mkTh("2000", "8000", "500", "1000", "12000", "20000", "服务器人口太少开合并", "服务器人口太多匹配延迟上升")},
			"lvlCapDays": {
				CN:   mkTh("45", "90", "15", "30", "120", "180", "满级太快内容消耗过快", "满级太慢对轻玩家是 barrier"),
				GLOB: mkTh("45", "120", "15", "30", "150", "210", "满级太快内容消耗过快", "满级太慢对轻玩家是 barrier"),
			},
			"raidClears": {GLOB: mkTh("0.05", "0.20", "0", "0.02", "0.40", "0.60", "Raid 通关率太低数值或组织有问题", "Raid 通关率太高太浅无挑战")},
		},
	},

	// ============ MMORPG ============
	{
		ID: "mmorpg", EName: "MMORPG", CName: "MMORPG",
		Notes: "MMORPG 是 MMO + RPG 成长。周 raid 次数 2–4 次；世界 BOSS 参与率 ≥ 25%。",
		Metrics: []Metric{
			{"raidWeekly", "周 raid 次数", "Raids per guild per week"},
			{"worldBoss", "世界 BOSS 参与率", "% players that join weekly"},
			{"raidClearRate", "Raid 通关率", "% raids cleared"},
			{"lvlCapDays", "满级天数", "Days for normal player"},
		},
		Advices: []string{
			"Raid 周频率 2–4 次；低于 2 次公会内容不足，高于 4 次变成工作",
			"世界 BOSS 参与率 ≥ 25% 说明活动组织正常",
			"满级天数 60–120 天；偏长对轻玩家是 barrier",
		},
		Thresholds: map[string]map[string]Threshold{
			"raidWeekly": {GLOB: mkTh("2", "4", "0", "1", "5", "7", "Raid 太少公会内容不足", "Raid 太多玩家变成打工")},
			"worldBoss":  {GLOB: mkTh("0.25", "0.60", "0", "0.15", "0.80", "1.0", "世界 BOSS 参与率太低组织有问题", "世界 BOSS 参与率太高碾压设计失败")},
			"lvlCapDays": {GLOB: mkTh("60", "120", "15", "30", "150", "210", "满级太快内容消耗过快", "满级太慢对轻玩家是 barrier")},
		},
	},

	// ============ MATCH3 ============
	{
		ID: "match3", EName: "Match-3", CName: "三消",
		Notes: "三消的关卡难度曲线是核心。Hard stage 通关率 ≥ 5%。",
		Metrics: []Metric{
			{"staminaLimit", "体力上限", "Max stamina per day"},
			{"clearRate", "通关率", "% players clear hard stages"},
			{"eventCadence", "活动间隔", "Days between themed events"},
		},
		Advices: []string{
			"Hard stage 通关率 ≥ 5%；低于 2% 关卡数值爆了",
			"体力上限 60–120；低于 40 玩家卡关会退，高于 240 弱化广告驱动",
			"活动间隔 ≤ 14 天；超过 21 天内容耗尽",
		},
		Thresholds: map[string]map[string]Threshold{
			"staminaLimit": {GLOB: mkTh("60", "120", "0", "40", "200", "300", "体力上限太小玩家卡关会退", "体力上限太大弱化广告驱动")},
			"clearRate":    {GLOB: mkTh("0.05", "0.15", "0", "0.02", "0.25", "0.40", "Hard stage 通关率太低数值爆了", "Hard stage 通关率太高关卡太浅")},
			"eventCadence": {GLOB: mkTh("7", "14", "0", "5", "21", "35", "活动间隔太短开发压力大", "活动间隔太长内容耗尽")},
		},
	},

	// ============ Idle ============
	{
		ID: "idle", EName: "Idle / Incremental", CName: "放置/增量",
		Notes: "放置的核心是离线倍率 ≤ 2× 和重置周期 30–90 天。",
		Metrics: []Metric{
			{"offlineMult", "离线倍率", "Production multiplier for offline time"},
			{"resetDays", "重置周期", "Days between prestige/ascension"},
			{"passiveShare", "被动占比", "% total XP from offline"},
			{"retention_d30", "D30 留存", "% day-30 retention"},
		},
		Advices: []string{
			"离线倍率 ≤ 2×；高于 2× 玩家只挂机不玩，广告收入崩",
			"重置周期 30–90 天；太短重置疲劳，太长内容撑不住",
		},
		RegionAdvices: []RegionAdvice{
			{CN, "CN Idle D30 ≥ 6%；SEA ≥ 5%；US ≥ 4%"},
		},
		Thresholds: map[string]map[string]Threshold{
			"offlineMult":  {GLOB: mkTh("1.5", "2.0", "0", "1.2", "2.5", "3.0", "离线倍率太低玩家挂机无动力", "离线倍率太高只挂机不玩广告收入崩")},
			"resetDays":    {GLOB: mkTh("30", "90", "10", "15", "150", "200", "重置周期太短玩家疲劳", "重置周期太长内容撑不住")},
			"passiveShare": {GLOB: mkTh("0.50", "0.70", "0", "0.35", "0.80", "0.90", "被动占比太低失去放置意义", "被动占比太高不再是放置玩法")},
			"retention_d30": {
				CN:  mkTh("0.06", "0.12", "0.02", "0.04", "0.15", "0.20", "D30 留存低于 CN 基准", "D30 留存异常偏高需检查数据污染"),
				SEA: mkTh("0.05", "0.10", "0.02", "0.03", "0.13", "0.18", "D30 留存低于 SEA 基准", "D30 留存异常偏高需检查数据污染"),
				US:  mkTh("0.04", "0.08", "0.01", "0.02", "0.10", "0.15", "D30 留存低于 US 基准", "D30 留存异常偏高需检查数据污染"),
			},
		},
	},

	// ============ Battle Royale ============
	{
		ID: "battle-royale", EName: "Battle Royale", CName: "大逃杀",
		Notes: "大逃杀是纯皮肤 + 战令。房间人数 60–100；对局 15–20 分钟。",
		Metrics: []Metric{
			{"matchTime", "对局时长", "Minutes"},
			{"dropSize", "房间人数", "Players per match"},
			{"battlePassAdoption", "战令购买率", "%"},
		},
		Advices: []string{
			"房间人数 60–100；过少地图空荡，过多服务器扛不住",
			"对局时长 15–20 分钟；超过 25 分钟退潮率 > 30%",
			"战令购买率 ≥ 20% 健康",
		},
		Thresholds: map[string]map[string]Threshold{
			"matchTime": {GLOB: mkTh("15", "20", "5", "10", "25", "35", "对局太短没有生存决策时间", "对局太长退潮率飙升")},
			"dropSize":  {GLOB: mkTh("60", "100", "30", "45", "150", "200", "房间人数太少地图空荡", "房间人数太多服务器扛不住")},
		},
	},

	// ============ Card ============
	{
		ID: "card", EName: "Collectible Card Game", CName: "卡牌",
		Notes: "卡牌 gacha 池周期 90–180 天；万能率 ≥ 5% 防重复疲劳。",
		Metrics: []Metric{
			{"setCycleDays", "卡包周期", "Days between new set rotation"},
			{"wildcardRate", "万能率", "% cards that are wildcards"},
			{"payRate", "付费率", "% players that buy packs"},
		},
		Advices: []string{
			"卡包周期 90–180 天；太短环境动荡，太长内容枯竭",
			"万能率 ≥ 5%；无万能卡重复抽会让玩家失去动力",
		},
		Thresholds: map[string]map[string]Threshold{
			"setCycleDays": {GLOB: mkTh("90", "180", "30", "60", "270", "365", "卡包周期太短环境动荡", "卡包周期太长内容枯竭")},
			"wildcardRate": {GLOB: mkTh("0.05", "0.15", "0", "0.02", "0.25", "0.35", "万能率太低重复抽会让玩家失去动力", "万能率太高卡包失去收藏意义")},
			"payRate":      {GLOB: mkTh("0.02", "0.05", "0", "0.01", "0.08", "0.10", "付费率太低卡包吸引力不足", "付费率太高玩家流失")},
		},
	},

	// ============ FPS / TPS / Battle Royale 其余品类 ============
	// 这些品类的 threshold 表类似——简略处理。
	{
		ID: "tps", EName: "Third-Person Shooter", CName: "TPS/第三人称射击",
		Notes: "TPS 与 FPS 共用战令和皮肤池。掩体覆盖率 30–40%。",
		Metrics: []Metric{
			{"matchTime", "对局时长", "Minutes per match"},
			{"queueTime", "排队时长", "Seconds"},
			{"coverRate", "掩体覆盖率", "% map area under cover"},
		},
		Advices: []string{
			"和 FPS 共用战令和皮肤池；不要重复制作",
			"掩体覆盖率 30–40% 最优",
		},
		RegionAdvices: []RegionAdvice{
			{CN, "CN TPS Pay rate ≤ 2.5%"},
			{US, "US TPS Pay rate ≤ 3%"},
			{SEA, "SEA TPS Pay rate ≤ 2%"},
		},
		Thresholds: map[string]map[string]Threshold{
			"coverRate": {GLOB: mkTh("0.30", "0.40", "0", "0.20", "0.55", "0.70", "掩体覆盖率太低战术深度不足", "掩体覆盖率太高玩家只会蹲点")},
		},
	},

	// ============ Idle done above ============

	// ============ Survival ============
	{
		ID: "survival", EName: "Survival", CName: "生存",
		Notes: "生存的重置周期 30–90 天；死亡惩罚 10–30%。",
		Metrics: []Metric{
			{"wipeCycle", "重置周期", "Days between world wipes"},
			{"deathPunish", "死亡惩罚", "% lost progress on death"},
			{"raidFrequency", "PvP raid 频率", "% players participate per week"},
		},
		Advices: []string{
			"重置周期 30–90 天；太短玩家没积累感，太长新玩家进不去",
			"死亡惩罚 10–30%；0% 没压力，≥ 50% 卸载",
			"PvP raid 参与率 ≥ 40%",
		},
		Thresholds: map[string]map[string]Threshold{
			"wipeCycle":     {GLOB: mkTh("30", "90", "10", "15", "150", "200", "重置周期太短玩家没积累感", "重置周期太长新玩家进不去")},
			"deathPunish":   {GLOB: mkTh("0.10", "0.30", "0", "0.05", "0.40", "0.60", "死亡惩罚太低失去紧张感", "死亡惩罚太高玩家卸载")},
			"raidFrequency": {GLOB: mkTh("0.40", "0.70", "0", "0.25", "0.85", "1.0", "PvP raid 参与率太低对抗不足", "PvP raid 参与率太高 PvE 玩家流失")},
		},
	},

	// ============ Simulation ============
	{
		ID: "simulation", EName: "Simulation", CName: "模拟经营",
		Notes: "模拟经营货币种类 ≤ 5；通胀漂移 ≤ 10%/季度。",
		Metrics: []Metric{
			{"currencyTypes", "货币种类", "Total distinct resource currencies"},
			{"inflationDrift", "通胀漂移", "Price index 90-day change"},
			{"contentCadence", "内容更新间隔", "Days"},
		},
		Advices: []string{
			"货币种类 ≤ 5 种；超过 6 种玩家记不住",
			"通胀漂移 ≤ 10% / 季度；高于 15% 必须强制 rebalance",
		},
		Thresholds: map[string]map[string]Threshold{
			"currencyTypes":  {GLOB: mkTh("3", "5", "0", "1", "6", "8", "货币种类太少经济模型简单", "货币种类太多玩家记不住")},
			"inflationDrift": {GLOB: mkTh("-0.05", "0.10", "-0.15", "-0.10", "0.15", "0.25", "通胀漂移太低经济死水一潭", "通胀漂移太高必须强制 rebalance")},
			"contentCadence": {GLOB: mkTh("45", "90", "15", "30", "120", "180", "内容更新太短开发压力大", "内容更新太长玩家内容消耗完")},
		},
	},

	// ============ Sandbox ============
	{
		ID: "sandbox", EName: "Sandbox", CName: "沙盒",
		Notes: "沙盒 UGC 占比 ≥ 50%；多人参与率 ≥ 30%。",
		Metrics: []Metric{
			{"ugcShare", "UGC 占比", "% time players spend on user content"},
			{"multiplayer", "多人参与率", "% players join multiplayer sessions"},
		},
		Advices: []string{
			"UGC 占比 ≥ 50% 说明玩家自发活跃",
			"多人参与率 ≥ 30%；低于 20% 说明单人体验太重",
		},
		Thresholds: map[string]map[string]Threshold{
			"ugcShare":    {GLOB: mkTh("0.50", "0.80", "0", "0.30", "0.90", "1.0", "UGC 占比太低官方内容不够", "UGC 占比太高需要审核压力大")},
			"multiplayer": {GLOB: mkTh("0.30", "0.60", "0", "0.20", "0.80", "1.0", "多人参与率太低单人体验太重", "多人参与率太高新玩家入门门槛高")},
		},
	},

	// ============ Roguelike ============
	{
		ID: "roguelike", EName: "Roguelike / Roguelite", CName: "肉鸽",
		Notes: "肉鸽构建种类 ≥ 15；通关率 ≈ 5–10%；元进度周期 30–60 天。",
		Metrics: []Metric{
			{"buildCount", "构建数量", "Distinct meta build archetypes"},
			{"clearRate", "通关率", "% runs cleared (blind) vs expected"},
			{"metaProgDays", "元进度周期", "Days to unlock every meta perk"},
		},
		Advices: []string{
			"构建种类 ≥ 15；低于 8 深度不足",
			"元进度周期 30–60 天；太短 meta 疲劳，太长新玩家卡关",
			"通关率 ≈ 5–10%；高于 15% 太浅，低于 2% 太难",
		},
		Thresholds: map[string]map[string]Threshold{
			"buildCount":   {GLOB: mkTh("15", "40", "0", "8", "60", "100", "构建种类太少深度不足", "构建种类太多无法平衡")},
			"clearRate":    {GLOB: mkTh("0.05", "0.10", "0", "0.02", "0.20", "0.30", "通关率太低难度曲线过高", "通关率太高游戏太浅")},
			"metaProgDays": {GLOB: mkTh("30", "60", "0", "15", "90", "150", "元进度太短 meta 疲劳", "元进度太长新玩家卡关")},
		},
	},

	// ============ Fighting ============
	{
		ID: "fighting", EName: "Fighting", CName: "格斗",
		Notes: "格斗 Top-3 tier 覆盖率 ≥ 50%；平衡补丁 30–60 天。",
		Metrics: []Metric{
			{"patchCadence", "平衡补丁", "Days"},
			{"tierSpread", "平衡度", "% characters in top 3 tiers"},
		},
		Advices: []string{
			"平衡补丁 30–60 天；超过 90 天 scene 萎缩",
			"Top-3 tier 覆盖率 ≥ 50% 角色；低于 30% 太偏",
		},
		Thresholds: map[string]map[string]Threshold{
			"patchCadence": {GLOB: mkTh("30", "60", "0", "15", "90", "120", "平衡补丁太短玩家没时间适应", "平衡补丁太长 scene 萎缩")},
			"tierSpread":   {GLOB: mkTh("0.50", "0.70", "0", "0.30", "0.85", "1.0", "Top-3 tier 覆盖率太低平衡偏", "Top-3 tier 覆盖率太高角色趋同")},
		},
	},

	// ============ Platformer ============
	{
		ID: "platformer", EName: "Platformer", CName: "平台跳跃",
		Notes: "平台跳跃主线 8–12 小时；全通率 ≥ 10%。",
		Metrics: []Metric{
			{"runLength", "主线时长", "Hours to 100%"},
			{"%100", "全通率", "% players hit 100%"},
		},
		Advices: []string{"主线时长 8–12 小时", "全通率 ≥ 10%"},
		Thresholds: map[string]map[string]Threshold{
			"runLength": {GLOB: mkTh("8", "12", "2", "5", "18", "25", "主线太短内容消耗快", "主线太长玩家疲劳")},
			"%100":      {GLOB: mkTh("0.10", "0.25", "0", "0.05", "0.40", "0.60", "全通率太低难度曲线过高", "全通率太高游戏太浅")},
		},
	},

	// ============ Sports ============
	{
		ID: "sports", EName: "Sports", CName: "体育",
		Notes:   "体育赛季严格同步真实世界 180–365 天。",
		Metrics: []Metric{{"seasonDays", "赛季时长", "Days per in-game season"}},
		Advices: []string{"赛季时长严格同步真实世界（180–365 天）"},
		Thresholds: map[string]map[string]Threshold{
			"seasonDays": {GLOB: mkTh("180", "365", "90", "120", "420", "540", "赛季太短同步不上真实世界", "赛季太长超过真实世界逻辑")},
		},
	},

	// ============ Racing ============
	{
		ID: "racing", EName: "Racing", CName: "赛车",
		Notes:   "赛车车款胜率方差 ≤ ±3%。",
		Metrics: []Metric{{"carBalance", "车款平衡", "Win-rate variance across cars"}},
		Advices: []string{"车款胜率方差 ≤ ±3% 健康"},
		Thresholds: map[string]map[string]Threshold{
			"carBalance": {GLOB: mkTh("0.01", "0.03", "0", "0.005", "0.05", "0.10", "车款平衡方差太大需补丁", "车款平衡方差太小需增加差异化")},
		},
	},

	// ============ Rhythm ============
	{
		ID: "rhythm", EName: "Rhythm", CName: "音乐节奏",
		Notes:   "音乐谱面间隔 14–30 天；Easy/Normal/Hard 30/40/30 最优。",
		Metrics: []Metric{{"chartCadence", "谱面间隔", "Days between new charts"}},
		Advices: []string{
			"谱面间隔 14–30 天；超过 45 天内容枯竭",
			"Easy/Normal/Hard 覆盖率 30/40/30 最优",
		},
		Thresholds: map[string]map[string]Threshold{
			"chartCadence": {GLOB: mkTh("14", "30", "0", "7", "45", "60", "谱面间隔太短开发压力大", "谱面间隔太长内容枯竭")},
		},
	},

	// ============ Tower Defense ============
	{
		ID: "tower-defense", EName: "Tower Defense", CName: "塔防",
		Notes: "塔防每关 15–30 波；通关率 5–20%；塔胜率方差 ≤ ±3%。",
		Metrics: []Metric{
			{"waveCount", "波次", "Waves per level"},
			{"clearRate", "通关率", "% clear all waves"},
			{"towerBalance", "塔种类平衡", "Win-rate variance"},
		},
		Advices: []string{"每关 15–30 波", "通关率 5–20%"},
		Thresholds: map[string]map[string]Threshold{
			"waveCount":    {GLOB: mkTh("15", "30", "5", "8", "45", "60", "波次太少深度不够", "波次太长疲劳")},
			"clearRate":    {GLOB: mkTh("0.05", "0.20", "0", "0.02", "0.35", "0.50", "通关率太低关卡太难", "通关率太高关卡太浅")},
			"towerBalance": {GLOB: mkTh("0.01", "0.03", "0", "0.005", "0.05", "0.10", "塔胜率方差太大需平衡", "塔胜率方差太小需差异化")},
		},
	},

	// ============ Visual Novel ============
	{
		ID: "visual-novel", EName: "Visual Novel", CName: "视觉小说",
		Notes: "视觉小说章节间隔 90–180 天；Premium 价格 3000–6000 base units。",
		Metrics: []Metric{
			{"chapterCadence", "章节间隔", "Days"},
			{"premiumPrice", "Premium 价格", "Base-unit cost of premium chapters"},
		},
		Advices: []string{"章节间隔 90–180 天；太长粉丝流失"},
		Thresholds: map[string]map[string]Threshold{
			"chapterCadence": {GLOB: mkTh("90", "180", "0", "60", "270", "365", "章节间隔太短开发压力大", "章节间隔太长粉丝流失")},
		},
	},

	// ============ Adventure ============
	{
		ID: "adventure", EName: "Adventure", CName: "冒险",
		Notes:   "冒险主线 10–20 小时；全通率 ≥ 15%。",
		Metrics: []Metric{{"playtimeHours", "主线时长", "Hours"}, {"%Complete", "全通率", "% players hit 100%"}},
		Advices: []string{"主线时长 10–20 小时", "全通率 ≥ 15%"},
		Thresholds: map[string]map[string]Threshold{
			"playtimeHours": {GLOB: mkTh("10", "20", "3", "6", "30", "40", "主线太短内容消耗快", "主线太长玩家疲劳")},
			"%Complete":     {GLOB: mkTh("0.15", "0.30", "0", "0.08", "0.45", "0.60", "全通率太低难度曲线过高", "全通率太高游戏太浅")},
		},
	},

	// ============ Horror ============
	{
		ID: "horror", EName: "Horror", CName: "恐怖",
		Notes:   "恐怖主线 6–15 小时；Jump scare 密度 ≤ 3/hour。",
		Metrics: []Metric{{"playtimeHours", "主线时长", "Hours"}},
		Advices: []string{"主线时长 6–15 小时", "Jump scare 密度 ≤ 3/hour"},
		Thresholds: map[string]map[string]Threshold{
			"playtimeHours": {GLOB: mkTh("6", "15", "2", "3", "25", "35", "主线太短内容消耗快", "主线太长恐怖效果递减")},
		},
	},

	// ============ RPG ============
	{
		ID: "rpg", EName: "Role-Playing", CName: "RPG",
		Notes: "RPG 满级 90–180 天；职业胜率方差 ≤ ±3%。",
		Metrics: []Metric{
			{"lvlCapDays", "满级天数", "Days for normal player"},
			{"classBalance", "职业平衡", "Win-rate variance"},
			{"sideQuestCount", "支线数量", "Side quests"},
		},
		Advices: []string{"满级天数 90–180 天", "职业胜率方差 ≤ ±3%"},
		Thresholds: map[string]map[string]Threshold{
			"lvlCapDays":     {GLOB: mkTh("90", "180", "15", "30", "240", "365", "满级太快内容消耗过快", "满级太慢对轻玩家是 barrier")},
			"classBalance":   {GLOB: mkTh("0.01", "0.03", "0", "0.005", "0.05", "0.10", "职业胜率方差太大需平衡", "职业胜率方差太小需差异化")},
			"sideQuestCount": {GLOB: mkTh("20", "60", "5", "10", "100", "150", "支线太少内容不足", "支线太多主线被淹没")},
		},
	},

	// ============ Action RPG ============
	{
		ID: "action-rpg", EName: "Action RPG", CName: "动作 RPG",
		Notes: "动作 RPG 连段上限 ≤ 20 hits 防 input reading；新内容间隔 21–45 天。",
		Metrics: []Metric{
			{"lvlCapDays", "满级天数", "Days"},
			{"comboCeiling", "连段上限", "Max hits per combo"},
		},
		Advices: []string{"连段上限 ≤ 20 hits 防作弊", "新内容间隔 21–45 天"},
		Thresholds: map[string]map[string]Threshold{
			"comboCeiling": {GLOB: mkTh("10", "20", "3", "5", "30", "50", "连段上限太少连招不自由", "连段上限太多 input reading 作弊可达成")},
		},
	},

	// ============ Strategy (RTS/4X) ============
	{
		ID: "strategy", EName: "Strategy", CName: "策略",
		Notes: "策略 RTS/4X 单位胜率方差 ≤ ±2%；平衡补丁 30–60 天。",
		Metrics: []Metric{
			{"patchCadence", "平衡补丁", "Days"},
			{"unitVariance", "单位平衡", "Win-rate variance"},
		},
		Advices: []string{"单位胜率方差 ≤ ±2%", "平衡补丁 30–60 天"},
		Thresholds: map[string]map[string]Threshold{
			"patchCadence": {GLOB: mkTh("30", "60", "0", "15", "90", "120", "平衡补丁太短玩家没适应", "平衡补丁太长 scene 萎缩")},
			"unitVariance": {GLOB: mkTh("0.01", "0.02", "0", "0.005", "0.04", "0.08", "单位胜率方差太大需平衡", "单位胜率方差太小需差异化")},
		},
	},

	// ============ Puzzle ============
	{
		ID: "puzzle", EName: "Puzzle", CName: "解谜",
		Notes: "解谜关卡 ≥ 300；活动间隔 14–30 天。",
		Metrics: []Metric{
			{"puzzleCount", "关卡数", "Total levels"},
			{"eventCadence", "活动间隔", "Days"},
		},
		Advices: []string{"总关卡 ≥ 300", "活动间隔 14–30 天"},
		Thresholds: map[string]map[string]Threshold{
			"puzzleCount":  {GLOB: mkTh("300", "800", "50", "100", "1500", "3000", "关卡太少内容消耗快", "关卡太多开发周期无法覆盖")},
			"eventCadence": {GLOB: mkTh("14", "30", "0", "7", "45", "60", "活动间隔太短开发压力大", "活动间隔太长内容枯竭")},
		},
	},
}

// ---- Generator ----

func renderStrategy(s StrategySpec, pkgName string) string {
	var b stringsBuilder

	// Header + imports.
	fmt.Fprintf(&b, `// Package %s — strategy for %s (%s).
// Auto-generated by gen-strategies.go. Edit the spec, rerun.
package %s

import (
	"fmt"
	"strconv"

	"github.com/neko233-com/game-ops/factory"
	"github.com/neko233-com/game-ops/num"
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
	fmt.Fprintf(&b, "\t}\n}\n\n")

	// Advice() — merge default + region-filtered.
	fmt.Fprintf(&b, "func (strategy) Advice() []string {\n\treturn []string{\n")
	for _, a := range s.Advices {
		fmt.Fprintf(&b, "\t\t%q,\n", a)
	}
	fmt.Fprintf(&b, "\t}\n}\n\n")

	// Thresholds table.
	if len(s.Thresholds) > 0 {
		fmt.Fprintf(&b, "var thresholds = map[string]map[string]factory.ThresholdTableKey{\n")
		keys := make([]string, 0, len(s.Thresholds))
		for k := range s.Thresholds {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, key := range keys {
			fmt.Fprintf(&b, "\t%q: {\n", key)
			regionMap := s.Thresholds[key]
			regs := make([]string, 0, len(regionMap))
			for r := range regionMap {
				regs = append(regs, r)
			}
			sort.Strings(regs)
			for _, r := range regs {
				t := regionMap[r]
				fmt.Fprintf(&b, "\t\t%q: {MinLow: %q, Min: %q, IdealLo: %q, IdealHi: %q, Max: %q, MaxHigh: %q, MsgLow: %q, MsgHigh: %q, Unit: %q},\n",
					r, t.MinLow, t.Min, t.IdealLo, t.IdealHi, t.Max, t.MaxHigh, t.MsgLow, t.MsgHigh, t.Unit)
			}
			fmt.Fprintf(&b, "\t},\n")
		}
		fmt.Fprintf(&b, "}\n\n")
	}

	// RegionAdvices table.
	if len(s.RegionAdvices) > 0 {
		fmt.Fprintf(&b, "var regionAdvices = []factory.RegionAdvice{\n")
		for _, ra := range s.RegionAdvices {
			fmt.Fprintf(&b, "\t{Region: %q, Advice: %q},\n", ra.Region, ra.Advice)
		}
		fmt.Fprintf(&b, "}\n\n")
	}

	// Evaluate — real implementation.
	fmt.Fprintf(&b, `func (s strategy) Evaluate(ctx factory.StrategyContext) (factory.StrategyResult, error) {
	res := factory.StrategyResult{
		Genre:  %q,
		Region: ctx.Region,
		Notes: %s,
	}

	// 1. Emit GenreMetrics for observed values.
	for _, ml := range s.GenreMetrics() {
		v, ok := ctx.Metrics[ml.Key]
		if !ok {
			continue
		}
		res.GenreMetrics = append(res.GenreMetrics, factory.StrategyMetric{
			Key:   ml.Key,
			Label: ml.Label,
			Value: v,
		})
	}

	// 2. Threshold checks — region-aware.
`, s.ID, "`"+s.Notes+"`")

	if len(s.Thresholds) > 0 {
		fmt.Fprintf(&b, `	for metricKey, regionMap := range thresholds {
		v, ok := ctx.Metrics[metricKey]
		if !ok {
			continue
		}
		t, ok := regionMap[string(ctx.Region)]
		if !ok {
			t, ok = regionMap["GLOBAL"]
			if !ok {
				continue
			}
		}
		if n, err := num.FromString(t.Min); err == nil && v.Cmp(n) < 0 {
			sev := "warn"
			if nLo, err := num.FromString(t.MinLow); err == nil && v.Cmp(nLo) < 0 {
				sev = "critical"
			}
			res.Risks = append(res.Risks, factory.RiskItem{
				Metric: metricKey, Observed: v, Ideal: n, Severity: sev,
				Message: fmt.Sprintf("%%s (%%s): observed %%s", t.MsgLow, metricKey, v.DecimalString(6)),
			})
		}
		if x, err := num.FromString(t.Max); err == nil && v.Cmp(x) > 0 {
			sev := "warn"
			if xHi, err := num.FromString(t.MaxHigh); err == nil && v.Cmp(xHi) > 0 {
				sev = "critical"
			}
			res.Risks = append(res.Risks, factory.RiskItem{
				Metric: metricKey, Observed: v, Ideal: x, Severity: sev,
				Message: fmt.Sprintf("%%s (%%s): observed %%s", t.MsgHigh, metricKey, v.DecimalString(6)),
			})
		}
	}

`)
	}

	// 3. Region-aware advice append.
	if len(s.RegionAdvices) > 0 {
		fmt.Fprintf(&b, `	region := string(ctx.Region)
	for _, ra := range regionAdvices {
		if ra.Region == "GLOBAL" || ra.Region == region {
			res.Advice = append(res.Advice, factory.AdviceItem{Reason: ra.Advice})
		}
	}
`)
	}

	fmt.Fprintf(&b, `	return res, nil
}

// Silence unused strconv warning when the file happens to not use it.
var _ = strconv.Itoa
`)

	_ = strings.Count("", "")
	return b.String()
}

// ---- Utility ----

type stringsBuilder struct{ buf []byte }

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
	sort.Slice(spec, func(i, j int) bool { return spec[i].ID < spec[j].ID })

	for _, s := range spec {
		pkg := strings.ReplaceAll(s.ID, "-", "_")
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
