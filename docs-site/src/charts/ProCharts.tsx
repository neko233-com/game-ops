/**
 * 面向数值主策划的专业图表（Apache ECharts 6）。
 *
 * 只用 echarts 核心 API：useRef + useEffect + echarts.init，ResizeObserver 自适应，
 * 卸载时 dispose。不引入 echarts-for-react 之类的包装库。
 *
 * 所有数值都来自主仓库引擎常量 / 配表常量，展示层只做确定性的四则运算，
 * 引擎里的真实计算使用精确有理数 num.Num（见 prose 说明）。
 */

import { useEffect, useMemo, useRef, useState } from "react";
import * as echarts from "echarts/core";
import {
  BarChart,
  CustomChart,
  HeatmapChart,
  LineChart,
  SankeyChart,
  ScatterChart,
} from "echarts/charts";
import {
  DataZoomComponent,
  GridComponent,
  LegendComponent,
  MarkLineComponent,
  TitleComponent,
  TooltipComponent,
  VisualMapComponent,
} from "echarts/components";
import { CanvasRenderer } from "echarts/renderers";
import type { ECharts, EChartsCoreOption } from "echarts/core";
import type { ChartId } from "../types";

echarts.use([
  BarChart,
  CustomChart,
  HeatmapChart,
  LineChart,
  SankeyChart,
  ScatterChart,
  DataZoomComponent,
  GridComponent,
  LegendComponent,
  MarkLineComponent,
  TitleComponent,
  TooltipComponent,
  VisualMapComponent,
  CanvasRenderer,
]);

/* ------------------------------------------------------------------ */
/* 主题与共用外壳                                                      */
/* ------------------------------------------------------------------ */

/** 专业数据分析配色：高对比、色觉友好、印刷友好。 */
export const PROFESSIONAL_COLORS = {
  additive: "#1f6feb",
  exponential: "#d97706",
  teal: "#0f766e",
  violet: "#7c3aed",
  amber: "#b45309",
  rose: "#be123c",
  ink: "#1e293b",
  grid: "#e2e8f0",
  axis: "#64748b",
};

const SERIES_PALETTE = [
  PROFESSIONAL_COLORS.additive,
  PROFESSIONAL_COLORS.exponential,
  PROFESSIONAL_COLORS.teal,
  PROFESSIONAL_COLORS.violet,
  PROFESSIONAL_COLORS.amber,
  PROFESSIONAL_COLORS.rose,
];

const CHART_FONT =
  '"PingFang SC", "Microsoft YaHei", "Noto Sans CJK SC", -apple-system, BlinkMacSystemFont, sans-serif';

const integerFormat = new Intl.NumberFormat("zh-CN", { maximumFractionDigits: 0 });
const oneDecimalFormat = new Intl.NumberFormat("zh-CN", {
  minimumFractionDigits: 1,
  maximumFractionDigits: 1,
});

function formatInt(value: number): string {
  return integerFormat.format(value);
}

function formatOne(value: number): string {
  return oneDecimalFormat.format(value);
}

/** 通用 tooltip / 网格 / 图例设置，保证五张图视觉语言一致。 */
function baseOption(): EChartsCoreOption {
  return {
    color: SERIES_PALETTE,
    backgroundColor: "transparent",
    textStyle: { fontFamily: CHART_FONT, color: PROFESSIONAL_COLORS.ink },
    animationDuration: 450,
  };
}

/** 可复用的 ECharts React 组件：init → setOption → ResizeObserver → dispose。 */
export function EChart({
  option,
  height = 380,
  className,
}: {
  option: EChartsCoreOption;
  height?: number;
  className?: string;
}) {
  const hostRef = useRef<HTMLDivElement | null>(null);
  const chartRef = useRef<ECharts | null>(null);

  useEffect(() => {
    const host = hostRef.current;
    if (!host) {
      return;
    }

    const chart = echarts.init(host, undefined, { renderer: "canvas" });
    chartRef.current = chart;

    const observer = new ResizeObserver(() => {
      chart.resize();
    });
    observer.observe(host);

    return () => {
      observer.disconnect();
      chart.dispose();
      chartRef.current = null;
    };
  }, []);

  useEffect(() => {
    chartRef.current?.setOption(option, true);
  }, [option]);

  return (
    <div
      ref={hostRef}
      className={className ?? "echart-canvas"}
      style={{ width: "100%", height }}
      role="img"
      aria-label="ECharts 数值分析图表"
    />
  );
}

