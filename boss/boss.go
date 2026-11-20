// Package boss turns exact game-ops analysis into decisions a non-specialist
// can act on. It is the "boss layer": conclusions first, then the choices that
// need a human sign-off, then executable instructions — no formulas in the
// headline, no unexplained metric keys.
//
// The layer does NOT replace the engine; it translates it. Every number shown
// here still comes from num.Num exact arithmetic in ops/genre/factory.
package boss

import (
	"fmt"
	"sort"
	"time"

	"github.com/neko233-com/game-ops/genre"
	"github.com/neko233-com/game-ops/num"
	"github.com/neko233-com/game-ops/ops"
)

// DutyItem is one concrete job in the lead-designer / operations workflow,
// bound to an engine capability and a measurable standard.
type DutyItem struct {
	ID       string `json:"id"`
	Role     string `json:"role"`     // 主策划 | 运营
	Area     string `json:"area"`     // 领域
	Duty     string `json:"duty"`     // 事务
	Standard string `json:"standard"` // 判定标准
	Tool     string `json:"tool"`     // 引擎命令/API
	Metric   string `json:"metric"`   // 关联指标（可空）
	Status   string `json:"status"`   // ok | warn | critical | idle
	Finding  string `json:"finding"`  // 当前结论
}

// DutyMatrix is the full workflow coverage for both roles.
var DutyMatrix = []DutyItem{
	// ---- 主策划 ----
	{"D01", "主策划", "价值基准", "定义基准单位并把所有资源归一化", "所有资源可换算到 1 个基准单位，兑换率精确", "cmd/normalize · normalize.Ledger", "", "idle", "待定义基准单位"},
	{"D02", "主策划", "战斗数值", "设计攻击/生命/战力曲线", "加法曲线，单调、有 cap，无全局指数", "cmd/numgen · design.AdditiveCurve", "", "idle", "待设计曲线"},
	{"D03", "主策划", "成长节奏", "设计升级经验曲线", "相邻等级增长比受控；分段加法而非指数", "cmd/curve · design.AdditivePiecewise", "progression_pace", "idle", "待设计"},
	{"D04", "主策划", "商业化定价", "定价付费点与单抽价值", "单抽价值 ÷ 同类产出 ≤ 1.5×", "cmd/numgen · design.Plan.Price", "monetization_pay_rate", "idle", "待定价"},
	{"D05", "主策划", "抽卡概率", "保底/掉率/期望值设计", "保底抽数与掉率精确；CN 需公示并 80 保底", "cmd/numcalc geom/binom", "", "idle", "待设计"},
	{"D06", "主策划", "经济系统", "平衡货币产出与消耗", "source/sink ≈ 1，通胀指数 ≈ 1", "cmd/econcheck", "economy_source_sink_ratio", "idle", "待平衡"},
	{"D07", "主策划", "内容节奏", "规划版本内容投放间隔", "按品类基准（如 MMO 7 天、休闲 14 天）", "cmd/catalog eval", "content_cadence_days", "idle", "待规划"},
	{"D08", "主策划", "版本评审", "评审当前版本数值健康", "健康度 ≥ 0.8 才能进版本", "cmd/opsreport · ops.Evaluate", "", "idle", "待评审"},
	{"D09", "主策划", "配表规范", "校验配置表数值分布", "无 >10× 异常跨度，无负值/空值", "cmd/ingest", "", "idle", "待扫描"},
	{"D10", "主策划", "满级时间", "测算常规玩家满级天数", "P50 玩家可在留存窗口内触达 cap", "cmd/proanalysis", "progression_pace", "idle", "待测算"},
	{"D11", "主策划", "品类基准", "为新品类建立 lever 基准", "11 项 lever 完整，权重 >0", "cmd/catalog derive/register", "", "idle", "待建立"},
	{"D12", "主策划", "指数守卫", "拒绝全局指数公式", "任何 base*1.2^n 即使有 cap 也拒绝", "cmd/numgen -formula", "", "idle", "待检查"},

	// ---- 运营 ----
	{"O01", "运营", "次日留存", "监控 D1 并定位核心循环问题", "D1 ≥ 品类理想，低于下限立即介入", "cmd/catalog eval", "retention_d1", "idle", "待观测"},
	{"O02", "运营", "长期留存", "监控 D7/D30 衰减", "D7/D30 在容差带内", "cmd/catalog eval", "retention_d7", "idle", "待观测"},
	{"O03", "运营", "付费转化", "监控付费率与付费点吸引", "付费率在品类区间，异常高排查玩法付费", "cmd/catalog eval", "monetization_pay_rate", "idle", "待观测"},
	{"O04", "运营", "付费深度", "监控 ARPU 与 LTV", "ARPU/LTV 达品类基准", "cmd/catalog eval", "monetization_ltv", "idle", "待观测"},
	{"O05", "运营", "通胀监控", "监控物价与货币存量", "通胀指数 ≈ 1，偏离 >10% 强制 rebalance", "cmd/econcheck", "economy_inflation", "idle", "待观测"},
	{"O06", "运营", "活动排期", "安排活动与内容投放", "间隔不超品类上限，提前 7 天预告", "cmd/catalog eval", "content_cadence_days", "idle", "待排期"},
	{"O07", "运营", "内容负债", "跟踪内容积压比", "积压比 ≈ 1，>2 需扩产", "cmd/catalog eval", "content_debt_ratio", "idle", "待观测"},
	{"O08", "运营", "成长节奏", "监控玩家实际推进速度", "实际/设计 ≈ 1", "cmd/proanalysis", "progression_pace", "idle", "待观测"},
	{"O09", "运营", "数据盘点", "定期扫描配表与观察分布", "扫描全部配置，异常列重点检查", "cmd/ingest", "", "idle", "待扫描"},
	{"O10", "运营", "风险预警", "排序最严重偏差并预警", "按加权偏差 Top3 生成决策包", "cmd/judge", "", "idle", "待生成"},
	{"O11", "运营", "玩家分层", "按常规/轻度/重度看满级触达", "P50 加权，报告满级时留存", "cmd/proanalysis", "", "idle", "待分层"},
	{"O12", "运营", "合规检查", "按地区核对法律要求", "CN：抽卡公示/保底/防沉迷；海外按地区", "cmd/catalog strategy -region", "", "idle", "待核对"},
	{"O13", "运营", "版本节奏", "比对内容消耗与产出速率", "消耗 ≤ 产出，避免内容枯竭", "cmd/catalog eval", "content_debt_ratio", "idle", "待分析"},
	{"O14", "运营", "反馈闭环", "收集社区反馈并回灌数值", "数值无法衡量趣味/社交，需人工复核", "人工 + cmd/judge", "", "idle", "需人工"},
}

