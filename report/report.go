// Package report turns exact game-ops analysis into a self-contained HTML
// document for the lead numerical designer. Styling follows the Vercel design
// language: monochrome, hairline borders, Geist-style type, monospace numbers.
// All CSS and SVG are inline, so the file works offline and in email.
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

// PillarView is one of the five pillars.
type PillarView struct {
	Name, Desc, Status string
	Percent            float64
}

// DiagnosticView is one metric row.
type DiagnosticView struct {
	Metric, Meaning, Observed, Ideal, Status, Fitness string
	Critical                                          bool
	Warn                                              bool
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
	ScoreText, ScoreExact, Verdict    string
	VerdictClass                      string
	Critical, Warn                    bool
	OneLiner                          string
	MethodLines                       []string
	Pillars                           []PillarView
	Diagnostics                       []DiagnosticView
	Flow                              FlowView
	Curve                             CurveView
	Decisions                         []string
	Generated                         string
	GaugeFill, GaugeRest              float64
}

var metricMeaning = map[string]string{
	"retention_d1":              "D1 次日留存（越高越好）",
	"retention_d7":              "D7 七日留存（越高越好）",
	"retention_d30":             "D30 三十日留存（越高越好）",
	"economy_source_sink_ratio": "货币产出 ÷ 消耗，理想值 1",
	"economy_inflation":         "物价指数，1 表示稳定",
	"monetization_pay_rate":     "付费转化率",
	"monetization_arpu":         "ARPU 归一化指数",
	"monetization_ltv":          "LTV 归一化指数",
	"progression_pace":          "实际 ÷ 设计推进速度",
	"content_cadence_days":      "内容投放间隔（天）",
	"content_debt_ratio":        "内容积压比，1 健康",
}

var pillarDefs = []struct {
	name, desc string
	keys       []string
}{
	{"留存 / Retention", "留存衰减与回流动量", []string{"retention_d1", "retention_d7", "retention_d30"}},
	{"经济 / Economy", "产出、消耗与通胀", []string{"economy_source_sink_ratio", "economy_inflation"}},
	{"商业化 / Monetization", "转化率、ARPU、LTV", []string{"monetization_pay_rate", "monetization_arpu", "monetization_ltv"}},
	{"成长 / Progression", "推进速度是否符合设计", []string{"progression_pace"}},
	{"内容 / Content", "投放节奏与内容负债", []string{"content_cadence_days", "content_debt_ratio"}},
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
		ScoreText: result.Score.DecimalString(4), ScoreExact: result.Score.String(),
		Verdict:   verdictLabel(result.Verdict),
		Generated: time.Now().Format("2006-01-02 15:04:05"),
	}
	switch result.Verdict {
	case "healthy":
		d.VerdictClass = "good"
	case "attention":
		d.VerdictClass, d.Warn = "warn", true
	default:
		d.VerdictClass, d.Critical = "bad", true
	}
	const circumference = 2 * 3.14159 * 54
	d.GaugeFill = result.Score.Float64() * circumference
	d.GaugeRest = circumference - d.GaugeFill

	base := s.BaseUnit
	if base == "" {
		base = "基准单位"
	}
	d.MethodLines = []string{
		fmt.Sprintf("单位价值：所有资源以精确有理数折算到「%s」，汇率来自配表而非手填。", base),
		fmt.Sprintf("source/sink = %s：产出与消耗的比率，理想值 1；偏离即通胀或稀缺。", divSafe(sumStrings(s.Sources), sumStrings(s.Sinks))),
		"fitness = clamp(1 − |observed − ideal| ÷ ((max − min)/2), 0, 1)；score = Σ(fitness × weight) ÷ Σweight。",
	}

	byMetric := map[string]ops.Diagnostic{}
	for _, diag := range result.Diagnostics {
		byMetric[diag.Metric] = diag
	}
	for _, def := range pillarDefs {
		view := PillarView{Name: def.name, Desc: def.desc, Status: "ok"}
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
		view.Status = map[string]string{"ok": "ok", "warn": "warn", "critical": "critical"}[worst]
		d.Pillars = append(d.Pillars, view)
	}

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
			Observed: diag.Observed.DecimalString(6), Ideal: diag.Ideal.DecimalString(6),
			Fitness:  diag.Fitness.DecimalString(4),
			Status:   map[string]string{"ok": "ok", "warn": "warn", "critical": "critical"}[diag.Status],
			Critical: diag.Status == "critical", Warn: diag.Status == "warn",
		})
	}

	sr, sink := sumStrings(s.Sources), sumStrings(s.Sinks)
	d.Flow = FlowView{
		SourceTotal: sr.DecimalString(4), SinkTotal: sink.DecimalString(4),
		Verdict: "balanced", Note: "产出与消耗平衡，经济稳定。", Good: true,
	}
	if !sink.IsZero() {
		ratio, err := sr.Div(sink)
		if err == nil {
			d.Flow.Ratio = ratio.DecimalString(4)
			switch {
			case ratio.Cmp(num.MustFromString("11/10")) > 0:
				d.Flow.Verdict, d.Flow.Note, d.Flow.Good = "inflation", "净注入过多，货币存量与物价上行。", false
			case ratio.Cmp(num.MustFromString("9/10")) < 0:
				d.Flow.Verdict, d.Flow.Note, d.Flow.Good = "scarcity", "回收大于产出，成长会卡住。", false
			}
		}
	}

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

	for i, diag := range d.Diagnostics {
		if i >= 3 || (!diag.Critical && !diag.Warn) {
			break
		}
		d.Decisions = append(d.Decisions, fmt.Sprintf(
			"%s：observed=%s, ideal=%s, fitness=%s — 需评估是否调整配表。",
			diag.Metric, diag.Observed, diag.Ideal, diag.Fitness))
	}
	if len(d.Decisions) == 0 {
		d.Decisions = []string{"无越界指标。维持当前节奏，按内容 cadence 复盘。"}
	}
	d.OneLiner = "数值健康，按当前节奏推进。"
	if result.Verdict == "attention" {
		d.OneLiner = "存在关注项，建议本版本内修正后复核。"
	} else if result.Verdict == "critical" {
		d.OneLiner = "存在越界指标，建议先修复再推进新内容。"
	}
	return d, nil
}