/* ------------------------------------------------------------------ */
/* 图 1：成长曲线对比（加法 vs 指数）                                  */
/* ------------------------------------------------------------------ */

const GROWTH_BASE = 100;
const GROWTH_STEP = 25;
const GROWTH_CAP = 10000;
const GROWTH_RATE = 1.1;
const MAX_LEVEL = 60;

function additiveAt(level: number): number {
  return Math.min(GROWTH_BASE + GROWTH_STEP * (level - 1), GROWTH_CAP);
}

function exponentialAt(level: number): number {
  return GROWTH_BASE * Math.pow(GROWTH_RATE, level - 1);
}

const growthLevels = Array.from({ length: MAX_LEVEL }, (_, i) => i + 1);

function buildGrowthOption(logScale: boolean): EChartsCoreOption {
  const l40Add = additiveAt(40);
  const l40Exp = exponentialAt(40);

  return {
    ...baseOption(),
    title: {
      text: "成长曲线对比：加法优先 vs 指数复利",
      subtext: "加法 100+25(n−1)（封顶 10000） · 指数 100×1.1^(n−1) · 1.1^40 ≈ 45.26×",
      left: 12,
      textStyle: { fontSize: 15 },
      subtextStyle: { fontSize: 11, color: PROFESSIONAL_COLORS.axis },
    },
    tooltip: {
      trigger: "axis",
      axisPointer: { type: "cross", label: { backgroundColor: PROFESSIONAL_COLORS.ink } },
      valueFormatter: (value: unknown) =>
        typeof value === "number" ? formatInt(value) : String(value),
    },
    legend: {
      top: 58,
      data: ["加法优先 (封顶)", "指数复利 1.1^(n−1)"],
      textStyle: { fontSize: 12 },
    },
    toolbox: {
      right: 12,
      top: 8,
      feature: {
        dataZoom: { title: { zoom: "框选缩放", back: "还原缩放" } },
        restore: { title: "重置" },
        saveAsImage: { title: "导出 PNG" },
      },
    },
    grid: { left: 64, right: 64, top: 96, bottom: 72, containLabel: true },
    xAxis: {
      type: "category",
      name: "等级 L",
      nameLocation: "middle",
      nameGap: 26,
      data: growthLevels,
      boundaryGap: false,
      axisLabel: { interval: 4, color: PROFESSIONAL_COLORS.axis },
      axisLine: { lineStyle: { color: PROFESSIONAL_COLORS.grid } },
    },
    yAxis: {
      type: logScale ? "log" : "value",
      name: logScale ? "数值（对数轴）" : "数值（线性轴）",
      min: logScale ? 100 : 0,
      max: logScale ? 40000 : undefined,
      nameTextStyle: { fontSize: 11 },
      axisLabel: {
        color: PROFESSIONAL_COLORS.axis,
        formatter: (value: number) =>
          value >= 1000 ? `${formatInt(value / 1000)}k` : formatInt(value),
      },
      splitLine: { lineStyle: { color: PROFESSIONAL_COLORS.grid, type: "dashed" } },
    },
    dataZoom: [
      { type: "slider", start: 0, end: 100, height: 22, bottom: 18, borderColor: "#cbd5e1" },
      { type: "inside", start: 0, end: 100 },
    ],
    series: [
      {
        name: "加法优先 (封顶)",
        type: "line",
        symbol: "none",
        lineStyle: { width: 3, color: PROFESSIONAL_COLORS.additive },
        itemStyle: { color: PROFESSIONAL_COLORS.additive },
        data: growthLevels.map(additiveAt),
      },
      {
        name: "指数复利 1.1^(n−1)",
        type: "line",
        symbol: "none",
        lineStyle: { width: 3, color: PROFESSIONAL_COLORS.exponential },
        itemStyle: { color: PROFESSIONAL_COLORS.exponential },
        data: growthLevels.map(exponentialAt),
        markLine: {
          symbol: "none",
          silent: true,
          lineStyle: { color: PROFESSIONAL_COLORS.exponential, type: "dashed", width: 1 },
          label: {
            formatter: `L40: 加法 ${formatInt(l40Add)} / 乘法 ≈${formatInt(l40Exp)}`,
            position: "insideEndTop",
            fontSize: 11,
          },
          data: [{ xAxis: 40 }],
        },
        markPoint: {
          symbol: "circle",
          symbolSize: 10,
          itemStyle: { color: PROFESSIONAL_COLORS.exponential },
          label: { show: false },
          data: [{ name: "L40 指数点", coord: [40, Number(l40Exp.toFixed(2))] }],
        },
      },
    ],
  };
}