// Translate turns a raw metric key into boss language plus the value/flow/
// pacing root-cause family. This is the core "主策/运营 → 老板" translation.
type Translation struct {
	Key      string
	Plain    string // 经营语言
	Question string // 三问归属：价值 / 流动 / 节奏
	Lever    string // 可操作的杠杆
	Direction string // 高好 / 低好 / 平衡
}

var translations = map[string]Translation{
	"retention_d1":              {"retention_d1", "新用户第二天还愿不愿意回来", "价值", "核心循环首次反馈、新手引导、首日奖励", "高好"},
	"retention_d7":              {"retention_d7", "新用户第一周是否留下来", "价值", "首周目标链、社交绑定、成长获得感", "高好"},
	"retention_d30":             {"retention_d30", "一个月后还有多少人在玩", "价值", "版本内容、长期目标、公会/赛季", "高好"},
	"economy_source_sink_ratio": {"economy_source_sink_ratio", "货币产出和消耗是否平衡", "流动", "产出阀门、消耗口、回收活动", "平衡"},
	"economy_inflation":         {"economy_inflation", "物价是否稳定", "流动", "货币回收、产出上限、定价校正", "平衡"},
	"monetization_pay_rate":     {"monetization_pay_rate", "多少人愿意付钱", "价值", "付费点价值感、首充设计、定价", "高好"},
	"monetization_arpu":         {"monetization_arpu", "每个用户平均贡献", "价值", "付费深度、礼包结构、活动", "高好"},
	"monetization_ltv":          {"monetization_ltv", "一个用户全生命周期值多少", "价值", "留存 × 付费深度、长期内容", "高好"},
	"progression_pace":          {"progression_pace", "玩家升级速度是否符合设计", "节奏", "经验曲线、日常投放、体力限制", "平衡"},
	"content_cadence_days":      {"content_cadence_days", "多久给玩家一次新内容", "节奏", "内容产能、活动排期、版本节奏", "平衡"},
	"content_debt_ratio":        {"content_debt_ratio", "内容积压是否失控", "节奏", "扩产、降低消耗速率、复用到期内容", "平衡"},
}

// Translate returns boss language for a metric key.
func Translate(key string) Translation {
	if t, ok := translations[key]; ok {
		return t
	}
	return Translation{Key: key, Plain: key, Question: "价值", Lever: "检查设计与配置", Direction: "平衡"}
}

