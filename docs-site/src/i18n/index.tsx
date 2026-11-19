/**
 * 轻量级 i18n：无第三方依赖，仅用 React Context + localStorage。
 *
 * 设计决策（优先中文）：
 *   - 中文 (zh-CN) 是默认且权威的语言。
 *   - 首次访问时即使浏览器语言是 en / ja，也**不会**自动切到英文/日文；
 *     只有用户通过顶栏切换器显式选择后才会改变。
 *   - navigator.language 仍会被读取：仅当浏览器语言是中文系时“确认”中文；
 *     其它情况一律回落到 zh-CN。这样满足“检测浏览器语言”，又坚守“优先中文”。
 *   - 用户选择会持久化到 localStorage，刷新后保持。
 */

import { createContext, useContext, useEffect, useMemo, useState } from "react";
import type { ReactNode } from "react";
import { messages } from "./ui";
import type { UiKey } from "./ui";

export type Lang = "zh-CN" | "en" | "ja";

export const DEFAULT_LANG: Lang = "zh-CN";
export const LANGS: readonly Lang[] = ["zh-CN", "en", "ja"];

const STORAGE_KEY = "docs-site-lang";

function isLang(value: string | null): value is Lang {
  return value === "zh-CN" || value === "en" || value === "ja";
}

/** 计算首次访问使用的语言。顺序：持久化 → 浏览器语言 → 默认中文。 */
export function detectInitialLang(): Lang {
  if (typeof window !== "undefined") {
    try {
      const saved = window.localStorage.getItem(STORAGE_KEY);
      if (isLang(saved)) {
        return saved;
      }
    } catch {
      // localStorage 可能被禁用，忽略即可。
    }

    const browserLang = window.navigator.language ?? "";
    // 浏览器是中文系 → 确认中文；其它语言 → 仍然优先中文（见文件头说明）。
    if (/^zh\b/i.test(browserLang)) {
      return "zh-CN";
    }
  }

  return DEFAULT_LANG;
}

export interface I18nContextValue {
  lang: Lang;
  setLang: (lang: Lang) => void;
  t: (key: UiKey) => string;
}

const I18nContext = createContext<I18nContextValue | undefined>(undefined);

export function I18nProvider({ children }: { children: ReactNode }) {
  const [lang, setLangState] = useState<Lang>(detectInitialLang);

  useEffect(() => {
    document.documentElement.lang = lang;
    try {
      window.localStorage.setItem(STORAGE_KEY, lang);
    } catch {
      // 忽略持久化失败。
    }
  }, [lang]);

  const value = useMemo<I18nContextValue>(
    () => ({
      lang,
      setLang: setLangState,
      t: (key: UiKey) => messages[lang][key],
    }),
    [lang],
  );

  return <I18nContext.Provider value={value}>{children}</I18nContext.Provider>;
}

export function useLang(): I18nContextValue {
  const ctx = useContext(I18nContext);
  if (!ctx) {
    throw new Error("useLang 必须在 I18nProvider 内使用");
  }
  return ctx;
}