/** 成长图带一个 React 端的线性 / 对数切换按钮。 */
function GrowthCurvesBlock() {
  const [logScale, setLogScale] = useState(false);
  const option = useMemo(() => buildGrowthOption(logScale), [logScale]);

  return (
    <div className="echart-card">
      <div className="echart-toolbar">
        <span className="echart-toolbar-label">纵轴刻度</span>
        <div className="echart-segment">
          <button
            type="button"
            className={!logScale ? "echart-seg-on" : ""}
            onClick={() => setLogScale(false)}
          >
            线性
          </button>
          <button
            type="button"
            className={logScale ? "echart-seg-on" : ""}
            onClick={() => setLogScale(true)}
          >
            对数
          </button>
        </div>
        <span className="echart-toolbar-note">
          拖动下方滑块查看任意等级段；L40 处指数约为加法的{" "}
          {(exponentialAt(40) / additiveAt(40)).toFixed(2)} 倍
        </span>
      </div>
      <EChart option={option} height={430} />
    </div>
  );
}

/* ------------------------------------------------------------------ */
/* 图 2：满级时间分布（分位数柱 + 留存修正到达率）                     */
/* ------------------------------------------------------------------ */

const TOTAL_XP = 40120;

type Percentile = "P10" | "P25" | "P50" | "P75" | "P90";

const PERCENTILES: Percentile[] = ["P10", "P25", "P50", "P75", "P90"];

/** 每个分位对应的“日经验效率系数”：快 cohorts 拿更高分位（天数更短）。 */
const DISPERSION: Record<Percentile, number> = {
  P10: 1.25,
  P25: 1.1,
  P50: 1,
  P75: 0.9,
  P90: 0.75,
};

type Archetype = {
  name: string;
  dailyXP: number;
  weight: number;
  d1: number;
  d7: number;
  d30: number;
};

/** 日经验来自设计假设（常规 900 / 轻度 480 / 重度 1500）；权重为人口占比。 */
const ARCHETYPES: Archetype[] = [
  { name: "常规", dailyXP: 900, weight: 0.5, d1: 0.55, d7: 0.35, d30: 0.18 },
  { name: "轻度", dailyXP: 480, weight: 0.3, d1: 0.4, d7: 0.2, d30: 0.09 },
  { name: "重度", dailyXP: 1500, weight: 0.2, d1: 0.7, d7: 0.5, d30: 0.3 },
];

function daysAtPercentile(dailyXP: number, percentile: Percentile): number {
  return TOTAL_XP / (dailyXP * DISPERSION[percentile]);
}

