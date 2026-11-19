# 大道归一 · 游戏数值与长线运营手册（文档站）

这是一个配套主仓库（纯 Go 的「游戏长线运营」引擎 + agent skill）的**独立文档站**。它用中文系统讲解数值设计与长线运营的基础概念，用统一的「大道归一」理论串起来。

技术栈：**Vite+（vite-plus / VoidZero）+ React + TypeScript**，图解用 **React Flow（`@xyflow/react` v12）**，专业图表用 **Apache ECharts（`echarts` v6，直接的 React hook，无 echarts-for-react 包装）**。只使用 npm，不依赖全局 `vp` CLI。

## 运行

需要 Node.js 20.19+ / 22.12+（推荐 Node 26）与 npm 11+。在 `docs-site/` 目录下：

```bash
npm install
npm run dev      # 启动开发服务器
npm run build    # 生产构建，输出到 dist/
npm run preview  # 本地预览生产构建
npm run check    # Vite+ 检查：格式化检查 + lint + 类型检查
```

`dev` / `build` / `preview` / `check` 脚本内部调用项目本地的 `vp`（来自 `vite-plus` 依赖），因此不需要全局安装 Vite+ CLI。

## 多语言（i18n）

文档站支持三种语言：`zh-CN`（简体中文，默认且权威）、`en`（English）、`ja`（日本語）。顶栏右侧的分段控件 `中文 / EN / 日本語` 可即时切换，切换时**保留当前页面 hash 路由**，不需要重新导航；`<html lang>` 会同步更新。

**默认语言规则（优先中文）**：首次访问时**始终显示简体中文**。即使浏览器语言是英文或日文，也不会自动切换到 en/ja；只有用户在顶栏显式选择后才会改变，并写入 `localStorage`（键 `docs-site-lang`）持久化。后续访问优先读取 `localStorage`。`navigator.language` 仍会被读取，但仅用于确认中文系浏览器环境；非中文一律回落到 `zh-CN`。这是刻意的产品决策：中文是权威完整版本，翻译是增量交付。

### 实现方式

- `src/i18n/index.ts`：`Lang` 类型（`"zh-CN" | "en" | "ja"`）、`DEFAULT_LANG`、`I18nProvider`、`useLang()` hook（返回 `{ lang, setLang, t }`）、`detectInitialLang()`（localStorage → navigator.language → 默认中文）。无额外依赖。
- `src/i18n/ui.ts`：所有界面文案的字典 `messages`；侧边栏分组名用 `sectionLabel()`，阅读时间用 `formatReadingTime()`（把 "5 分钟" 转成对应语言的语序）。涵盖分组名、品牌标语、翻页按钮、TOC 标题、复制按钮、加载文案、提示框徽标、页脚、语言切换器标签等全部 UI chrome。
- `src/content/*.ts`：每页 `DocPage` 带可选的 `i18n: { en?, ja? }` 字段（`PageTranslation`：`nav` / `title` / `subtitle` / `blocks`）。中文正文始终留在原文件里，是唯一权威版本。
- `src/content/i18n/en.ts` / `ja.ts`：译文数据。前 3 页（概览、大道归一、归一化与单位价值）提供**完整正文翻译**；其余页面提供导航/标题/副标题翻译。`src/content/i18n/index.ts` 的 `translationsFor(id)` 把译文挂到对应 `DocPage.i18n`。
- `src/content/index.ts` 的 `localizePage(page, lang)`：中文直接返回原文；译文若没有 `blocks`，则把所有原文块包成 `untranslated` 块，渲染时显示「原文为中文 / Original: Chinese / 原文は中国語」徽标并完整展示中文原文。这样绝不会出现半翻译的句子。未翻译页面不显示 TOC（避免目录里出现中文标题与英文页面混排）。

### 增加/修改翻译

1. 改 UI 文案：编辑 `src/i18n/ui.ts` 中三种语言的 `messages` 条目。
2. 改某页译文：编辑 `src/content/i18n/en.ts` 或 `ja.ts` 中对应 id 的 `PageTranslation`。
3. 想把某页从未翻译升级为完整翻译：在该页译文里补上完整的 `blocks` 数组（复用 `src/types.ts` 的 `p` / `h2` / `table` / `callout` 等构造函数）。`localizePage` 会自动改用译文 blocks，不再显示原文徽标。
4. 新增页面：在 `src/content/` 建数据文件并加入 `pages`；译文可选，缺省时 en/ja 会显示 nav/title/subtitle 的中文回退 + 原文徽标（此时建议至少补上 nav/title/subtitle）。

> 图表页（React Flow / ECharts）的正文在 en/ja 下同样以原文徽标呈现，保证语言一致性；图表内部节点/标题目前沿用中文（数值与公式本身是通用记号），待后续逐图补齐译文。

## 内容结构

正文全部放在 `src/content/*.ts`，每个文件导出一个 `DocPage` 数据对象；`App.tsx` 只负责渲染，不写正文。新增一页只需：

