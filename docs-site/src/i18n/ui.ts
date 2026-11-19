/**
 * 界面文案（sidebar / 顶栏 / 翻页 / 提示框徽标等“UI chrome”）。
 *
 * 正文内容不在这里：正文走 src/content 的 DocPage.i18n。
 * 中文是权威版本；en / ja 只负责界面词汇。
 */

import type { Lang } from "./index";

export interface UiMessages {
  brandTitle: string;
  brandSubtitle: string;
  topbarTag: string;
  menuToggleLabel: string;
  closeSidebarLabel: string;
  sidebarFoot: string;
  tocTitle: string;
  prevPage: string;
  nextPage: string;
  pageNavLabel: string;
  copy: string;
  copied: string;
  chartLoading: string;
  languageSwitcherLabel: string;
  langZh: string;
  langEn: string;
  langJa: string;
  originalBadge: string;
  originalBadgeNote: string;
  untranslatedNoticeTitle: string;
  untranslatedNotice: string;
  toneNote: string;
  toneTip: string;
  toneWarn: string;
  toneDanger: string;
}

export type UiKey = keyof UiMessages;

export const messages: Record<Lang, UiMessages> = {
  "zh-CN": {
    brandTitle: "大道归一",
    brandSubtitle: "游戏数值与长线运营手册",
    topbarTag: "companion docs",
    menuToggleLabel: "切换侧边栏",
    closeSidebarLabel: "关闭侧边栏",
    sidebarFoot: "数值是体检报告，不是命运判词。",
    tocTitle: "本页目录",
    prevPage: "← 上一篇",
    nextPage: "下一篇 →",
    pageNavLabel: "页面导航",
    copy: "复制",
    copied: "已复制",
    chartLoading: "加载图表…",
    languageSwitcherLabel: "语言",
    langZh: "中文",
    langEn: "EN",
    langJa: "日本語",
    originalBadge: "原文为中文",
    originalBadgeNote: "本块尚未翻译，下面展示完整中文原文。",
    untranslatedNoticeTitle: "翻译状态",
    untranslatedNotice: "本页正文尚未翻译，以下展示完整中文原文；每个原文块都带有标记。",
    toneNote: "说明",
    toneTip: "建议",
    toneWarn: "注意",
    toneDanger: "重要",
  },
  en: {
    brandTitle: "Da Dao Gui Yi",
    brandSubtitle: "Game Numbers & Live-Ops Manual",
    topbarTag: "companion docs",
    menuToggleLabel: "Toggle sidebar",
    closeSidebarLabel: "Close sidebar",
    sidebarFoot: "Numbers are a check-up report, not a verdict on your fate.",
    tocTitle: "On this page",
    prevPage: "← Previous",
    nextPage: "Next →",
    pageNavLabel: "Page navigation",
    copy: "Copy",
    copied: "Copied!",
    chartLoading: "Loading chart…",
    languageSwitcherLabel: "Language",
    langZh: "中文",
    langEn: "EN",
    langJa: "日本語",
    originalBadge: "Original: Chinese",
    originalBadgeNote: "This block is not translated yet; the full original Chinese text follows.",
    untranslatedNoticeTitle: "Translation status",
    untranslatedNotice:
      "This page's body is not translated yet. The complete original Chinese text follows, with every original block clearly marked.",
    toneNote: "Note",
    toneTip: "Tip",
    toneWarn: "Warning",
    toneDanger: "Important",
  },
  ja: {
    brandTitle: "大道帰一",
    brandSubtitle: "ゲーム数値と長期運用ハンドブック",
    topbarTag: "コンパニオンドキュメント",
    menuToggleLabel: "サイドバーを切り替え",
    closeSidebarLabel: "サイドバーを閉じる",
    sidebarFoot: "数値は健康診断の結果であって、運命の判決ではない。",
    tocTitle: "このページの目次",
    prevPage: "← 前のページ",
    nextPage: "次のページ →",
    pageNavLabel: "ページナビゲーション",
    copy: "コピー",
    copied: "コピーしました",
    chartLoading: "チャートを読み込み中…",
    languageSwitcherLabel: "言語",
    langZh: "中文",
    langEn: "EN",
    langJa: "日本語",
    originalBadge: "原文は中国語",
    originalBadgeNote: "このブロックは未翻訳です。以下に中国語の原文を全文表示します。",
    untranslatedNoticeTitle: "翻訳状況",
    untranslatedNotice:
      "このページの本文は未翻訳です。以下に中国語の原文を全文掲載し、各原文ブロックに明示的なマークを付けています。",
    toneNote: "メモ",
    toneTip: "ヒント",
    toneWarn: "注意",
    toneDanger: "重要",
  },
};

/** 读取时间仍然是中文数据源（如 "5 分钟"）；这里只替换单位与语序。 */
export function formatReadingTime(lang: Lang, readingTimeZh: string): string {
  const matched = readingTimeZh.match(/\d+/);
  const minutes = matched ? matched[0] : readingTimeZh;

  if (lang === "en") {
    return `~${minutes} min read`;
  }
  if (lang === "ja") {
    return `約${minutes}分`;
  }
  return `约 ${minutes} 分钟`;
}

/** 侧边栏分组名沿用中文 section 作为稳定的分组键，显示时再本地化。 */
const sectionLabels: Record<Lang, Record<string, string>> = {
  "zh-CN": {
    开始: "开始",
    地基: "地基",
    图解: "图解",
    专业: "专业",
    运营: "运营",
    设计: "设计",
    AI: "AI",
    安装: "安装",
  },
  en: {
    开始: "Start",
    地基: "Foundations",
    图解: "Diagrams",
    专业: "Professional",
    运营: "Operations",
    设计: "Design",
    AI: "AI",
    安装: "Install",
  },
  ja: {
    开始: "はじめに",
    地基: "基礎",
    图解: "図解",
    专业: "プロフェッショナル",
    运营: "運用",
    设计: "デザイン",
    AI: "AI",
    安装: "インストール",
  },
};

export function sectionLabel(lang: Lang, sectionZh: string): string {
  return sectionLabels[lang][sectionZh] ?? sectionZh;
}
