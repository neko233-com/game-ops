// Package report turns exact game-ops analysis into a self-contained HTML
// dashboard for bosses and non-numerical readers. No external assets: all CSS
// and SVG are inline so the file can be opened offline or emailed.
package report

import (
	"bytes"
	"fmt"
	"html/template"
	"sort"
	"time"

	"github.com/neko233/game-ops/design"
	"github.com/neko233/game-ops/genre"
	"github.com/neko233/game-ops/num"
	"github.com/neko233/game-ops/ops"
)

// Spec is the input to Build. Numbers are strings to preserve precision.
type Spec struct {
	Title    string            `json:"title"`
	Subtitle string            `json:"subtitle"`
	Genre    string            `json:"genre"`
	Values   map[string]string `json:"values"`
	BaseUnit string            `json:"baseUnit"`
	Rates    []struct {
		From  string `json:"from"`
		To    string `json:"to"`
		Value string `json:"value"`
	} `json:"rates"`
	Conversions []struct {
		Label  string `json:"label"`
		From   string `json:"from"`
		Amount string `json:"amount"`
	} `json:"conversions"`
	Sources map[string]string `json:"sources"`
	Sinks   map[string]string `json:"sinks"`
	Plan    *planSpec         `json:"plan"`
	Levels  int               `json:"levels"`
}

type planSpec struct {
	Name        string    `json:"name"`
	Growth      curveSpec `json:"growth"`
	CombatPower curveSpec `json:"combatPower"`
	Damage      curveSpec `json:"damage"`
	HP          curveSpec `json:"hp"`
	Reward      curveSpec `json:"reward"`
	Price       curveSpec `json:"price"`
}

type curveSpec struct {
	Name string `json:"name"`
	Base string `json:"base"`
	Step string `json:"step"`
	Cap  string `json:"cap"`
}

func (c curveSpec) curve() design.Curve {
	return design.Curve{
		Name: c.Name, Base: num.MustFromString(c.Base),
		Step: num.MustFromString(c.Step), Cap: num.MustFromString(c.Cap),
	}
}

// PillarView is one of the five pillars in the dashboard.
type PillarView struct {
	Name, Desc, Status string
	Percent            float64
}

// DiagnosticView is one metric row.
type DiagnosticView struct {
	Metric, Meaning, Observed, Ideal, Status string
	Critical                                 bool
	Warn                                     bool
}

// FlowView holds the source/sink balance summary.
type FlowView struct {
	SourceTotal, SinkTotal, Ratio, Verdict, Note string
	Good                                         bool
}

// CurveView holds precomputed chart geometry.
type CurveView struct {
	AdditivePoints string
	Title          string
}

// Data is all the view data the template needs.
type Data struct {
	Title, Subtitle, Genre, GenreName string
	ScoreText, Verdict                string
	VerdictClass                      string
	OneLiner                          string
	ThreeLines                        []string
	Pillars                           []PillarView
	Diagnostics                       []DiagnosticView
	Flow                              FlowView
	Curve                             CurveView
	Decisions                         []string
	Generated                         string
	GaugeFill, GaugeRest              float64
}

var metricMeaning = map[string]string{
	"retention_d1":              "新玩家第二天还回来吗（越高越好）",
	"retention_d7":              "新玩家第七天还回来吗（越高越好）",
	"retention_d30":             "新玩家第三十天还回来吗（越高越好）",
	"economy_source_sink_ratio": "货币产出 ÷ 消耗，接近 1 最健康",
	"economy_inflation":         "物价指数，1 表示稳定",
	"monetization_pay_rate":     "付费玩家占比",
	"monetization_arpu":         "人均收入指数",
	"monetization_ltv":          "玩家终身价值指数",
	"progression_pace":          "玩家推进速度是否符合设计",
	"content_cadence_days":      "两次内容更新间隔天数",
	"content_debt_ratio":        "内容积压程度，1 表示刚好",
}

