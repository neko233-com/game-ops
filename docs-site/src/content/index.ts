import type { DocPage } from "../types";
import { overview } from "./overview";
import { oneness } from "./oneness";
import { normalization } from "./normalization";
import { additiveFirst } from "./additive-first";
import { valueFlow } from "./value-flow";
import { systemMap } from "./system-map";
import { pillars } from "./pillars";
import { numbers } from "./numbers";
import { aiNumbers } from "./ai-numbers";
import { flow } from "./flow";
import { install } from "./install";

export const pages: DocPage[] = [
  overview,
  oneness,
  normalization,
  additiveFirst,
  valueFlow,
  systemMap,
  pillars,
  numbers,
  aiNumbers,
  flow,
  install,
];

/** 侧边栏分组顺序。未在列表里的分组按出现顺序排在末尾。 */
export const sectionOrder = ["开始", "地基", "图解", "运营", "设计", "AI", "安装"];

export function groupedPages(): Array<{ section: string; pages: DocPage[] }> {
  const groups = new Map<string, DocPage[]>();

  for (const page of pages) {
    const bucket = groups.get(page.section);

    if (bucket) {
      bucket.push(page);
    } else {
      groups.set(page.section, [page]);
    }
  }

  const ordered: Array<{ section: string; pages: DocPage[] }> = [];

  for (const section of sectionOrder) {
    const bucket = groups.get(section);

    if (bucket) {
      ordered.push({ section, pages: bucket });
      groups.delete(section);
    }
  }

  for (const [section, bucket] of groups) {
    ordered.push({ section, pages: bucket });
  }

  return ordered;
}

export function getPage(id: string): DocPage | undefined {
  return pages.find((page) => page.id === id);
}

export function neighbours(id: string): { prev?: DocPage; next?: DocPage } {
  const index = pages.findIndex((page) => page.id === id);

  if (index < 0) {
    return {};
  }

  return {
    prev: index > 0 ? pages[index - 1] : undefined,
    next: index < pages.length - 1 ? pages[index + 1] : undefined,
  };
}