/** 与引擎 progression.SurvivalAt 相同的分段线性留存插值。 */
function survivalAt(a: Archetype, day: number): number {
  const anchors: Array<[number, number]> = [
    [0, 1],
    [1, a.d1],
    [7, a.d7],
    [30, a.d30],
  ];
  if (day > 30) {
    const slope = (a.d30 - a.d7) / (30 - 7);
    return Math.max(0, Math.min(1, a.d30 + (day - 30) * slope));
  }
  for (let i = 1; i < anchors.length; i++) {
    const [loDay, loVal] = anchors[i - 1];
    const [hiDay, hiVal] = anchors[i];
    if (day <= hiDay) {
      const value = loVal + ((day - loDay) / (hiDay - loDay)) * (hiVal - loVal);
      return Math.max(0, Math.min(1, value));
    }
  }
  return 1;
}

function buildCapacityOption(): EChartsCoreOption {
  const reach = ARCHETYPES.map((a) =>
    Number((survivalAt(a, daysAtPercentile(a.dailyXP, "P50")) * 100).toFixed(2)),
  );

  return {
    ...baseOption(),
    title: {
      text: "满级时间分布与留存修正到达率",
      subtext: `总经验 ${formatInt(TOTAL_XP)} · 日经验：常规 900 / 轻度 480 / 重度 1500 · 柱=天数，线=仍活跃占比`,
      left: 12,
      textStyle: { fontSize: 15 },
      subtextStyle: { fontSize: 11, color: PROFESSIONAL_COLORS.axis },
    },
    tooltip: {
      trigger: "axis",
      axisPointer: { type: "shadow" },
      valueFormatter: (value: unknown) =>
        typeof value === "number" ? formatOne(value) : String(value),
    },
    legend: { top: 58, textStyle: { fontSize: 12 } },
    toolbox: {
      right: 12,
      top: 8,
      feature: { restore: { title: "重置" }, saveAsImage: { title: "导出 PNG" } },
    },
    grid: { left: 56, right: 64, top: 96, bottom: 52, containLabel: true },
    xAxis: {
      type: "category",
      data: ARCHETYPES.map((a) => a.name),
      axisLabel: { fontSize: 13, color: PROFESSIONAL_COLORS.ink },
      axisLine: { lineStyle: { color: PROFESSIONAL_COLORS.grid } },
    },
    yAxis: [
      {
        type: "value",
        name: "到达满级天数",
        nameTextStyle: { fontSize: 11 },
        splitLine: { lineStyle: { color: PROFESSIONAL_COLORS.grid, type: "dashed" } },
      },
      {
        type: "value",
        name: "留存到达率 %",
        min: 0,
        max: 100,
        nameTextStyle: { fontSize: 11 },
        splitLine: { show: false },
        axisLabel: { formatter: "{value}%", color: PROFESSIONAL_COLORS.rose },
      },
    ],
    series: [
      ...PERCENTILES.map((percentile) => ({
        name: percentile,
        type: "bar" as const,
        barMaxWidth: 26,
        itemStyle: { borderRadius: [3, 3, 0, 0] },
        label: {
          show: true,
          position: "top" as const,
          formatter: "{c}",
          fontSize: 10,
          color: PROFESSIONAL_COLORS.axis,
        },
        data: ARCHETYPES.map((a) => Number(daysAtPercentile(a.dailyXP, percentile).toFixed(2))),
      })),
      {
        name: "P50 留存到达率",
        type: "line",
        yAxisIndex: 1,
        symbol: "circle",
        symbolSize: 10,
        lineStyle: { width: 3, color: PROFESSIONAL_COLORS.rose },
        itemStyle: { color: PROFESSIONAL_COLORS.rose },
        label: { show: true, formatter: "{c}%", fontSize: 11, color: PROFESSIONAL_COLORS.rose },
        data: reach,
      },
    ],
  };
}

/* ------------------------------------------------------------------ */
/* 图 3：经济流向 Sankey                                               */
/* ------------------------------------------------------------------ */

