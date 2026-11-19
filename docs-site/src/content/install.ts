import type { DocPage } from "../types";
import { callout, code, h2, h3, ol, p, table, ul } from "../types";

export const install: DocPage = {
  id: "install",
  nav: "全运行时安装",
  title: "全运行时安装（项目优先）",
  subtitle:
    "opencode / codex / dsh / claude 四种运行时都支持。先把 skill 装进项目，再考虑全局；从 GitHub 复制源码并读 README 快速开始。",
  section: "安装",
  readingTime: "10 分钟",
  blocks: [
    h2("一、总原则：项目优先，全局可选"),
    p(
      "这套能力由两部分组成：一个纯 Go 的引擎（num / genre / ops / formula / cmd/opsreport）和一个 agent skill。两者都应该先以项目级方式安装——随仓库版本化、随代码一起 review、换机器后能原样复现。全局安装只提供便利，不应成为团队依赖。",
    ),
    table(
      ["维度", "项目级（推荐）", "全局"],
      [
        ["位置", "写在项目目录内的 skill/运行时配置目录", "用户主目录下的运行时全局目录"],
        ["版本", "随仓库 pin 住，可与版本对应", "所有人可能不同，难以复现"],
        ["协作", "可 review、可 diff、可随分支切换", "不进版本控制"],
        ["适用", "真实项目、团队、CI", "个人临时试验、跨项目复用"],
      ],
    ),
    callout(
      "tip",
      "决策规则",
      "如果你会把它用于一个真实仓库，就一定装在项目级。只有当你在多个不相干的个人项目里做快速试验时，才考虑全局安装，且仍应在项目里保留一份副本。",
    ),

    h2("二、从 GitHub 复制 + 读 README"),
    p(
      "最稳妥的安装方式不是依赖包管理器，而是直接复制仓库里的 skill 源码。它保证你拿到的是 README 里描述的那一份，也方便后续手工升级。",
    ),
    ol(
      "从 GitHub 获取仓库源码（clone 或下载 zip 均可）。",
      "把其中的 skill 目录复制到目标项目的运行时配置目录（见下表）。",
      "打开复制进来的 README，按 Quickstart 一节逐条执行，确认能跑通最小示例。",
      "把 Go 引擎放到项目约定位置，运行 go run ./cmd/opsreport examples/gacha.json 验证。",
      "提交到版本控制，确保其他人 checkout 后得到同样的能力。",
    ),
    code(
      "bash",
      `git clone <repo-url> game-ops
# 然后把 game-ops/.opencode/skills/game-long-term-operation
# 复制到你的项目对应目录（见下表）`,
    ),
    code(
      "bash",
      `go run ./cmd/opsreport examples/gacha.json
# 期望：输出 score、verdict 与 diagnostics，而不是报错`,
      "最小验证：任何运行时安装成功后都应能跑通这一步",
    ),

    h2("三、四种运行时的项目级位置"),
    p("不同运行时约定的配置目录不同。下表给出典型的放置方式；具体路径仍以各运行时当前文档为准。"),
    table(
      ["运行时", "项目级位置（示意）", "全局位置（示意）", "备注"],
      [
        [
          "opencode",
          ".opencode/skills/<skill-name>/",
          "用户主目录下的 opencode 配置",
          "本仓库自带 .opencode/skills 示例，可直接复制",
        ],
        [
          "codex",
          ".codex/skills/<skill-name>/",
          "用户主目录下的 codex 配置",
          "以项目内 instructions/skills 目录为准",
        ],
        [
          "dsh",
          ".dsh/skills/<skill-name>/",
          "用户主目录下的 dsh 配置",
          "按 dsh 的 skill 加载约定放置",
        ],
        [
          "claude",
          ".claude/skills/<skill-name>/",
          "用户主目录下的 claude 配置",
          "项目级 CLAUDE.md / skills 优先于全局",
        ],
      ],
    ),
    p(
      "共同点：每个运行时都识别「项目内的 skill 目录」。把复制过来的 skill 放在项目目录里，agent 在当前仓库工作时就会加载它，而不需要任何全局配置。",
    ),

    h3("opencode 示例"),
    code(
      "text",
      `<project>/.opencode/skills/game-long-term-operation/
├── SKILL.md              # 读这里的 Quickstart
└── references/
    ├── metrics.md        # 11 个指标定义与目标带
    └── genres.md         # 品类杠杆与注册方式`,
    ),

    h2("四、README Quickstart 应该包含什么"),
    ul(
      "前置条件：Go 版本、Node（若用本站文档）与对应运行时。",
      "安装步骤：复制 skill 到项目目录的命令。",
      "最小示例：一份 JSON 快照 + go run ./cmd/opsreport 的预期输出。",
      "关键概念链接：指向本站的「五大支柱」「归一化」等页面。",
      "升级与卸载：如何覆盖旧版本、如何移除。",
    ),
    callout(
      "warn",
      "不要跳过 README",
      "运行时约定会随版本变化。复制文件只是第一步，真正保证可用的是按 README 验证一遍；如果 README 与本站描述不一致，以 README 为准。",
    ),

    h2("五、常见错误"),
    table(
      ["错误", "后果", "修正"],
      [
        ["只装全局不装项目级", "团队其他成员没有该能力，CI 无法复现", "把 skill 复制进项目并提交"],
        ["只复制 SKILL.md 不带 references", "指标定义缺失，agent 凭记忆作答", "整目录复制"],
        ["把配置数值写成 float", "精度丢失，报告不可复算", "用字符串 + 精确有理数"],
        ["跳过最小示例", "安装未完成就投入生产", "先跑通 go run 示例"],
        ["skill 与引擎版本不一致", "指标键或杠杆表对不上", "一起版本化，一起升级"],
      ],
    ),

    h2("六、安装之后"),
    p(
      "回到本站的「首页 / 概览」按五分钟流程走一遍，或者直接进入「五大支柱与 11 个指标」，把你项目的真实数据填进快照。记住：一次评分只是体检；在每个内容节奏点重新快照，才是运营。",
    ),
  ],
};