1. 在 `src/content/` 新建一个数据文件；
2. 在 `src/content/index.ts` 里导入并加入 `pages` 数组；
3. （可选）在 `sectionOrder` 里加分组名。

可用的内容块类型定义在 `src/types.ts`：段落、二级/三级标题、有序/无序列表、代码块、表格、提示框（note/tip/warn/danger）、公式、步骤、引用，以及 `flow` 交互式图解块（`value-flow` / `system-map`，由 `src/flow/FlowDiagrams.tsx` 渲染）和 `echarts` 专业图表块（chart id 由 `src/charts/ProCharts.tsx` 渲染）。正文支持少量行内标记：`**加粗**` 与 `` `行内代码` ``。

## 交互式图解（React Flow）

新增两页面向老板 / 非数值读者的可视化图解，节点可拖动、画布可缩放平移、点节点出大白话解释：

- **价值流动图（互动）**：产出源 → 玩家钱包（基准单位金币）→ 消耗口，边上标数值，配「进 ≈ 出，比值≈1」读图口诀。
- **系统关系图（互动）**：战斗数值 → 战力 → 成长曲线 → 经济系统 → 商业化 → 留存，带「加法优先 / 乘法叠加」开关，对比安全直线与 1.1ⁿ 复利风险表。

两页在侧边栏归到「图解」分组。`@xyflow/react` 的基础样式在 `src/main.tsx` 引入，自定义样式在 `src/flow/flow.css`。

## 专业数值分析（Apache ECharts）

新增一页面向数值主策划的密度视图，侧边栏归到「专业」分组。图表用 `echarts` npm 包直接通过一个小 React hook 驱动（`src/charts/ProCharts.tsx`：`useRef` + `useEffect` + `echarts.init`，`ResizeObserver` 自适应，卸载 `dispose`），没有使用 `echarts-for-react` 包装库。五张图均带中文图例、tooltip、坐标轴与精确数值：

- **成长曲线对比**：加法（100+25(n−1)，封顶 10000）vs 指数（100×1.1^(n−1)），双折线 + dataZoom 滑块 + 线性/对数纵轴切换 + L40 标注（指数≈加法 3.83×，1.1^40≈45.26×）。
- **满级时间分布**：总经验 40120 / 日经验 900·480·1500 推导的 P10–P90 分组柱，叠加 D1/D7/D30 留存插值的「留存到达率」折线（次轴）。
- **经济流向 Sankey**：每日任务 100 / 主线关卡 50 / 充值 30（钻石折算）→ 玩家钱包 180 → 升级 100 / 抽卡 50 / 入场券 30。
- **数值敏感度热力图**：gacha 品类付费率 × D1 留存的 fitness 加权均值，含 visualMap 与理想点标记；fitness=clamp(1−|obs−ideal|/band,0,1)。
- **配表异常散点**：各配置列 max/min 倍数（对数轴），10× 审查阈值线，红/青区分越线与否。

所有底层引擎计算使用精确有理数 `num.Num`（配表字符串→分数，全程无浮点误差），图表只在展示层格式化小数。相关样式在 `src/styles.css` 的「专业图表」小节。

页面清单：

| 页面                        | 主题                                                                 |
| --------------------------- | -------------------------------------------------------------------- |
| 首页 / 概览                 | 站点定位与五分钟快速上手                                             |
| 大道归一：数值的核心        | 价值的定义、流动、节奏；数值能否预测游戏寿命的诚实讨论               |
| 归一化与单位价值            | 基准单位、精确有理数交换率、可传递性、五资源折算算例                 |
| 为什么优先加减法而非乘法    | 加法优先原则、六种乘法代价、何时乘法合法、行业踩坑表                 |
| 价值流动图（互动）          | React Flow：产出 → 钱包 → 消耗，进≈出的读图口诀                      |
| 系统关系图（互动）          | React Flow：六大系统因果链 + 加法/乘法对比开关                       |
| 专业数值分析（ECharts）     | ECharts：成长曲线、满级时间、经济 Sankey、敏感度热力图、配表异常散点 |
| 五大支柱与 11 个指标        | 通用词汇表、健康评分公式与阈值、控制回路                             |
| 战斗 / 成长 / 商业化数值    | TTK 反推、价值预算、价格曲线与 LTV 纪律                              |
| AI 全权设计数值与百万级配表 | 八阶段工作流、流式摄取、仿真判定、人工闸门                           |
| 心流分析 + codegraph233     | 心流代理指标、AST GraphRAG 集成（可选、较慢、未端到端验证）          |
| 全运行时安装                | opencode / codex / dsh / claude，项目优先 + 全局可选                 |

## 说明

- 本站是纯前端静态站点，不涉及任何后端或数据库。
- 文档中的 Go 示例来自主仓库的 API（`num` / `genre` / `ops` / `formula`），仅作说明，不在此项目中编译。
- 构建产物在 `dist/`，可以直接部署到任意静态托管服务。