function buildSankeyOption(): EChartsCoreOption {
  return {
    ...baseOption(),
    title: {
      text: "每日经济流向（Sankey）",
      subtext: "钻石按 1:10 折成金币 · 进 180 = 出 180，比值 1.0",
      left: 12,
      textStyle: { fontSize: 15 },
      subtextStyle: { fontSize: 11, color: PROFESSIONAL_COLORS.axis },
    },
    tooltip: {
      trigger: "item",
      triggerOn: "mousemove",
      formatter: (params: unknown) => {
        const item = params as {
          name?: string;
          value?: number;
          data?: { source?: string; target?: string; value?: number };
        };
        if (item.data?.source) {
          return `${item.data.source} → ${item.data.target}<br/><b>${formatInt(Number(item.data.value))} 金币/日</b>`;
        }
        return `${item.name ?? ""}<br/>合计 <b>${formatInt(Number(item.value))} 金币/日</b>`;
      },
    },
    series: [
      {
        type: "sankey",
        left: 24,
        right: 132,
        top: 96,
        bottom: 24,
        nodeWidth: 18,
        nodeGap: 18,
        draggable: false,
        emphasis: { focus: "adjacency" },
        data: [
          { name: "每日任务", itemStyle: { color: PROFESSIONAL_COLORS.teal } },
          { name: "主线关卡", itemStyle: { color: PROFESSIONAL_COLORS.teal } },
          { name: "充值 3 钻石", itemStyle: { color: PROFESSIONAL_COLORS.teal } },
          { name: "玩家钱包", itemStyle: { color: PROFESSIONAL_COLORS.additive } },
          { name: "升级", itemStyle: { color: PROFESSIONAL_COLORS.amber } },
          { name: "抽卡", itemStyle: { color: PROFESSIONAL_COLORS.amber } },
          { name: "入场券", itemStyle: { color: PROFESSIONAL_COLORS.amber } },
        ],
        links: [
          { source: "每日任务", target: "玩家钱包", value: 100 },
          { source: "主线关卡", target: "玩家钱包", value: 50 },
          { source: "充值 3 钻石", target: "玩家钱包", value: 30 },
          { source: "玩家钱包", target: "升级", value: 100 },
          { source: "玩家钱包", target: "抽卡", value: 50 },
          { source: "玩家钱包", target: "入场券", value: 30 },
        ],
        label: {
          formatter: "{b}  {c}",
          fontSize: 12,
          color: PROFESSIONAL_COLORS.ink,
        },
        lineStyle: { color: "gradient", opacity: 0.42, curveness: 0.5 },
      },
    ],
  };
}

/* ------------------------------------------------------------------ */
/* 图 4：数值敏感度热力图（pay_rate × retention_d1）                   */
/* ------------------------------------------------------------------ */

/** gacha 品类的两条 lever（与 genre/genre.go 一致）。 */
const PAY = { ideal: 0.06, min: 0.02, max: 0.25, weight: 1.3 };
const RET = { ideal: 0.4, min: 0.15, max: 0.7, weight: 1.0 };

const PAY_VALUES = [0.01, 0.02, 0.04, 0.06, 0.08, 0.1, 0.13, 0.16, 0.2, 0.25];
const RET_VALUES = [0.1, 0.15, 0.25, 0.4, 0.55, 0.7];

function fitness(observed: number, ideal: number, min: number, max: number): number {
  const band = (max - min) / 2;
  return Math.max(0, Math.min(1, 1 - Math.abs(observed - ideal) / band));
}

function healthScore(payRate: number, retentionD1: number): number {
  const fPay = fitness(payRate, PAY.ideal, PAY.min, PAY.max);
  const fRet = fitness(retentionD1, RET.ideal, RET.min, RET.max);
  return (fPay * PAY.weight + fRet * RET.weight) / (PAY.weight + RET.weight);
}

function pct(value: number): string {
  return `${(value * 100).toFixed(value < 0.1 ? 1 : 0)}%`;
}