// divSafe formats an exact ratio without panicking on a zero denominator.
func divSafe(a, b num.Num) string {
	if b.IsZero() {
		return "n/a"
	}
	r, err := a.Div(b)
	if err != nil {
		return "n/a"
	}
	return r.DecimalString(4)
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
	return map[string]string{"healthy": "healthy", "attention": "attention", "critical": "critical"}[v]
}

// Render produces a self-contained Vercel-styled HTML document.
func Render(d Data) ([]byte, error) {
	tmpl, err := template.New("dashboard").Parse(pageTemplate)
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
:root{--fg:#000;--muted:#666;--line:#eaeaea;--accent:#0070f3;--bg:#fff;--sub:#fafafa;--good:#0cce6b;--warn:#f5a623;--bad:#ee0000}
*{box-sizing:border-box}html,body{margin:0;padding:0}body{background:var(--bg);color:var(--fg);font-family:Geist,Inter,-apple-system,BlinkMacSystemFont,"Segoe UI","PingFang SC","Microsoft YaHei",sans-serif;font-size:14px;line-height:1.5;-webkit-font-smoothing:antialiased}
code,pre,.mono{font-family:"Geist Mono",ui-monospace,SFMono-Regular,Menlo,Consolas,"Liberation Mono",monospace}
.wrap{max-width:1080px;margin:0 auto;padding:0 24px 80px}
.topbar{height:56px;border-bottom:1px solid var(--line);display:flex;align-items:center;justify-content:space-between;margin-bottom:48px}
.brand{display:flex;align-items:center;gap:10px;font-weight:600;letter-spacing:-.02em}.brand .mark{width:20px;height:20px;background:var(--fg);border-radius:4px;display:inline-flex;align-items:center;justify-content:center;color:#fff;font-size:12px}
.topbar .meta{color:var(--muted);font-size:12px}
.eyebrow{font-size:11px;text-transform:uppercase;letter-spacing:.12em;color:var(--muted);margin-bottom:12px}
h1{font-size:40px;line-height:1.15;letter-spacing:-.04em;margin:0 0 12px;font-weight:650}
.sub{color:var(--muted);font-size:15px;margin:0 0 36px}
.hero{border:1px solid var(--line);border-radius:8px;display:flex;align-items:center;gap:36px;padding:28px 32px}
.gauge{position:relative;width:132px;height:132px;flex:none}.gauge svg{transform:rotate(-90deg)}
.gauge .val{position:absolute;inset:0;display:flex;flex-direction:column;align-items:center;justify-content:center;gap:2px}
.gauge .val b{font-size:24px;letter-spacing:-.04em}.gauge .val small{color:var(--muted);font-size:10px;text-transform:uppercase;letter-spacing:.1em}
.badge{display:inline-flex;align-items:center;gap:6px;border:1px solid var(--line);border-radius:999px;padding:3px 10px;font-size:12px;font-family:ui-monospace,monospace}
.badge.good b{color:var(--good)}.badge.warn b{color:var(--warn)}.badge.bad b{color:var(--bad)}
.hero h2{font-size:18px;letter-spacing:-.02em;margin:14px 0 8px;font-weight:600}
.formula{border-top:1px solid var(--line);margin-top:28px;padding-top:20px;display:grid;grid-template-columns:repeat(3,1fr);gap:24px}
.formula div span{display:block;font-size:11px;text-transform:uppercase;letter-spacing:.1em;color:var(--muted);margin-bottom:8px}
.formula div p{margin:0;font-family:ui-monospace,monospace;font-size:12px;line-height:1.7;word-break:break-word}
section{margin-top:52px}.sec-head{display:flex;align-items:baseline;gap:12px;border-bottom:1px solid var(--line);padding-bottom:12px;margin-bottom:20px}
.sec-head h2{font-size:20px;letter-spacing:-.03em;margin:0;font-weight:600}.sec-head span{font-size:11px;color:var(--muted);font-family:ui-monospace,monospace}
.pillars{display:grid;grid-template-columns:repeat(5,1fr);border:1px solid var(--line);border-radius:8px;overflow:hidden}
.pillar{padding:18px;border-right:1px solid var(--line)}.pillar:last-child{border-right:0}
.pillar .name{font-size:13px;font-weight:600}.pillar .desc{font-size:11px;color:var(--muted);margin:4px 0 14px;min-height:28px}
.pillar .pct{font-family:ui-monospace,monospace;font-size:20px;letter-spacing:-.04em}
.bar{height:2px;background:var(--line);margin:10px 0 8px}.bar i{display:block;height:100%;background:var(--fg)}
.st{font-family:ui-monospace,monospace;font-size:10px;text-transform:uppercase;letter-spacing:.08em}.st.good{color:var(--good)}.st.warn{color:var(--warn)}.st.bad{color:var(--bad)}
table{width:100%;border-collapse:collapse;border:1px solid var(--line);border-radius:8px;overflow:hidden}
th,td{padding:10px 14px;text-align:left;border-bottom:1px solid var(--line);font-size:12px}
th{font-size:10px;text-transform:uppercase;letter-spacing:.1em;color:var(--muted);background:var(--sub);font-weight:500}
td.mono{font-family:ui-monospace,monospace}tr:last-child td{border-bottom:0}
.dot{display:inline-block;width:6px;height:6px;border-radius:50%;margin-right:8px;vertical-align:middle}.dot.ok{background:var(--good)}.dot.warn{background:var(--warn)}.dot.bad{background:var(--bad)}
.flow{display:flex;align-items:stretch;gap:0;border:1px solid var(--line);border-radius:8px;overflow:hidden}
.box{flex:1;padding:18px}.box+.box{border-left:1px solid var(--line)}
.box .lbl{font-size:10px;text-transform:uppercase;letter-spacing:.1em;color:var(--muted)}
.box .num{font-family:ui-monospace,monospace;font-size:18px;margin-top:8px;letter-spacing:-.03em}
.box .list{font-size:11px;color:var(--muted);margin-top:8px;line-height:1.9}
.flow .arrow{display:flex;align-items:center;padding:0 8px;color:var(--muted)}
.note{margin-top:12px;font-size:12px;color:var(--muted)}.note b{color:var(--fg);font-family:ui-monospace,monospace}
.chart{border:1px solid var(--line);border-radius:8px;padding:20px}
ol{border:1px solid var(--line);border-radius:8px;padding:18px 18px 18px 40px;margin:0}li{margin:8px 0;font-family:ui-monospace,monospace;font-size:12px}
footer{border-top:1px solid var(--line);margin-top:64px;padding-top:20px;color:var(--muted);font-size:11px;font-family:ui-monospace,monospace}
@media(max-width:760px){.formula,.pillars{grid-template-columns:1fr}.hero{flex-direction:column;align-items:flex-start}.flow{flex-direction:column}.box+.box{border-left:0;border-top:1px solid var(--line)}}
</style></head><body>
<div class="wrap">
<div class="topbar"><div class="brand"><span class="mark">▲</span> game-ops / 数值主策划工作台</div><div class="meta">{{.Genre}} · {{.Generated}} · exact rational math</div></div>
<div class="eyebrow">Long-term operation / numerical health</div>
<h1>{{.Title}}</h1>
<p class="sub">{{.Subtitle}} · 品类 {{.GenreName}}（<code>{{.Genre}}</code>）</p>

<div class="hero">
 <div class="gauge"><svg width="132" height="132" viewBox="0 0 132 132">
  <circle cx="66" cy="66" r="54" fill="none" stroke="#eaeaea" stroke-width="6"/>
  <circle cx="66" cy="66" r="54" fill="none" stroke="#000" stroke-width="6" stroke-linecap="butt" stroke-dasharray="{{printf "%.2f %.2f" .GaugeFill .GaugeRest}}"/>
 </svg><div class="val"><b class="mono">{{.ScoreText}}</b><small>weighted score</small></div></div>
 <div><span class="badge {{.VerdictClass}}"><span class="dot {{if .Critical}}bad{{else if .Warn}}warn{{else}}ok{{end}}"></span><b>{{.Verdict}}</b></span>
 <h2>{{.OneLiner}}</h2><p class="note">精确值：<b>{{.ScoreExact}}</b> · 0–1 加权分。<code>healthy ≥ 0.8</code>，<code>attention ≥ 0.6</code>，否则 <code>critical</code>。</p></div>
</div>
<div class="formula">
 {{range .MethodLines}}<div><span>method</span><p>{{.}}</p></div>{{end}}
</div>

<section><div class="sec-head"><h2>五大支柱</h2><span>5 pillars / 11 metrics</span></div>
<div class="pillars">
 {{range .Pillars}}<div class="pillar"><div class="name">{{.Name}}</div><div class="desc">{{.Desc}}</div>
 <div class="pct">{{printf "%.1f" .Percent}}<span style="font-size:12px;color:#666">%</span></div>
 <div class="bar"><i style="width:{{printf "%.1f" .Percent}}%"></i></div>
 <div class="st {{if eq .Status "critical"}}bad{{else if eq .Status "warn"}}warn{{else}}good{{end}}">{{.Status}}</div></div>{{end}}
</div></section>

<section><div class="sec-head"><h2>指标明细</h2><span>sorted by fitness ascending</span></div>
<table><tr><th>metric</th><th>含义</th><th>observed</th><th>ideal</th><th>fitness</th><th>status</th></tr>
{{range .Diagnostics}}<tr><td><span class="dot {{if .Critical}}bad{{else if .Warn}}warn{{else}}ok{{end}}"></span><span class="mono">{{.Metric}}</span></td>
<td>{{.Meaning}}</td><td class="mono">{{.Observed}}</td><td class="mono">{{.Ideal}}</td><td class="mono">{{.Fitness}}</td><td class="mono">{{.Status}}</td></tr>{{end}}
</table></section>

<section><div class="sec-head"><h2>价值流动</h2><span>source / sink / base unit</span></div>
<div class="flow">
 <div class="box"><div class="lbl">sources</div><div class="num">{{.Flow.SourceTotal}}</div><div class="list">任务 / 关卡 / 充值<br>精确有理数求和</div></div>
 <div class="arrow">→</div>
 <div class="box"><div class="lbl">ratio</div><div class="num">{{.Flow.Ratio}}</div><div class="list">产出 ÷ 消耗<br>理想值 1</div></div>
 <div class="arrow">→</div>
 <div class="box"><div class="lbl">sinks</div><div class="num">{{.Flow.SinkTotal}}</div><div class="list">升级 / 抽卡 / 门票<br>精确有理数求和</div></div>
</div><p class="note"><b>{{.Flow.Verdict}}</b> — {{.Flow.Note}}</p></section>

<section><div class="sec-head"><h2>成长曲线</h2><span>additive-first</span></div>
<div class="chart"><svg width="740" height="260" viewBox="0 0 740 260" role="img" aria-label="additive growth curve">
 <line x1="50" y1="230" x2="710" y2="230" stroke="#eaeaea"/><line x1="50" y1="40" x2="50" y2="230" stroke="#eaeaea"/>
 <line x1="50" y1="140" x2="710" y2="140" stroke="#eaeaea" stroke-dasharray="2 4"/>
 <polyline points="{{.Curve.AdditivePoints}}" fill="none" stroke="#000" stroke-width="2"/>
 <text x="50" y="252" font-size="11" fill="#666" font-family="monospace">level 1</text>
 <text x="650" y="252" font-size="11" fill="#666" font-family="monospace">level N</text>
 <text x="56" y="32" font-size="11" fill="#000" font-family="monospace">{{.Curve.Title}}: base + step·(n−1), capped</text>
</svg><p class="note">乘法叠加 <code>base·1.1^n</code> 在 L40 约 45×；加法曲线边际增量恒定，可预测、可封顶、可回退。</p></div></section>

<section><div class="sec-head"><h2>风险与调整建议</h2><span>top 3 by severity</span></div>
<ol>{{range .Decisions}}<li>{{.}}</li>{{end}}</ol></section>
<footer>game-ops · 所有数值以 big.Rat 精确计算；字符串配置保留精度 · fitness = clamp(1 − |observed − ideal| ÷ band, 0, 1)</footer>
</div></body></html>`