var pillarDefs = []struct {
	name, desc string
	keys       []string
}{
	{"留存", "玩家是否愿意回来", []string{"retention_d1", "retention_d7", "retention_d30"}},
	{"经济", "货币是否平衡", []string{"economy_source_sink_ratio", "economy_inflation"}},
	{"商业化", "收入是否健康", []string{"monetization_pay_rate", "monetization_arpu", "monetization_ltv"}},
	{"成长", "推进速度是否合理", []string{"progression_pace"}},
	{"内容", "更新是否跟得上消耗", []string{"content_cadence_days", "content_debt_ratio"}},
}

// Build assembles the view data from exact computations.
func Build(s Spec) (Data, error) {
	g, ok := genre.Get(s.Genre)
	if !ok {
		return Data{}, fmt.Errorf("report: unknown genre %q", s.Genre)
	}
	values := make(map[string]num.Num, len(s.Values))
	for k, v := range s.Values {
		n, err := num.FromString(v)
		if err != nil {
			return Data{}, fmt.Errorf("report: value %q: %w", k, err)
		}
		values[k] = n
	}
	result, err := ops.Evaluate(ops.Snapshot{GenreID: s.Genre, Values: values})
	if err != nil {
		return Data{}, err
	}

	d := Data{
		Title: s.Title, Subtitle: s.Subtitle, Genre: s.Genre, GenreName: g.Name,
		ScoreText: result.Score.DecimalString(2) + " / 1.00",
		Verdict:   verdictLabel(result.Verdict),
		Generated: time.Now().Format("2006-01-02 15:04"),
	}
	switch result.Verdict {
	case "healthy":
		d.VerdictClass = "good"
	case "attention":
		d.VerdictClass = "warn"
	default:
		d.VerdictClass = "bad"
	}
	const circumference = 2 * 3.14159 * 54
	fill := result.Score.Float64() * circumference
	d.GaugeFill = fill
	d.GaugeRest = circumference - fill

	// Three plain-language sentences: value, flow, pacing.
	base := s.BaseUnit
	if base == "" {
		base = "基准单位"
	}
	d.ThreeLines = []string{
		fmt.Sprintf("价值标尺：所有奖励和价格都换算成「%s」再比较，就像所有货币先换成人民币。", base),
		fmt.Sprintf("经济流动：钱从哪里来、到哪里去；产出与消耗的平衡比率是 %.2f（1.00 最理想）。", economyRatio(s.Sources, s.Sinks)),
		"价值节奏：成长和内容更新要保持平稳的线性节奏，避免乘法叠加造成数值膨胀。",
	}

	byMetric := map[string]ops.Diagnostic{}
	for _, diag := range result.Diagnostics {
		byMetric[diag.Metric] = diag
	}
	for _, def := range pillarDefs {
		view := PillarView{Name: def.name, Desc: def.desc, Status: "良好"}
		total := 0.0
		worst := "ok"
		for _, key := range def.keys {
			diag, ok := byMetric[key]
			if !ok {
				continue
			}
			total += diag.Fitness.Float64()
			if diag.Status == "critical" {
				worst = "critical"
			} else if diag.Status == "warn" && worst == "ok" {
				worst = "warn"
			}
		}
		if len(def.keys) > 0 {
			view.Percent = total / float64(len(def.keys)) * 100
		}
		view.Status = map[string]string{"ok": "良好", "warn": "关注", "critical": "告警"}[worst]
		d.Pillars = append(d.Pillars, view)
	}

	// Diagnostics sorted worst-first, with a plain-language meaning.
	diags := append([]ops.Diagnostic(nil), result.Diagnostics...)
	sort.Slice(diags, func(i, j int) bool {
		fi, fj := diags[i].Fitness.Float64(), diags[j].Fitness.Float64()
		if fi != fj {
			return fi < fj
		}
		return diags[i].Metric < diags[j].Metric
	})
	for _, diag := range diags {
		meaning := metricMeaning[diag.Metric]
		if meaning == "" {
			meaning = diag.Metric
		}
		d.Diagnostics = append(d.Diagnostics, DiagnosticView{
			Metric: diag.Metric, Meaning: meaning,
			Observed: diag.Observed.DecimalString(4), Ideal: diag.Ideal.DecimalString(4),
			Status:   map[string]string{"ok": "正常", "warn": "关注", "critical": "告警"}[diag.Status],
			Critical: diag.Status == "critical", Warn: diag.Status == "warn",
		})
	}

	// Value flow summary.
	sr, sink := sumStrings(s.Sources), sumStrings(s.Sinks)
	d.Flow = FlowView{
		SourceTotal: sr.DecimalString(4), SinkTotal: sink.DecimalString(4),
		Verdict: "平衡", Note: "钱进来的速度约等于花出去的速度，经济稳定。", Good: true,
	}
	if !sink.IsZero() {
		ratio, err := sr.Div(sink)
		if err == nil {
			d.Flow.Ratio = ratio.DecimalString(4)
			switch {
			case ratio.Cmp(num.MustFromString("11/10")) > 0:
				d.Flow.Verdict, d.Flow.Note, d.Flow.Good = "通胀风险", "钱进来太多、花出去太少，玩家手里的货币会越来越不值钱。", false
			case ratio.Cmp(num.MustFromString("9/10")) < 0:
				d.Flow.Verdict, d.Flow.Note, d.Flow.Good = "流通不足", "钱不够花，玩家会在成长上卡住。", false
			}
		}
	}

	// Additive growth chart geometry.
	if s.Plan != nil {
		curve := s.Plan.Growth.curve()
		if curve.Name == "" {
			curve.Name = "growth"
		}
		levels := s.Levels
		if levels < 2 {
			levels = 20
		}
		max := curve.At(uint(levels)).Float64()
		if max <= 0 {
			max = 1
		}
		points := ""
		for l := uint(1); l <= uint(levels); l++ {
			x := 50 + float64(l-1)/float64(levels-1)*660
			y := 230 - curve.At(l).Float64()/max*190
			points += fmt.Sprintf("%.1f,%.1f ", x, y)
		}
		d.Curve = CurveView{AdditivePoints: points, Title: curve.Name}
	}

	// Boss decisions from the worst diagnostics.
	for i, diag := range d.Diagnostics {
		if i >= 3 || (!diag.Critical && !diag.Warn) {
			break
		}
		d.Decisions = append(d.Decisions, fmt.Sprintf(
			"%s（当前 %s，目标 %s）：%s。请决定是否按建议调整配表。",
			diag.Meaning, diag.Observed, diag.Ideal, diag.Status))
	}
	if len(d.Decisions) == 0 {
		d.Decisions = []string{"当前没有需要紧急决策的数值问题；保持现有内容节奏并按周期复盘即可。"}
	}

	one := "整体数值健康，可以继续按当前节奏运营。"
	if result.Verdict == "attention" {
		one = "整体可用，但有几项指标需要关注，建议本版本内修正。"
	} else if result.Verdict == "critical" {
		one = "整体存在明显风险，建议先修复告警指标再推进新内容。"
	}
	d.OneLiner = one
	return d, nil
}