function buildHeatmapOption(): EChartsCoreOption {
  const cells: Array<[number, number, number]> = [];
  for (let xi = 0; xi < PAY_VALUES.length; xi++) {
    for (let yi = 0; yi < RET_VALUES.length; yi++) {
      cells.push([xi, yi, Number(healthScore(PAY_VALUES[xi], RET_VALUES[yi]).toFixed(4))]);
    }
  }

  const idealX = PAY_VALUES.indexOf(PAY.ideal);
  const idealY = RET_VALUES.indexOf(RET.ideal);

  return {
    ...baseOption(),
    title: {
      text: "健康度敏感度：付费率 × 次日留存",
      subtext:
        "gacha lever · fitness=clamp(1−|obs−ideal|/band,0,1) · 加权均值(付费 1.3 : 留存 1.0)",
      left: 12,
      textStyle: { fontSize: 15 },
      subtextStyle: { fontSize: 11, color: PROFESSIONAL_COLORS.axis },
    },
    tooltip: {
      position: "top",
      formatter: (params: unknown) => {
        const item = params as { value?: number[] };
        const [xi, yi, score] = item.value ?? [0, 0, 0];
        const pay = PAY_VALUES[xi];
        const ret = RET_VALUES[yi];
        return [
          `付费率 ${pct(pay)} · D1 留存 ${pct(ret)}`,
          `健康度 <b>${Number(score).toFixed(4)}</b>`,
          `付费 fitness ${fitness(pay, PAY.ideal, PAY.min, PAY.max).toFixed(4)}`,
          `留存 fitness ${fitness(ret, RET.ideal, RET.min, RET.max).toFixed(4)}`,
        ].join("<br/>");
      },
    },
    grid: { left: 88, right: 28, top: 92, bottom: 72, containLabel: true },
    xAxis: {
      type: "category",
      name: "付费率 monetization_pay_rate",
      nameLocation: "middle",
      nameGap: 30,
      data: PAY_VALUES.map(pct),
      splitArea: { show: true },
      axisLabel: { fontSize: 10, rotate: 30 },
    },
    yAxis: {
      type: "category",
      name: "D1 留存",
      data: RET_VALUES.map(pct),
      splitArea: { show: true },
      axisLabel: { fontSize: 11 },
    },
    visualMap: {
      min: 0,
      max: 1,
      calculable: true,
      orient: "horizontal",
      left: "center",
      bottom: 6,
      itemWidth: 14,
      itemHeight: 140,
      precision: 2,
      inRange: {
        color: ["#d73027", "#fc8d59", "#fee08b", "#a6d96a", "#1a9850"],
      },
      text: ["健康 1.00", "危险 0.00"],
      textStyle: { fontSize: 11 },
    },
    series: [
      {
        name: "健康度",
        type: "heatmap",
        data: cells,
        label: { show: true, formatter: "{c}", fontSize: 9, color: "#0f172a" },
        emphasis: { itemStyle: { shadowBlur: 8, shadowColor: "rgba(0,0,0,0.3)" } },
      },
      {
        name: "gacha 理想点",
        type: "scatter",
        data: [[idealX, idealY]],
        symbolSize: 18,
        itemStyle: { color: "#ffffff", borderColor: PROFESSIONAL_COLORS.ink, borderWidth: 2 },
        label: {
          show: true,
          formatter: "理想点",
          position: "top",
          fontSize: 11,
          fontWeight: "bold",
        },
        tooltip: { show: false },
      },
    ],
  };
}

/* ------------------------------------------------------------------ */
/* 图 5：配表异常散点（max/min 倍数，阈值 10×）                        */
/* ------------------------------------------------------------------ */

const ANOMALY_COLUMNS: Array<{ name: string; min: number; max: number }> = [
  { name: "gold 金币", min: 1, max: 1000000 },
  { name: "drop_rate 掉率", min: 0.0005, max: 0.025 },
  { name: "price 价格档", min: 6, max: 648 },
  { name: "exp 经验", min: 100, max: 12500 },
  { name: "crit 暴击率", min: 0.01, max: 0.2 },
  { name: "attack 攻击", min: 100, max: 1200 },
  { name: "hp 生命", min: 50, max: 1850 },
  { name: "reward 奖励", min: 20, max: 120 },
  { name: "damage 伤害", min: 5, max: 25 },
  { name: "move_speed 移速", min: 2, max: 3 },
];