// Finding is one diagnosed problem in boss language.
type Finding struct {
	Key       string
	Plain     string
	Question  string
	Lever     string
	Severity  string // critical | warn
	Observed  string
	Ideal     string
	Gap       string // 精确差距
	Headline  string // 一句话说明问题
	Consequence string // 不处理的后果
}

// DecisionItem is something only the boss can sign off.
type DecisionItem struct {
	ID          string
	Question    string // 需要拍板的问题
	Context     string // 为什么现在要决定
	Evidence    string // 精确数字证据
	Options     []string
	Recommend   string // 系统建议
	IfNot       string // 不决定的后果
	Owner       string // 主策 / 运营 / 老板
	Deadline    string
	Priority    string // P0 / P1 / P2
}

// Action is an executable instruction for the team.
type Action struct {
	ID       string
	Owner    string
	Task     string
	Tool     string
	Verify   string
	Priority string
}

// Brief is the boss-facing decision packet.
type Brief struct {
	Genre       string
	GenreName   string
	Headline    string   // 一句话结论
	Call        string   // GO / HOLD / STOP
	Score       string
	Confidence  string
	Findings    []Finding
	Decisions   []DecisionItem
	Actions     []Action
	Duties      []DutyItem
	Limits      []string // 数值无法判断的事
	Generated   string
}

// Analyze builds a boss-ready decision packet from exact metric values.
// Unknown genres are auto-derived, so ANY category works.
func Analyze(genreID string, values map[string]num.Num, flags map[string]bool) (Brief, error) {
	g := genre.GetAny(genreID)
	if values == nil {
		values = map[string]num.Num{}
	}
	r, err := ops.EvaluateAny(ops.Snapshot{GenreID: genreID, Values: values})
	if err != nil {
		return Brief{}, err
	}
	if len(values) == 0 {
		for _, l := range g.Levers {
			if _, ok := values[l.Key]; !ok {
				values[l.Key] = l.Ideal
			}
		}
		r, err = ops.EvaluateAny(ops.Snapshot{GenreID: genreID, Values: values})
		if err != nil {
			return Brief{}, err
		}
	}

	b := Brief{
		Genre: g.ID, GenreName: g.Name, Score: r.Score.DecimalString(4),
		Generated: nowString(),
	}
	// 结论一行：经营语言，不出现指标 key。
	switch r.Verdict {
	case "healthy":
		b.Call = "GO"
		b.Headline = "数值健康，可以按计划推进；保持当前节奏并持续监控。"
	case "attention":
		b.Call = "HOLD"
		b.Headline = "有指标偏离目标，建议暂缓发布，先完成本简报中的拍板项。"
	default:
		b.Call = "STOP"
		b.Headline = "存在严重数值风险，不建议发布；需主策与运营联合整改后重测。"
	}
	b.Confidence = "基于 11 项通用指标与品类 lever 的精确计算；数值是领先指标，不能替代对趣味性、社交与创意的判断。"

	// Findings：按加权偏差排序，翻译为经营语言。
	type weighted struct {
		d ops.Diagnostic
		w num.Num
	}
	leverWeight := map[string]num.Num{}
	for _, l := range g.Levers {
		leverWeight[l.Key] = l.Weight
	}
	var ws []weighted
	for _, d := range r.Diagnostics {
		if d.Status == "ok" {
			continue
		}
		ws = append(ws, weighted{d, d.Deviation.Mul(leverWeight[d.Metric])})
	}
	sort.SliceStable(ws, func(i, j int) bool { return ws[i].w.Cmp(ws[j].w) > 0 })
	for _, item := range ws {
		d := item.d
		t := Translate(d.Metric)
		f := Finding{
			Key: d.Metric, Plain: t.Plain, Question: t.Question, Lever: t.Lever,
			Severity: d.Status, Observed: d.Observed.DecimalString(6),
			Ideal: d.Ideal.DecimalString(6), Gap: d.Deviation.DecimalString(6),
		}
		f.Headline = fmt.Sprintf("%s%s：当前 %s，目标 %s，差距 %s。", map[string]string{"critical": "严重偏离", "warn": "偏离"}[d.Status], t.Plain, f.Observed, f.Ideal, f.Gap)
		f.Consequence = consequence(d.Metric, d.Observed.Cmp(d.Ideal) > 0)
		b.Findings = append(b.Findings, f)
	}

	// Decisions：Top3 拍板项。
	for i, f := range b.Findings {
		if i >= 3 {
			break
		}
		prio := "P1"
		if f.Severity == "critical" {
			prio = "P0"
		}
		b.Decisions = append(b.Decisions, DecisionItem{
			ID: fmt.Sprintf("D%d", i+1),
			Question: fmt.Sprintf("是否批准按建议调整「%s」？", f.Plain),
			Context: f.Headline,
			Evidence: fmt.Sprintf("精确值：当前 %s / 目标 %s / 差距 %s（%s问题）", f.Observed, f.Ideal, f.Gap, f.Question),
			Options: []string{
				"批准：按建议杠杆调整并进入复测",
				"暂缓：接受当前偏差，记录为已知风险并设复盘日期",
				"改方案：要求主策/运营提交替代方案后再拍板",
			},
			Recommend: "批准方案一，要求 7 天内复测。",
			IfNot: f.Consequence,
			Owner: "老板拍板 · 主策执行" ,
			Deadline: "7 天内",
			Priority: prio,
		})
	}

	// Actions：对应的执行指令。
	for i, f := range b.Findings {
		if i >= 3 {
			break
		}
		b.Actions = append(b.Actions, Action{
			ID: fmt.Sprintf("A%d", i+1),
			Owner: map[string]string{"价值": "主策划", "流动": "主策划 + 运营", "节奏": "运营"}[f.Question],
			Task: fmt.Sprintf("调整「%s」：%s（当前 %s → 目标 %s）", f.Plain, f.Lever, f.Observed, f.Ideal),
			Tool: "cmd/catalog eval · cmd/opsreport",
			Verify: "改完后重跑 catalog eval，健康度需 ≥ 0.8",
			Priority: map[string]string{"critical": "P0", "warn": "P1"}[f.Severity],
		})
	}

	// Duties：把工作事务矩阵的当前状态填进去。
	byMetric := map[string]ops.Diagnostic{}
	for _, d := range r.Diagnostics {
		byMetric[d.Metric] = d
	}
	for _, duty := range DutyMatrix {
		item := duty
		if item.Metric != "" {
			if d, ok := byMetric[item.Metric]; ok {
				item.Status = d.Status
				item.Finding = fmt.Sprintf("%s：当前 %s / 目标 %s", Translate(item.Metric).Plain,
					d.Observed.DecimalString(4), d.Ideal.DecimalString(4))
			} else {
				item.Status = "idle"
				item.Finding = "本次未观测到该项数据"
			}
		}
		b.Duties = append(b.Duties, item)
	}

	b.Limits = []string{
		"数值无法判断趣味、社交黏性与创意质量——需人工体验评审。",
		"健康度基于 11 项通用指标；品类专属问题请用 catalog strategy 查看。",
		"未知品类由标签派生基准，建议用 JSON spec 校准后再用于正式决策。",
		"所有数值为精确有理数；展示值经过四舍五入，决策以精确值为准。",
	}
	if flags["region"] {
		b.Limits = append(b.Limits, "地区合规（CN 抽卡公示/防沉迷等）请用 catalog strategy -region 单独核对。")
	}
	return b, nil
}

