/** 把 en/ja 译文按页面 id 挂到 DocPage.i18n 上。 */

import type { DocPage } from "../../types";
import { enTranslations } from "./en";
import { jaTranslations } from "./ja";

export function translationsFor(id: string): DocPage["i18n"] {
  return {
    en: enTranslations[id],
    ja: jaTranslations[id],
  };
}
