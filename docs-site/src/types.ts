/**
 * 内容块的数据模型。
 *
 * 所有页面正文都保存在 src/content/*.ts 里，App 只负责渲染这些块。
 * 这样新增一页只需要写一个数据文件，不需要改组件。
 */

export type Block =
  | { type: "p"; text: string }
  | { type: "h2"; text: string }
  | { type: "h3"; text: string }
  | { type: "ul"; items: string[] }
  | { type: "ol"; items: string[] }
  | { type: "code"; lang: string; source: string; caption?: string }
  | { type: "table"; head: string[]; rows: string[][]; caption?: string }
  | { type: "callout"; tone: Tone; title: string; text: string }
  | { type: "formula"; expression: string; note?: string }
  | { type: "steps"; items: Array<{ title: string; text: string }> }
  | { type: "quote"; text: string; attribution?: string }
  | { type: "flow"; diagram: FlowDiagram }
  | { type: "echarts"; chart: ChartId }
  /**
   * 翻译占位块：正文尚未翻译时，用目标语言明确标记，
   * 然后完整展示中文原文。避免出现半翻译的句子。
   */
  | { type: "untranslated"; original: Block };

export type Tone = "note" | "tip" | "warn" | "danger";

/** 交互式图解的类型：由 src/flow/FlowDiagrams.tsx 渲染。 */
export type FlowDiagram = "value-flow" | "system-map";

/**
 * 专业图表的稳定 id：由 src/charts/ProCharts.tsx 渲染。
 * 每个 id 对应一份精确的 ECharts option（数值全部来自引擎 / 配表常量）。
 */
export type ChartId =
  | "growth-curves"
  | "capacity-days"
  | "economy-sankey"
  | "sensitivity-heatmap"
  | "config-anomaly-scatter";

/**
 * 一页的可选译文。未提供的字段回退到中文原文。
 * `blocks` 缺省时，localizePage 会把所有原文块包成 `untranslated` 块。
 */
export interface PageTranslation {
  nav?: string;
  title?: string;
  subtitle?: string;
  blocks?: Block[];
}

export interface DocPage {
  /** URL hash 用的稳定 id，例如 #/overview */
  id: string;
  /** 侧边栏显示的短标题 */
  nav: string;
  /** 正文大标题 */
  title: string;
  /** 标题下方的一句话摘要 */
  subtitle: string;
  /** 侧边栏分组名 */
  section: string;
  /** 预估阅读时间，仅用于展示 */
  readingTime: string;
  blocks: Block[];
  /** 多语言译文。中文 (zh-CN) 始终是权威完整版本，不在这里重复。 */
  i18n?: {
    en?: PageTranslation;
    ja?: PageTranslation;
  };
}

export const p = (text: string): Block => ({ type: "p", text });
export const h2 = (text: string): Block => ({ type: "h2", text });
export const h3 = (text: string): Block => ({ type: "h3", text });
export const ul = (...items: string[]): Block => ({ type: "ul", items });
export const ol = (...items: string[]): Block => ({ type: "ol", items });
export const code = (lang: string, source: string, caption?: string): Block => ({
  type: "code",
  lang,
  source,
  caption,
});
export const table = (head: string[], rows: string[][], caption?: string): Block => ({
  type: "table",
  head,
  rows,
  caption,
});
export const callout = (tone: Tone, title: string, text: string): Block => ({
  type: "callout",
  tone,
  title,
  text,
});
export const formula = (expression: string, note?: string): Block => ({
  type: "formula",
  expression,
  note,
});
export const steps = (...items: Array<{ title: string; text: string }>): Block => ({
  type: "steps",
  items,
});
export const quote = (text: string, attribution?: string): Block => ({
  type: "quote",
  text,
  attribution,
});
export const flow = (diagram: FlowDiagram): Block => ({ type: "flow", diagram });
export const echart = (chart: ChartId): Block => ({ type: "echarts", chart });