func sumStrings(m map[string]string) num.Num {
	total := num.Zero()
	for _, v := range m {
		if n, err := num.FromString(v); err == nil {
			total = total.Add(n)
		}
	}
	return total
}

func verdictLabel(v string) string {
	return map[string]string{"healthy": "健康", "attention": "需关注", "critical": "有风险"}[v]
}

func economyRatio(sources, sinks map[string]string) float64 {
	s, k := sumStrings(sources), sumStrings(sinks)
	if k.IsZero() {
		return 0
	}
	r, err := s.Div(k)
	if err != nil {
		return 0
	}
	return r.Float64()
}

// Render produces a self-contained HTML document.
func Render(d Data) ([]byte, error) {
	tmpl, err := template.New("dashboard").Funcs(template.FuncMap{
		"safe": func(s string) template.HTML { return template.HTML(s) },
	}).Parse(pageTemplate)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, d); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

const pageTemplate = `<!doctype html>
<html lang="zh-CN"><head><meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>{{.Title}}</title>
<style>
*{box-sizing:border-box}body{margin:0;font-family:-apple-system,"Segoe UI","Microsoft YaHei",sans-serif;background:#f4f6fa;color:#1c2430}
.wrap{max-width:1080px;margin:0 auto;padding:28px 20px 60px}
header{border-left:6px solid #3b6ef5;padding-left:16px;margin-bottom:24px}
header h1{margin:0;font-size:26px}header p{margin:6px 0 0;color:#667085}
.hero{display:flex;gap:28px;align-items:center;background:#fff;border-radius:16px;padding:26px;box-shadow:0 4px 18px #1c24300d}
.gauge{position:relative;width:140px;height:140px;flex:none}.gauge svg{transform:rotate(-90deg)}
.gauge .val{position:absolute;inset:0;display:flex;align-items:center;justify-content:center;font-size:22px;font-weight:700}
.badge{display:inline-block;padding:5px 14px;border-radius:999px;font-weight:700;font-size:14px}
.good{background:#e7f8ee;color:#0a7a3d}.warn{background:#fff4de;color:#9a6700}.bad{background:#ffe8e6;color:#b42318}
.hero h2{margin:12px 0 6px;font-size:22px}.hero p{margin:0;color:#475467;line-height:1.6}
.grid{display:grid;grid-template-columns:repeat(3,1fr);gap:14px;margin-top:18px}
.card{background:#fff;border-radius:14px;padding:16px 18px;box-shadow:0 2px 10px #1c24300a}
.card h3{margin:0 0 6px;font-size:15px}.card p{margin:0;color:#475467;font-size:14px;line-height:1.6}
.num{font-size:26px;font-weight:800;color:#3b6ef5;margin:6px 0}
section{margin-top:26px}section>h2{font-size:19px;margin:0 0 12px}
.pillars{display:grid;grid-template-columns:repeat(5,1fr);gap:10px}
.pillar{background:#fff;border-radius:12px;padding:14px;text-align:center;box-shadow:0 2px 10px #1c24300a}
.pillar .name{font-weight:700}.pillar .desc{font-size:12px;color:#667085;margin:4px 0 10px}
.bar{height:9px;background:#eef1f6;border-radius:9px;overflow:hidden}.bar i{display:block;height:100%;background:linear-gradient(90deg,#3b6ef5,#54b9f8)}
.pillar .st{font-size:12px;margin-top:8px;color:#0a7a3d}
table{width:100%;border-collapse:collapse;background:#fff;border-radius:12px;overflow:hidden;box-shadow:0 2px 10px #1c24300a}
th,td{padding:10px 12px;text-align:left;font-size:14px;border-bottom:1px solid #eef1f6}th{background:#f9fafb;color:#475467}
.dot{display:inline-block;width:9px;height:9px;border-radius:50%;margin-right:6px}.ok{background:#12b76a}.w2{background:#f79009}.c{background:#f04438}
.flow{display:flex;align-items:center;gap:12px;background:#fff;border-radius:14px;padding:20px;box-shadow:0 2px 10px #1c24300a}
.box{flex:1;background:#f8faff;border:2px solid #d5e2ff;border-radius:12px;padding:14px;text-align:center}
.box b{font-size:16px}.box small{display:block;color:#667085;margin-top:6px}
.arrow{font-size:26px;color:#3b6ef5}.flownote{margin-top:12px;font-size:14px;color:#475467}
ol{background:#fff;border-radius:12px;padding:18px 18px 18px 40px;box-shadow:0 2px 10px #1c24300a}
li{margin:8px 0;line-height:1.6}
footer{margin-top:30px;color:#98a2b3;font-size:12px;text-align:center}
@media(max-width:760px){.grid,.pillars{grid-template-columns:1fr 1fr}.hero{flex-direction:column}.flow{flex-direction:column}}
</style></head><body><div class="wrap">
<header><h1>{{.Title}}</h1><p>{{.Subtitle}} · 品类：{{.GenreName}}（{{.Genre}}） · 生成于 {{.Generated}}</p></header>
<div class="hero">
 <div class="gauge"><svg width="140" height="140" viewBox="0 0 140 140">
  <circle cx="70" cy="70" r="54" fill="none" stroke="#eef1f6" stroke-width="14"/>
  <circle cx="70" cy="70" r="54" fill="none" stroke="#3b6ef5" stroke-width="14" stroke-linecap="round" stroke-dasharray="{{printf "%.2f %.2f" .GaugeFill .GaugeRest}}"/>
 </svg><div class="val">{{.ScoreText}}</div></div>
 <div><span class="badge {{.VerdictClass}}">{{.Verdict}}</span><h2>{{.OneLiner}}</h2>
 <p>健康度是把留存、经济、商业化、成长、内容五类指标按重要性加权得到的 0–1 分数。越接近 1.00 越健康。</p></div>
</div>
<div class="grid">
 {{range .ThreeLines}}<div class="card"><p>{{.}}</p></div>{{end}}
</div>
<section><h2>五大支柱体检</h2><div class="pillars">
 {{range .Pillars}}<div class="pillar"><div class="name">{{.Name}}</div><div class="desc">{{.Desc}}</div>
 <div class="bar"><i style="width:{{printf "%.1f" .Percent}}%"></i></div><div class="st">{{.Status}} · {{printf "%.0f" .Percent}}%</div></div>{{end}}
</div></section>
<section><h2>指标明细（从最需要关注的排起）</h2><table>
<tr><th>指标</th><th>通俗解释</th><th>当前</th><th>目标</th><th>状态</th></tr>
{{range .Diagnostics}}<tr><td><span class="dot {{if .Critical}}c{{else if .Warn}}w2{{else}}ok{{end}}"></span>{{.Metric}}</td>
<td>{{.Meaning}}</td><td>{{.Observed}}</td><td>{{.Ideal}}</td><td>{{.Status}}</td></tr>{{end}}
</table></section>
<section><h2>钱是怎么流动的</h2><div class="flow">
 <div class="box"><b>钱从哪里来</b><small>任务 / 关卡 / 充值<br>合计 {{.Flow.SourceTotal}}</small></div>
 <div class="arrow">➜</div><div class="box"><b>玩家钱包</b><small>产出 ÷ 消耗 = {{.Flow.Ratio}}</small></div>
 <div class="arrow">➜</div><div class="box"><b>钱花到哪里去</b><small>升级 / 抽卡 / 门票<br>合计 {{.Flow.SinkTotal}}</small></div>
</div><p class="flownote"><b>{{.Flow.Verdict}}</b>：{{.Flow.Note}}</p></section>
<section><h2>成长曲线：为什么用加法</h2><div class="card">
<p>{{.Curve.Title}} 曲线采用「起点 + 每级固定增量」的加法式，玩家每升一级得到的好处一样多，既好预测也好封顶。</p>
<svg width="740" height="260" viewBox="0 0 740 260" role="img" aria-label="加法成长曲线">
 <line x1="50" y1="230" x2="710" y2="230" stroke="#98a2b3"/><line x1="50" y1="40" x2="50" y2="230" stroke="#98a2b3"/>
 <polyline points="{{.Curve.AdditivePoints}}" fill="none" stroke="#3b6ef5" stroke-width="3"/>
 <text x="50" y="252" font-size="12" fill="#667085">等级 1</text><text x="660" y="252" font-size="12" fill="#667085">高等级</text>
 <text x="60" y="32" font-size="13" fill="#3b6ef5">加法：平稳、可封顶</text>
</svg>
<p>如果用乘法（例如每级 ×1.1），第 40 级就会变成约 45 倍——数值膨胀会让老玩家碾压新玩家、经济失控。加法曲线避免这一点。</p>
</div></section>
<section><h2>老板需要决策的三件事</h2><ol>{{range .Decisions}}<li>{{.}}</li>{{end}}</ol></section>
<footer>本报告由 game-ops 自动生成 · 所有数值以精确有理数计算 · 指标计算式：fitness = clamp(1 − |当前−目标| ÷ 容差带, 0, 1)</footer>
</div></body></html>`
