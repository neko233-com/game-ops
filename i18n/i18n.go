// Package i18n provides the translation catalog for reports and CLIs.
// Chinese (zh) is the authoritative, complete language and the default;
// English and Japanese are supported alternatives. Unknown locales fall back
// to Chinese rather than English, per "优先中文".
package i18n

import (
	"sort"
	"strings"
)

// Lang is a supported language code.
type Lang string

const (
	ZH Lang = "zh" // authoritative primary
	EN Lang = "en"
	JA Lang = "ja"
)

// Default returns the primary language (Chinese).
func Default() Lang { return ZH }

// Langs returns all supported languages, primary first.
func Langs() []Lang { return []Lang{ZH, EN, JA} }

// Label returns a language's display name in its own script.
func Label(l Lang) string {
	switch l {
	case EN:
		return "English"
	case JA:
		return "日本語"
	default:
		return "中文"
	}
}

// Parse maps a locale string to a supported language. Anything unknown —
// including an empty string — falls back to Chinese.
func Parse(s string) Lang {
	lower := strings.ToLower(strings.TrimSpace(s))
	switch {
	case lower == "", lower == "zh", strings.HasPrefix(lower, "zh-"), lower == "中文", strings.HasPrefix(lower, "chinese"):
		return ZH
	case lower == "en", strings.HasPrefix(lower, "en-"), strings.HasPrefix(lower, "english"):
		return EN
	case lower == "ja", strings.HasPrefix(lower, "ja-"), lower == "日本語", strings.HasPrefix(lower, "japanese"):
		return JA
	default:
		return ZH
	}
}