func consequence(metric string, high bool) string {
	switch metric {
	case "retention_d1", "retention_d7", "retention_d30":
		if high {
			return "留存异常偏高通常意味着数据口径问题，先核查埋点再调整投放。"
		}
		return "用户持续流失，后续付费与内容投入都将被浪费，长期收入下滑。"
	case "economy_source_sink_ratio", "economy_inflation":
		if high {
			return "货币过剩引发通胀，老玩家资产贬值、经济信心崩塌。"
		}
		return "货币稀缺导致玩家卡进度，付费点被挤压，活跃下降。"
	case "monetization_pay_rate", "monetization_arpu", "monetization_ltv":
		if high {
			return "付费率异常偏高可能说明卖到了核心玩法，需排查合规与口碑风险。"
		}
		return "付费转化不足，收入无法覆盖内容成本，项目可持续性下降。"
	case "progression_pace":
		if high {
			return "玩家推进过快，内容被提前消耗，版本后期空洞化。"
		}
		return "玩家推进过慢，挫败感累积，留存与付费同时受损。"
	case "content_cadence_days", "content_debt_ratio":
		if high {
			return "内容供给跟不上消耗，玩家耗尽内容后流失。"
		}
		return "内容投放过密，团队疲于奔命且玩家来不及消化。"
	default:
		return "偏离品类基准，长期累积会放大为版本风险。"
	}
}

func nowString() string {
	return time.Now().Format("2006-01-02 15:04:05")
}
