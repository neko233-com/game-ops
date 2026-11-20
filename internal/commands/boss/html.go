package bosscmd

import (
	"fmt"
	"html/template"
	"os"
	"strings"

	"github.com/neko233-com/game-ops/boss"
)

const cockpitTpl = `<!doctype html>
<html lang="zh"><head><meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>决策驾驶舱 · {{.GenreName}}</title>
<style>
:root{--fg:#111;--muted:#666;--line:#e5e5e5;--good:#0a7d4f;--warn:#b26a00;--bad:#c0292b;--paper:#fff;--soft:#fafafa}
*{box-sizing:border-box}
body{margin:0;font-family:-apple-system,BlinkMacSystemFont,"Segoe UI","PingFang SC","Microsoft YaHei",sans-serif;color:var(--fg);background:var(--soft);line-height:1.6}
.wrap{max-width:960px;margin:0 auto;padding:32px 24px 64px}
header{display:flex;justify-content:space-between;align-items:flex-end;border-bottom:2px solid var(--fg);padding-bottom:14px}
h1{font-size:22px;margin:0;font-weight:700}
.meta{font-size:12px;color:var(--muted);text-align:right}
.call{display:inline-block;font-size:13px;font-weight:700;padding:4px 12px;border:1px solid;border-radius:4px;margin-right:10px}
.call.GO{color:var(--good);border-color:var(--good)}
.call.HOLD{color:var(--warn);border-color:var(--warn)}
.call.STOP{color:var(--bad);border-color:var(--bad)}
.verdict{font-size:16px;font-weight:600;margin:24px 0 8px}
.score{font-size:13px;color:var(--muted)}
section{background:var(--paper);border:1px solid var(--line);border-radius:8px;padding:20px;margin-top:18px}
h2{font-size:15px;margin:0 0 14px;border-left:3px solid var(--fg);padding-left:10px}
table{width:100%;border-collapse:collapse;font-size:13px}
th,td{text-align:left;padding:9px 8px;border-bottom:1px solid var(--line);vertical-align:top}
th{font-size:11px;text-transform:uppercase;color:var(--muted);letter-spacing:.04em}
.p0{color:var(--bad);font-weight:700}.p1{color:var(--warn);font-weight:700}
.tag{display:inline-block;font-size:11px;padding:1px 7px;border:1px solid var(--line);border-radius:10px;margin-right:6px}
.red{color:var(--bad)}.amber{color:var(--warn)}.green{color:var(--good)}
ol{margin:0;padding-left:20px}li{margin-bottom:8px}
.opts{font-size:12px;color:var(--muted);margin-top:4px}
.foot{font-size:12px;color:var(--muted);margin-top:28px;border-top:1px solid var(--line);padding-top:14px}
@media print{body{background:#fff}.wrap{padding:0}section{break-inside:avoid}}
</style></head>
<body><div class="wrap">
<header><h1>决策驾驶舱 · {{.GenreName}}</h1><div class="meta">健康度 {{.Score}} ｜ {{.Generated}}</div></header>

<div class="verdict"><span class="call {{.Call}}">{{.Call}}</span>{{.Headline}}</div>
<div class="score">判断依据：{{.Confidence}}</div>

{{if .Decisions}}
<section><h2>一、需要你拍板（{{len .Decisions}} 项）</h2>
{{range .Decisions}}
<div style="margin-bottom:16px;padding-bottom:14px;border-bottom:1px dashed var(--line)">
<div><span class="{{lower .Priority}}">{{.Priority}}</span> <strong>{{.Question}}</strong></div>
<div style="font-size:13px;margin-top:4px">{{.Context}}</div>
<div class="opts">证据：{{.Evidence}}</div>
<div class="opts">选项：{{join .Options " ／ "}}</div>
<div style="font-size:13px;margin-top:6px">建议：<strong>{{.Recommend}}</strong></div>
<div class="opts">不做后果：{{.IfNot}}</div>
<div class="opts">负责人：{{.Owner}} ｜ 期限：{{.Deadline}}</div>
</div>
{{end}}
</section>
{{end}}

{{if .Actions}}
<section><h2>二、执行指令</h2>
<table><thead><tr><th>级</th><th>负责人</th><th>任务（可操作）</th><th>工具</th><th>验收</th></tr></thead><tbody>
{{range .Actions}}<tr><td class="{{lower .Priority}}">{{.Priority}}</td><td>{{.Owner}}</td><td>{{.Task}}</td><td>{{.Tool}}</td><td>{{.Verify}}</td></tr>{{end}}
</tbody></table>
</section>
{{end}}

{{if .Findings}}
<section><h2>三、最严重的数值问题</h2>
<table><thead><tr><th>问题（经营语言）</th><th>归属</th><th>当前</th><th>目标</th><th>差距</th><th>后果</th></tr></thead><tbody>
{{range .Findings}}<tr><td><span class="{{sev .Severity}}">●</span> {{.Headline}}</td><td>{{.Question}}</td><td>{{.Observed}}</td><td>{{.Ideal}}</td><td>{{.Gap}}</td><td>{{.Consequence}}</td></tr>{{end}}
</tbody></table>
</section>
{{end}}

<section><h2>四、数值替你做不了的判断（必须人工复核）</h2>
<ul>{{range .Limits}}<li>{{.}}</li>{{end}}</ul>
</section>

{{if .ShowDuty}}
<section><h2>五、主策 + 运营工作事务覆盖</h2>
<table><thead><tr><th>编号</th><th>角色</th><th>领域</th><th>事务</th><th>判定标准</th><th>当前</th></tr></thead><tbody>
{{range .Duties}}<tr><td>{{.ID}}</td><td>{{.Role}}</td><td>{{.Area}}</td><td>{{.Duty}}</td><td>{{.Standard}}</td><td class="{{sev .Status}}">{{.Status}}</td></tr>{{end}}
</tbody></table>
</section>
{{end}}

<div class="foot">本驾驶舱由精确有理数引擎生成；展示值四舍五入，决策以精确值为准。数值是领先指标，不替代对趣味、社交与创意的人工判断。</div>
</div></body></html>`

func writeHTML(path string, b boss.Brief, full bool) {
	funcs := template.FuncMap{
		"lower": strings.ToLower,
		"join":  strings.Join,
		"sev": func(s string) string {
			switch s {
			case "critical", "STOP", "bad":
				return "red"
			case "warn", "attention", "HOLD":
				return "amber"
			case "ok", "GO":
				return "green"
			}
			return ""
		},
	}
	t := template.Must(template.New("cockpit").Funcs(funcs).Parse(cockpitTpl))
	data := struct {
		boss.Brief
		ShowDuty bool
	}{b, full}
	f, err := os.Create(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	defer f.Close()
	if err := t.Execute(f, data); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	fmt.Fprintf(os.Stderr, "wrote cockpit %s\n", path)
}