var catalog = map[string]map[Lang]string{
	"title.report":                     {"zh": "版本数值健康报告", "en": "Build Numerical Health Report", "ja": "バージョン数値健全性レポート"},
	"title.proanalysis":                {"zh": "满级节奏与配表专业分析（数值主策划）", "en": "Time-to-Cap & Config Analysis (Lead Designer)", "ja": "満級ペースと配表分析（数値主任）"},
	"label.generated":                  {"zh": "生成时间", "en": "Generated", "ja": "生成日時"},
	"label.baseUnit":                   {"zh": "基准单位", "en": "Base unit", "ja": "基準単位"},
	"label.genre":                      {"zh": "品类", "en": "Genre", "ja": "ジャンル"},
	"label.exact":                      {"zh": "精确有理数计算，无浮点误差", "en": "Exact rational arithmetic, no floating-point error", "ja": "厳密有理数計算、浮動小数点誤差なし"},
	"label.parallel":                   {"zh": "并行度", "en": "Workers", "ja": "並列数"},
	"label.scenarios":                  {"zh": "场景数", "en": "Scenarios", "ja": "シナリオ数"},
	"label.config":                     {"zh": "配表", "en": "Config tables", "ja": "設定表"},
	"label.files":                      {"zh": "文件", "en": "files", "ja": "ファイル"},
	"label.rows":                       {"zh": "行", "en": "rows", "ja": "行"},
	"label.elapsed":                    {"zh": "耗时", "en": "Elapsed", "ja": "所要時間"},
	"section.method":                   {"zh": "方法与输入", "en": "Method & inputs", "ja": "手法と入力"},
	"section.capacity":                 {"zh": "常规/轻度/重度玩家满级时间", "en": "Time-to-cap by archetype", "ja": "プレイヤー別の満級日数"},
	"section.sensitivity":              {"zh": "敏感度扫描", "en": "Sensitivity sweep", "ja": "感度スキャン"},
	"section.qa":                       {"zh": "配表异常", "en": "Config anomalies", "ja": "設定表の異常"},
	"section.recommendations":          {"zh": "结论与建议", "en": "Conclusions & recommendations", "ja": "結論と提案"},
	"section.pillars":                  {"zh": "五大支柱", "en": "Five pillars", "ja": "五本柱"},
	"section.metrics":                  {"zh": "指标明细", "en": "Metric details", "ja": "指標の詳細"},
	"section.flow":                     {"zh": "价值流动", "en": "Value flow", "ja": "価値の流れ"},
	"section.curve":                    {"zh": "成长曲线", "en": "Growth curve", "ja": "成長曲線"},
	"section.risks":                    {"zh": "风险与调整建议", "en": "Risks & adjustments", "ja": "リスクと調整案"},
	"method.totalXP":                   {"zh": "总经验 = Σ(l=1..满级-1) [base + step·(l−1)]，精确求和。", "en": "Total XP = Σ(l=1..cap-1) [base + step·(l−1)], summed exactly.", "ja": "総経験値 = Σ(l=1..満級-1) [base + step·(l−1)]、厳密に合計。"},
	"method.dailyXP":                   {"zh": "每日经验 = Σ(活动分钟 × 每分钟经验) × 玩家效率；活动按日上限顺序分配时间。", "en": "Daily XP = Σ(minutes × XP/min) × efficiency; time allocated in activity order under daily caps.", "ja": "1日あたり経験値 = Σ(分数 × XP/分) × 効率。日次上限の順に時間を配分。"},
	"method.days":                      {"zh": "满级天数 = 总经验 ÷ 每日经验（精确分数）；留存触达率 = D1/D7/D30 锚点线性插值外推。", "en": "Days = total ÷ daily (exact fraction); survival = linear interpolation/extrapolation of D1/D7/D30.", "ja": "満級日数 = 総経験 ÷ 日次経験（厳密分数）。生存率は D1/D7/D30 の線形補間・外挿。"},
	"column.archetype":                 {"zh": "玩家类型", "en": "Archetype", "ja": "プレイヤー種別"},
	"column.totalXP":                   {"zh": "总经验", "en": "Total XP", "ja": "総経験値"},
	"column.dailyXP":                   {"zh": "每日经验", "en": "Daily XP", "ja": "日次経験値"},
	"column.daysExact":                 {"zh": "满级天数（精确）", "en": "Days (exact)", "ja": "満級日数（厳密）"},
	"column.daysApprox":                {"zh": "满级天数（约）", "en": "Days (approx)", "ja": "満級日数（概算）"},
	"column.survival":                  {"zh": "满级时留存", "en": "Survival at cap", "ja": "満級時の生存率"},
	"column.weight":                    {"zh": "人口权重", "en": "Population weight", "ja": "人口比率"},
	"column.metric":                    {"zh": "指标", "en": "Metric", "ja": "指標"},
	"column.meaning":                   {"zh": "含义", "en": "Meaning", "ja": "意味"},
	"column.observed":                  {"zh": "当前", "en": "Observed", "ja": "観測値"},
	"column.ideal":                     {"zh": "目标", "en": "Ideal", "ja": "目標値"},
	"column.fitness":                   {"zh": "健康度", "en": "Fitness", "ja": "適合度"},
	"column.status":                    {"zh": "状态", "en": "Status", "ja": "状態"},
	"status.ok":                        {"zh": "正常", "en": "ok", "ja": "正常"},
	"status.warn":                      {"zh": "关注", "en": "warn", "ja": "注意"},
	"status.critical":                  {"zh": "告警", "en": "critical", "ja": "警告"},
	"verdict.healthy":                  {"zh": "健康", "en": "healthy", "ja": "健全"},
	"verdict.attention":                {"zh": "需关注", "en": "attention", "ja": "要注意"},
	"verdict.critical":                 {"zh": "有风险", "en": "critical", "ja": "リスク"},
	"advice.rebalance":                 {"zh": "建议调整该指标向目标靠拢，并说明对价值/流动/节奏的影响", "en": "Recommend moving this metric toward its ideal; explain the value/flow/pacing impact", "ja": "目標値に近づける提案と、価値・流れ・ペースへの影響を説明する"},
	"advice.verify":                    {"zh": "所有建议需用 numgen、ingest、ops 重新验证后再进入版本。", "en": "Re-validate every change with numgen, ingest and ops before shipping.", "ja": "すべての変更は numgen / ingest / ops で再検証してからリリースする。"},
	"note.noIssues":                    {"zh": "当前没有超出容差带的指标。保持节奏，按周期复盘。", "en": "No metric is outside its band. Keep the cadence and review periodically.", "ja": "許容範囲外の指標はありません。ペースを維持し定期レビューする。"},
	"note.million":                     {"zh": "支持百万级配表流式并行分析", "en": "Streaming parallel analysis of up to 1,000,000 config files", "ja": "最大100万件の設定表をストリーミング並列解析"},
	"chart.additive":                   {"zh": "加法：平稳、可封顶", "en": "Additive: smooth, capped", "ja": "加法：安定・上限あり"},
	"label.method":                     {"zh": "方法", "en": "method", "ja": "手法"},
	"label.sources":                    {"zh": "钱从哪里来", "en": "Sources", "ja": "流入"},
	"label.sinks":                      {"zh": "钱花到哪里去", "en": "Sinks", "ja": "流出"},
	"label.wallet":                     {"zh": "玩家钱包", "en": "Player wallet", "ja": "プレイヤーの財布"},
	"label.ratio":                      {"zh": "产出 ÷ 消耗", "en": "Sources ÷ sinks", "ja": "流入 ÷ 流出"},
	"label.ideal1":                     {"zh": "理想值 1", "en": "Ideal: 1", "ja": "理想値 1"},
	"label.top3":                       {"zh": "按严重度排序的前三项", "en": "Top 3 by severity", "ja": "重要度上位3件"},
	"label.score":                      {"zh": "加权分数", "en": "weighted score", "ja": "加重スコア"},
	"metric.retention_d1":              {"zh": "D1 次日留存（越高越好）", "en": "D1 next-day retention (higher is better)", "ja": "D1 翌日リテンション（高いほど良い）"},
	"metric.retention_d7":              {"zh": "D7 七日留存（越高越好）", "en": "D7 seven-day retention (higher is better)", "ja": "D7 七日リテンション（高いほど良い）"},
	"metric.retention_d30":             {"zh": "D30 三十日留存（越高越好）", "en": "D30 thirty-day retention (higher is better)", "ja": "D30 三十日リテンション（高いほど良い）"},
	"metric.economy_source_sink_ratio": {"zh": "货币产出 ÷ 消耗，理想值 1", "en": "Currency sources ÷ sinks; ideal 1", "ja": "通貨の流入 ÷ 流出、理想値 1"},
	"metric.economy_inflation":         {"zh": "物价指数，1 表示稳定", "en": "Price index; 1 is stable", "ja": "物価指数、1 で安定"},
	"metric.monetization_pay_rate":     {"zh": "付费转化率", "en": "Payer conversion rate", "ja": "課金転換率"},
	"metric.monetization_arpu":         {"zh": "ARPU 归一化指数", "en": "Normalized ARPU index", "ja": "正規化 ARPU 指数"},
	"metric.monetization_ltv":          {"zh": "LTV 归一化指数", "en": "Normalized LTV index", "ja": "正規化 LTV 指数"},
	"metric.progression_pace":          {"zh": "实际 ÷ 设计推进速度", "en": "Actual ÷ intended progression pace", "ja": "実測 ÷ 設計上の進行速度"},
	"metric.content_cadence_days":      {"zh": "内容投放间隔（天）", "en": "Days between content drops", "ja": "コンテンツ配信間隔（日）"},
	"metric.content_debt_ratio":        {"zh": "内容积压比，1 健康", "en": "Content debt ratio; 1 is healthy", "ja": "コンテンツ負債比率、1 が健全"},
	"score":                            {"zh": "分数", "en": "score", "ja": "スコア"},
	"meta.generated":                   {"zh": "生成时间", "en": "Generated", "ja": "生成日時"},
	"meta.curve":                       {"zh": "经验曲线", "en": "XP curve", "ja": "経験曲線"},
	"meta.baseMaxLevel":                {"zh": "基准满级", "en": "base cap", "ja": "基準満級"},
	"section.method.title":             {"zh": "方法与输入", "en": "Method & inputs", "ja": "手法と入力"},
	"section.capacity.title":           {"zh": "常规/轻度/重度玩家满级时间", "en": "Time-to-cap by archetype", "ja": "プレイヤー別の満級日数"},
	"section.sensitivity.title":        {"zh": "敏感度扫描", "en": "Sensitivity sweep", "ja": "感度スキャン"},
	"section.qa.title":                 {"zh": "配表异常", "en": "Config anomalies", "ja": "設定表の異常"},
	"section.recommendations.title":    {"zh": "结论与建议", "en": "Conclusions & recommendations", "ja": "結論と提案"},
	"column.scenario":                  {"zh": "场景", "en": "Scenario", "ja": "シナリオ"},
	"column.cap":                       {"zh": "满级", "en": "Cap", "ja": "満級"},
	"column.normalDays":                {"zh": "常规天数", "en": "Normal days", "ja": "標準日数"},
	"column.lightDays":                 {"zh": "轻度天数", "en": "Light days", "ja": "軽度日数"},
	"column.heavyDays":                 {"zh": "重度天数", "en": "Heavy days", "ja": "重度日数"},
	"column.p50Days":                   {"zh": "P50 天数", "en": "P50 days", "ja": "P50 日数"},
	"label.weightedPercentiles":        {"zh": "加权分位数（按人口权重）", "en": "Weighted percentiles (by population weight)", "ja": "加重分位数（人口比率別）"},
	"label.parseErrors":                {"zh": "解析错误", "en": "Parse errors", "ja": "解析エラー"},
	"label.columns":                    {"zh": "列数", "en": "Columns", "ja": "列"},
	"note.noColumns":                   {"zh": "无可分析列。", "en": "No analyzable columns.", "ja": "分析可能な列がありません。"},
	"flag.outlier":                     {"zh": "超过 10×，需人工确认", "en": "over 10×, needs manual review", "ja": "10倍超、手動確認が必要"},
	"conclusion.long":                  {"zh": "满级周期偏长：P50 超过 120 天。建议上调经验曲线 step、提高活动经验密度，或降低满级门槛；用第 3 节敏感度表选择目标档位。", "en": "Time-to-cap is too long: P50 exceeds 120 days. Raise the XP step, increase activity XP density, or lower the cap; pick a target from the sensitivity table.", "ja": "満級までの期間が長すぎます。P50が120日超。step引き上げ・経験密度増・満級引き下げを検討し、感度表から目標を選択。"},
	"conclusion.short":                 {"zh": "满级周期偏短：P50 不足 60 天，长期内容消耗可能过快。建议下调 step 或补充满级后追求目标。", "en": "Time-to-cap is too short: P50 under 60 days risks burning content too fast. Lower the step or add post-cap goals.", "ja": "満級期間が短すぎます。P50が60日未満ではコンテンツ消費が速すぎる可能性。stepを下げるか満級後の目標を追加。"},
	"conclusion.ok":                    {"zh": "满级周期在合理区间：P50 处于 60–120 天。", "en": "Time-to-cap is in the healthy range: P50 is 60–120 days.", "ja": "満級期間は適正範囲です。P50は60〜120日。"},
	"conclusion.survival":              {"zh": "满级触达率过低：常规玩家满级时留存仅 %s，多数玩家在满级前流失；建议把关键奖励前移，或降低满级所需总经验。", "en": "Survival at cap is too low: only %s of normal players remain at cap; most churn earlier. Move key rewards earlier or lower total XP.", "ja": "満級時の生存率が低すぎます。標準プレイヤーの生存率はわずか %s。重要報酬を前倒しするか総経験値を下げる。"},
	"conclusion.config":                {"zh": "配表异常：`%s` 的 max/min 达 %s，检查是否误填单位或指数公式。", "en": "Config anomaly: `%s` has max/min = %s; check for unit errors or exponential formulas.", "ja": "設定表の異常：`%s` の max/min が %s。単位の誤記や指数式を確認。"},
}

// T translates a key into a language, falling back to Chinese and finally to
// the key itself so a missing entry is visible rather than silent.
func T(lang Lang, key string) string {
	if m, ok := catalog[key]; ok {
		if v, ok := m[lang]; ok && v != "" {
			return v
		}
		if v, ok := m[ZH]; ok && v != "" {
			return v
		}
	}
	return key
}

// Keys returns all catalog keys, sorted for deterministic iteration.
func Keys() []string {
	out := make([]string, 0, len(catalog))
	for k := range catalog {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// All returns the full catalog (language -> key -> text).
func All() map[Lang]map[string]string {
	out := map[Lang]map[string]string{}
	for _, l := range Langs() {
		out[l] = map[string]string{}
		for key := range catalog {
			out[l][key] = T(l, key)
		}
	}
	return out
}
