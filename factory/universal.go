// Package factory — universal fallback strategy.
//
// When a genre has no hand-tuned Strategy, GetAny returns a UniversalStrategy
// driven by the genre catalog's levers. This makes factory.RunAny work for ANY
// category — including ones invented at runtime — instead of erroring out.
package factory

import (
	"fmt"
	"strings"

	"github.com/neko233-com/game-ops/genre"
	"github.com/neko233-com/game-ops/num"
)

// UniversalStrategy is a data-driven Strategy backed by genre levers.
// It is intentionally generic: genre-specific nuance lives in the 29 generated
// strategies, while this guarantees no category is ever unsupported.
type UniversalStrategy struct {
	id string
}

// NewUniversal builds a universal strategy for a genre ID.
func NewUniversal(id string) *UniversalStrategy { return &UniversalStrategy{id: id} }

func (u *UniversalStrategy) ID() string { return u.id }
func (u *UniversalStrategy) Name() string {
	g := genre.GetAny(u.id)
	return g.Name + " / 通用策略 (universal)"
}

func (u *UniversalStrategy) GenreMetrics() []MetricLabel {
	g := genre.GetAny(u.id)
	out := make([]MetricLabel, 0, len(g.Levers))
	for _, l := range g.Levers {
		meta, ok := genre.Metrics()[l.Key]
		label, desc := l.Key, "genre lever"
		if ok {
			label, desc = meta.Label, meta.Desc
		}
		out = append(out, MetricLabel{Key: l.Key, Label: label, Desc: desc})
	}
	return out
}

func (u *UniversalStrategy) Advice() []string {
	return []string{
		"通用策略：按 11 项通用指标的品类 lever 判定健康度",
		"用 ops.EvaluateAny 评估；用 genre derive/register 自定义 lever 后可获得专属判定",
		"数值必须加法优先、精确有理数；先归一化到基准单位再比较",
	}
}

func (u *UniversalStrategy) Evaluate(ctx StrategyContext) (StrategyResult, error) {
	g := genre.GetAny(ctx.GenreID)
	res := StrategyResult{
		Genre:  g.ID,
		Region: ctx.Region,
		Notes:  fmt.Sprintf("%s 使用通用策略：按品类 lever 检查 %d 项指标。", g.Name, len(g.Levers)),
	}
	if len(g.Tags) > 0 {
		res.Notes += " 标签：" + strings.Join(g.Tags, ", ") + "。"
	}

	for _, ml := range u.GenreMetrics() {
		if v, ok := ctx.Metrics[ml.Key]; ok {
			res.GenreMetrics = append(res.GenreMetrics, StrategyMetric{
				Key: ml.Key, Label: ml.Label, Value: v,
			})
		}
	}

	for _, l := range g.Levers {
		v, ok := ctx.Metrics[l.Key]
		if !ok {
			continue
		}
		if v.Cmp(l.Min) < 0 {
			res.Risks = append(res.Risks, RiskItem{
				Metric: l.Key, Observed: v, Ideal: l.Ideal, Severity: "critical",
				Message: fmt.Sprintf("%s 低于品类下限 %s (observed %s)", l.Key, l.Min.String(), v.DecimalString(6)),
			})
		} else if v.Cmp(l.Max) > 0 {
			res.Risks = append(res.Risks, RiskItem{
				Metric: l.Key, Observed: v, Ideal: l.Ideal, Severity: "critical",
				Message: fmt.Sprintf("%s 高于品类上限 %s (observed %s)", l.Key, l.Max.String(), v.DecimalString(6)),
			})
		}
	}
	return res, nil
}

// GetAny returns a registered Strategy, or a UniversalStrategy for the ID.
func GetAny(id string) Strategy {
	if s, ok := Get(id); ok {
		return s
	}
	return NewUniversal(id)
}

// RunAny dispatches to a registered strategy, falling back to universal.
func RunAny(ctx StrategyContext) (StrategyResult, error) {
	return GetAny(ctx.GenreID).Evaluate(ctx)
}

var _ Strategy = (*UniversalStrategy)(nil)

// Keep num imported for downstream spec helpers compiled in this package.
var _ = num.One