function buildAnomalyOption(): EChartsCoreOption {
  const points = ANOMALY_COLUMNS.map((column) => {
    const ratio = Number((column.max / column.min).toFixed(2));
    return {
      name: column.name,
      value: [column.name, ratio] as [string, number],
      min: column.min,
      max: column.max,
      ratio,
    };
  });

  return {
    ...baseOption(),
    title: {
      text: "配表异常扫描：各列 max / min 倍数",
      subtext: "纵轴为对数刻度 · 虚线为 10× 审查阈值 · 精确值见 tooltip",
      left: 12,
      textStyle: { fontSize: 15 },
      subtextStyle: { fontSize: 11, color: PROFESSIONAL_COLORS.axis },
    },
    tooltip: {
      trigger: "item",
      formatter: (params: unknown) => {
        const item = params as {
          data?: { name?: string; ratio?: number; min?: number; max?: number };
        };
        const data = item.data;
        if (!data) {
          return "";
        }
        const ratio = data.ratio ?? 0;
        return [
          `<b>${data.name ?? ""}</b>`,
          `min ${data.min} · max ${data.max}`,
          `倍数 <b>${formatInt(ratio)}×</b>`,
          ratio > 10 ? "⚠ 超过 10× 阈值：需人工复核" : "✓ 在阈值内",
        ].join("<br/>");
      },
    },
    grid: { left: 76, right: 32, top: 92, bottom: 88, containLabel: true },
    xAxis: {
      type: "category",
      data: ANOMALY_COLUMNS.map((column) => column.name),
      axisLabel: { rotate: 28, fontSize: 10, color: PROFESSIONAL_COLORS.axis },
      axisLine: { lineStyle: { color: PROFESSIONAL_COLORS.grid } },
    },
    yAxis: {
      type: "log",
      name: "max / min（对数）",
      min: 1,
      max: 2000000,
      nameTextStyle: { fontSize: 11 },
      axisLabel: {
        formatter: (value: number) => `${formatInt(value)}×`,
        color: PROFESSIONAL_COLORS.axis,
      },
      splitLine: { lineStyle: { color: PROFESSIONAL_COLORS.grid, type: "dashed" } },
    },
    series: [
      {
        name: "倍数",
        type: "scatter",
        symbolSize: 13,
        data: points,
        itemStyle: {
          color: (params: { dataIndex: number }) =>
            points[params.dataIndex].ratio > 10
              ? PROFESSIONAL_COLORS.rose
              : PROFESSIONAL_COLORS.teal,
        },
        label: {
          show: true,
          position: "top",
          formatter: "{c}×",
          fontSize: 9,
          color: PROFESSIONAL_COLORS.axis,
        },
        markLine: {
          symbol: "none",
          silent: true,
          lineStyle: { color: PROFESSIONAL_COLORS.amber, type: "dashed", width: 2 },
          label: { formatter: "审查阈值 10×", position: "insideEndTop", fontSize: 11 },
          data: [{ yAxis: 10 }],
        },
      },
    ],
  };
}

/* ------------------------------------------------------------------ */
/* 注册与分发                                                          */
/* ------------------------------------------------------------------ */

const chartOptions: Record<Exclude<ChartId, "growth-curves">, EChartsCoreOption> = {
  "capacity-days": buildCapacityOption(),
  "economy-sankey": buildSankeyOption(),
  "sensitivity-heatmap": buildHeatmapOption(),
  "config-anomaly-scatter": buildAnomalyOption(),
};

export function EChartsBlock({ chart }: { chart: ChartId }) {
  if (chart === "growth-curves") {
    return <GrowthCurvesBlock />;
  }

  return <EChart option={chartOptions[chart]} />;
}
