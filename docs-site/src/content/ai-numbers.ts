import type { DocPage } from "../types";
import { callout, code, h2, h3, ol, p, table, ul } from "../types";
import { translationsFor } from "./i18n";

export const aiNumbers: DocPage = {
  id: "ai-numbers",
  nav: "AI 全权设计数值与百万级配表",
  title: "AI 全权设计数值与百万级配表分析",
  subtitle:
    "让 agent 在人类给定的约束内设计全部数值、流式吞吐一百万份 Excel/JSON/CSV，并对当前版本给出数值健康判断。",
  section: "AI",
  readingTime: "13 分钟",
  i18n: translationsFor("ai-numbers"),
  blocks: [
    callout(
      "note",
      "「全权」的准确含义",
      "AI 全权设计数值，不等于 AI 自己决定目标。目标（理想值、容差带、权重、价值预算）由人类给定；AI 负责在约束内生成配置、自检和迭代。目标是设计输入，不是设计输出。",
    ),

    h2("一、工作流总览"),
    table(
      ["阶段", "输入", "输出", "护栏"],
      [
        [
          "1. 定义约束",
          "品类杠杆、SU 系数、预算、上限",
          "一份机器可读的 spec",
          "人类签字；缺项即报错",
        ],
        ["2. 生成配置", "spec + schema", "Excel / JSON / CSV 配表", "schema 校验、精确到字符串"],
        ["3. 流式摄取", "1,000,000 份文件", "规范化记录流 + 聚合摘要", "不整表入内存；逐文件校验"],
        ["4. 静态校验", "记录流", "错误清单", "范围、引用、单调性、传递性"],
        ["5. 仿真", "配置 + 行为模型", "11 个指标快照", "固定随机种子，可复现"],
        ["6. 判定", "快照 + 品类杠杆", "score / verdict / diagnostics", "精确有理数评分"],
        ["7. 迭代", "diagnostics", "最小改动补丁", "只改 worst deviation 对应系统"],
        ["8. 人工闸门", "报告 + 差异", "上线决定", "AI 不得自行发布"],
      ],
    ),

    h2("二、第一阶段：把意图写成约束"),
    p(
      "在 agent 动笔之前，必须先有一份 spec：基准单位定义、所有 SU 折算率、每个养成系统的价值预算、每条曲线的上下限、11 个指标的品类杠杆（ideal / min / max / weight）。这份 spec 是人类审美的载体，也是后面所有自动化判断的依据。",
    ),
    code(
      "json",
      `{
  "base_unit": "1 SU = 1 分钟标准有效游戏时间",
  "exchange_rates": { "gold": "1/20", "gem": "6/5", "stamina": "6" },
  "budgets": { "daily_su_in": "100", "daily_su_out": "100" },
  "curves": {
    "atk": { "base": "100", "step": "8", "cap": "1200", "piecewise": [] }
  },
  "levers": {
    "retention_d1": { "ideal": "0.44", "min": "0.30", "max": "0.58", "weight": "1.2" }
  }
}`,
      "约束用字符串写精确值；agent 只填允许范围内的数字",
    ),

    h2("三、第二阶段：生成配表"),
    p(
      "agent 按 schema 批量生成配置。关键点有三个：第一，所有数值字段是字符串，避免出现浮点字面量；第二，生成时带确定性种子，同一 spec 永远得到同一批配表；第三，每份文件带内容哈希，便于去重与回滚。",
    ),
    ul(
      "文件级校验：字段齐全、类型正确、值在允许区间内。",
      "跨表校验：引用的 ID 必须存在，不能有悬空引用。",
      "单调性校验：成本曲线单调不减，奖励曲线不得出现倒挂。",
      "传递性校验：任意两种资源的折算率乘起来一致。",
      "预算校验：所有系统的 SU 收支加总后落在 spec 的预算内。",
    ),

    h2("四、第三阶段：流式吞吐一百万份文件"),
    p(
      "一百万份配表不能一次性读进内存。正确做法是流式管道：逐文件解析 → 规范化成统一记录 → 写入追加式存储 → 同时更新聚合摘要。聚合用 sketch（例如区间直方图、count-min sketch 或分位数摘要），内存占用与文件数无关。",
    ),
    code(
      "text",
      `for file in stream_files("configs/**/*.{xlsx,json,csv}"):     # 逐文件
    rows = parse(file)                                        # 统一解析
    validate_schema(rows)                                     # 立即校验
    normalized = [to_su(row, exchange_rates) for row in rows] # 归一化
    append_to_store(normalized)                               # 追加落盘
    summary.update(normalized)                                # 只更新摘要
    checkpoint(file.hash)                                     # 可恢复

# 最终只从摘要和少量采样中读回结果，不重新加载全部数据`,
    ),
    h3("工程要点"),
    ol(
      "分片：按目录或哈希前缀分片，可并行处理且失败可单独重试。",
      "断点：每处理一个文件记录 checkpoint，中断后从最后一个继续。",
      "去重：内容哈希相同的文件只处理一次，避免重复计数。",
      "确定性：固定随机种子与遍历顺序，保证两次运行结果一致。",
      "可追溯：每份文件到每条记录的映射都要可查，报告里能指出具体来源。",
    ),
    callout(
      "warn",
      "Excel 的特殊处理",
      "xlsx 是压缩包，解析成本远高于 JSON/CSV，并且合并单元格、公式、多 sheet 都会引入歧义。建议先离线转换成规范化的 CSV/JSON 中间格式，再做流式分析；不要在一百万份文件上直接跑 Excel 解析器。",
    ),

    h2("五、第四到六阶段：仿真与判定"),
    p(
      "静态校验通过后，用行为模型做 Monte Carlo 仿真：模拟一批虚拟玩家在配置下的 progression、经济收支与付费行为，产出 11 个指标的快照。快照交给 ops.Evaluate，用品类杠杆算出 fitness 与加权 score。",
    ),
    code(
      "go",
      `snapshot := ops.Snapshot{
    GenreID: g.ID,
    Values: map[string]num.Num{
        genre.MetricRetentionD1: num.MustFromString(sim.D1.String()),
        // ... 11 个键齐全
    },
}
report, err := ops.Evaluate(snapshot)
// report.Diagnostics 按 critical 优先、deviation×weight 降序排列`,
    ),
    p(
      "报告要回答的不是「好不好」这种模糊判断，而是三个具体问题：当前构建的 score 是多少？哪几个杠杆是 critical 或 warn？每个问题最可能由哪个系统驱动？只有把问题归因到系统，agent 才知道该改哪里。",
    ),

    h2("六、第七到八阶段：迭代与人工闸门"),
    p(
      "agent 只修改驱动 worst deviation 的那个系统，生成一个最小补丁，然后重跑静态校验、仿真和评分。它会持续迭代直到 score 达标或触及改动预算上限，并保留每一步的差异记录，使整个过程可审计。",
    ),
    callout(
      "danger",
      "上线决定必须由人做",
      "数值健康只是上线门槛之一。手感、美术、叙事、合规、市场时机都不在评分里。「AI 全权设计数值」指的是执行层面的自主，不是责任层面的转移。",
    ),
    p(
      "下一页讨论一个可选的增强：把 codegraph233 接进来，让 agent 能沿代码调用链理解玩法流程，从而做「心流」分析。这部分较慢、可选，且在本站生成环境中未做端到端验证。",
    ),
  ],
};
