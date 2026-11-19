/**
 * 两张给老板 / 非数值读者看的交互式图解。
 *
 * 用 React Flow（@xyflow/react v12）画出「钱怎么流」和「系统怎么相互影响」。
 * 节点可拖动、画布可缩放平移，点节点会用大白话解释它在讲什么。
 */

import { useMemo, useState } from "react";
import {
  Background,
  Controls,
  Handle,
  MarkerType,
  MiniMap,
  Position,
  ReactFlow,
  useNodesState,
  type Edge,
  type Node,
  type NodeMouseHandler,
  type NodeProps,
} from "@xyflow/react";
import type { FlowDiagram } from "../types";

type NodeTone = "source" | "wallet" | "sink" | "system";

type FlowNodeData = {
  emoji: string;
  title: string;
  value: string;
  subtitle: string;
  tone: NodeTone;
  explain: string;
};

type DiagramNode = Node<FlowNodeData, "info">;

/** 一个通用的「信息卡片」节点，两张图都用它。 */
function DiagramNodeView({ data, selected }: NodeProps<DiagramNode>) {
  return (
    <div className={`rf-node rf-node-${data.tone}${selected ? " rf-node-selected" : ""}`}>
      <Handle type="target" position={Position.Left} className="rf-handle" />
      <div className="rf-node-head">
        <span className="rf-node-emoji" aria-hidden="true">
          {data.emoji}
        </span>
        <span className="rf-node-subtitle">{data.subtitle}</span>
      </div>
      <div className="rf-node-title">{data.title}</div>
      {data.value ? <div className="rf-node-value">{data.value}</div> : null}
      <Handle type="source" position={Position.Right} className="rf-handle" />
    </div>
  );
}

const nodeTypes = { info: DiagramNodeView };

/** 点击节点后，用大白话解释它。 */
function ExplainPanel({ data, idle }: { data?: FlowNodeData; idle: string }) {
  return (
    <div className={`rf-explain${data ? " rf-explain-active" : ""}`}>
      {data ? (
        <>
          <div className="rf-explain-head">
            <span className="rf-explain-emoji" aria-hidden="true">
              {data.emoji}
            </span>
            <strong>{data.title}</strong>
          </div>
          <p>{data.explain}</p>
        </>
      ) : (
        <p className="rf-explain-idle">{idle}</p>
      )}
    </div>
  );
}

/** 带controls / minimap / 点击解释的画布外壳。 */
function DiagramCanvas({
  nodes: initialNodes,
  edges,
  idleText,
  miniMapColor = "#6366f1",
}: {
  nodes: DiagramNode[];
  edges: Edge[];
  idleText: string;
  miniMapColor?: string;
}) {
  const [nodes, , onNodesChange] = useNodesState<DiagramNode>(initialNodes);
  const [selected, setSelected] = useState<FlowNodeData | undefined>(undefined);

  const onNodeClick: NodeMouseHandler<DiagramNode> = (_, node) => {
    setSelected(node.data);
  };

  return (
    <div className="rf-wrap">
      <div className="rf-canvas">
        <ReactFlow
          nodes={nodes}
          edges={edges}
          onNodesChange={onNodesChange}
          onNodeClick={onNodeClick}
          nodeTypes={nodeTypes}
          fitView
          fitViewOptions={{ padding: 0.2 }}
          nodesDraggable
          nodesConnectable={false}
          elementsSelectable
          panOnDrag
          zoomOnScroll
          zoomOnPinch
          minZoom={0.4}
          maxZoom={1.8}
        >
          <Background color="#cbd5e1" gap={18} size={1} />
          <Controls showInteractive={false} />
          <MiniMap pannable zoomable nodeColor={miniMapColor} maskColor="rgba(15,23,42,0.08)" />
        </ReactFlow>
      </div>
      <ExplainPanel data={selected} idle={idleText} />
    </div>
  );
}

function makeEdge(source: string, target: string, label: string, color: string): Edge {
  return {
    id: `${source}-${target}`,
    source,
    target,
    type: "smoothstep",
    animated: true,
    label,
    style: { stroke: color, strokeWidth: 2 },
    labelStyle: { fill: color, fontWeight: 700, fontSize: 12 },
    labelBgStyle: { fill: "#ffffff", fillOpacity: 0.95 },
    labelBgPadding: [6, 3],
    labelBgBorderRadius: 6,
    markerEnd: { type: MarkerType.ArrowClosed, color, width: 18, height: 18 },
  };
}

/* ------------------------------------------------------------------ */
/* 图 A：价值流动图                                                    */
/* ------------------------------------------------------------------ */

const valueNodes: DiagramNode[] = [
  {
    id: "daily",
    type: "info",
    position: { x: 0, y: 0 },
    data: {
      emoji: "📅",
      title: "每日任务",
      subtitle: "产出源 · 水龙头",
      value: "100 金币 / 日",
      tone: "source",
      explain:
        "这是钱进来的地方。玩家每天上线做任务，就能稳定拿到 100 金币。它像一直开着的水龙头，是经济里最可靠的一笔收入。",
    },
  },
  {
    id: "main",
    type: "info",
    position: { x: 0, y: 170 },
    data: {
      emoji: "🗺️",
      title: "主线关卡",
      subtitle: "产出源 · 一次性",
      value: "50 金币 / 关",
      tone: "source",
      explain:
        "推关也会给钱，但它是一次性的：关卡越推越深，玩家能领的次数越少。它适合开局，不适合当成长期收入。",
    },
  },
  {
    id: "recharge",
    type: "info",
    position: { x: 0, y: 340 },
    data: {
      emoji: "💎",
      title: "充值 / 钻石",
      subtitle: "产出源 · 花钱买",
      value: "3 钻石 = 30 金币",
      tone: "source",
      explain:
        "这是玩家真金白银买来的钱。钻石本身不能直接和金币相加，要先按 1 钻石 = 10 金币折一下：3 钻石 = 30 金币。折完才能和任务、关卡放在同一把尺子上比较。",
    },
  },
  {
    id: "wallet",
    type: "info",
    position: { x: 330, y: 170 },
    data: {
      emoji: "💰",
      title: "玩家钱包",
      subtitle: "基准单位：金币",
      value: "今日合计 180 金币",
      tone: "wallet",
      explain:
        "金币是这张图的尺子（基准单位）。别的资源都要先折成它再相加。这里把三路收入汇总：100 + 50 + 30 = 180 金币/日，这就是玩家一天能拿到的钱。",
    },
  },
  {
    id: "upgrade",
    type: "info",
    position: { x: 660, y: 0 },
    data: {
      emoji: "⬆️",
      title: "升级",
      subtitle: "消耗口 · 下水口",
      value: "100 金币 / 日",
      tone: "sink",
      explain:
        "这是钱花掉的地方。角色升级吃掉最多的金币，是最大的下水口。水龙头开得再大，也要有下水口接住，钱才不会堆在钱包里。",
    },
  },
  {
    id: "gacha",
    type: "info",
    position: { x: 660, y: 170 },
    data: {
      emoji: "🎴",
      title: "抽卡",
      subtitle: "消耗口 · 换奖励",
      value: "50 金币 / 日",
      tone: "sink",
      explain:
        "抽卡把金币换成随机奖励，是另一类消耗。它很吸引人，但要盯紧：如果抽卡花得比产出还快，玩家很快就会没得抽。",
    },
  },
  {
    id: "ticket",
    type: "info",
    position: { x: 660, y: 340 },
    data: {
      emoji: "🎟️",
      title: "入场券",
      subtitle: "消耗口 · 控节奏",
      value: "30 金币 / 日",
      tone: "sink",
      explain:
        "门票类消耗，控制玩家每天能打几次高收益副本。它花得不多，作用却很大：没有它，玩家可能一天就把一周的内容刷完。",
    },
  },
];

const valueEdges: Edge[] = [
  makeEdge("daily", "wallet", "100 金币", "#0f766e"),
  makeEdge("main", "wallet", "50 金币", "#0f766e"),
  makeEdge("recharge", "wallet", "3 钻石 = 30 金币", "#0f766e"),
  makeEdge("wallet", "upgrade", "100 金币", "#b45309"),
  makeEdge("wallet", "gacha", "50 金币", "#b45309"),
  makeEdge("wallet", "ticket", "30 金币", "#b45309"),
];

export function ValueFlowDiagram() {
  return (
    <div className="rf-section">
      <div className="rf-balance" role="note">
        <span className="rf-balance-item rf-balance-in">
          🚰 水龙头（进）：<strong>180 金币 / 日</strong>
        </span>
        <span className="rf-balance-eq">≈</span>
        <span className="rf-balance-item rf-balance-out">
          🚿 下水口（出）：<strong>180 金币 / 日</strong>
        </span>
        <span className="rf-balance-result">进出平衡 · 比值 ≈ 1.0 ✅</span>
      </div>
      <DiagramCanvas
        nodes={valueNodes}
        edges={valueEdges}
        idleText="👆 点一下图里的卡片，我用大白话告诉你它是什么、为什么重要。卡片也能拖动，画布可以缩放和平移。"
        miniMapColor="#4f46e5"
      />
      <p className="rf-footnote">
        左边是钱进来的地方（绿色），中间把钱折成同一种单位汇总，右边是钱花掉的地方（橙色）。
      </p>
    </div>
  );
}

/* ------------------------------------------------------------------ */
/* 图 B：系统关系图                                                    */
/* ------------------------------------------------------------------ */

const systemNodes: DiagramNode[] = [
  {
    id: "combat",
    type: "info",
    position: { x: 0, y: 0 },
    data: {
      emoji: "⚔️",
      title: "① 战斗数值",
      subtitle: "起点",
      value: "伤害 / 血量 / TTK",
      tone: "system",
      explain:
        "一切从战斗里的数字开始：一拳打多少、怪物有多少血、几回合打完。这些数值决定了后面所有系统的节奏。",
    },
  },
  {
    id: "power",
    type: "info",
    position: { x: 300, y: 0 },
    data: {
      emoji: "💪",
      title: "② 战力",
      subtitle: "集中体现",
      value: "一个看得见的数字",
      tone: "system",
      explain: "战力是把一堆战斗数值压缩成一个玩家看得懂的数字。它涨得多快，就是数值膨胀的速度。",
    },
  },
  {
    id: "curve",
    type: "info",
    position: { x: 600, y: 0 },
    data: {
      emoji: "📈",
      title: "③ 成长曲线",
      subtitle: "节奏",
      value: "每一级要多少投入",
      tone: "system",
      explain:
        "升级要多少经验、多少金币，连起来就是成长曲线。它决定玩家是「一路顺风」还是「卡到想放弃」。",
    },
  },
  {
    id: "econ",
    type: "info",
    position: { x: 600, y: 210 },
    data: {
      emoji: "🪙",
      title: "④ 经济系统",
      subtitle: "产出与消耗",
      value: "金币 / 钻石的进出",
      tone: "system",
      explain:
        "升级要花钱，钱从哪来、花到哪去，就是经济系统。升级成本一变，整个金币的进出都会跟着变。",
    },
  },
  {
    id: "monet",
    type: "info",
    position: { x: 300, y: 210 },
    data: {
      emoji: "💳",
      title: "⑤ 商业化",
      subtitle: "付费点",
      value: "卖什么、卖多少钱",
      tone: "system",
      explain: "经济里缺什么，就卖什么。产出和消耗之间的缺口，决定了付费点该定在什么价位上。",
    },
  },
  {
    id: "retention",
    type: "info",
    position: { x: 0, y: 210 },
    data: {
      emoji: "🔁",
      title: "⑥ 留存与内容消耗",
      subtitle: "终点也回到起点",
      value: "玩家还玩不玩",
      tone: "system",
      explain:
        "玩家花钱、升级、打完内容后还愿不愿意回来，决定了游戏能活多久。它也会倒过来逼我们回头改战斗数值——所以这是一条回路。",
    },
  },
];

const systemCausality: Array<{ source: string; target: string; label: string }> = [
  { source: "combat", target: "power", label: "决定战力膨胀速度" },
  { source: "power", target: "curve", label: "涨得多快 = 升级手感" },
  { source: "curve", target: "econ", label: "升级成本决定金币进出" },
  { source: "econ", target: "monet", label: "经济缺口决定付费定价" },
  { source: "monet", target: "retention", label: "付费进度影响内容消耗" },
];

function curvePoints(danger: boolean): string {
  const points: string[] = [];

  for (let level = 0; level <= 40; level += 2) {
    const x = 8 + (level / 40) * 296;
    const value = danger ? Math.pow(1.1, level) : 1 + 0.1 * level;
    const y = 108 - ((value - 1) / (danger ? 44 : 4)) * 88;
    points.push(`${x.toFixed(1)},${y.toFixed(1)}`);
  }

  return points.join(" ");
}

function CurveChart({ danger }: { danger: boolean }) {
  return (
    <svg
      className="rf-curve"
      viewBox="0 0 320 120"
      role="img"
      aria-label={danger ? "乘法叠加的指数曲线，越到后面越陡" : "加法优先的直线曲线，每级增长一样"}
    >
      <line x1="8" y1="108" x2="306" y2="108" stroke="#94a3b8" strokeWidth="1" />
      <line x1="8" y1="108" x2="8" y2="10" stroke="#94a3b8" strokeWidth="1" />
      <polyline
        points={curvePoints(danger)}
        fill="none"
        stroke={danger ? "#ef4444" : "#0f766e"}
        strokeWidth="2.5"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <text x="12" y="24" className={danger ? "rf-curve-label-danger" : "rf-curve-label-safe"}>
        {danger ? "1.1ⁿ｜越涨越猛" : "1 + 0.1n｜匀速前进"}
      </text>
    </svg>
  );
}

const pitfallRows: string[][] = [
  [
    "每个系统各给一个乘算增伤",
    "上线数值好看",
    "版本越多越强，老内容被碾压",
    "同类先加总，乘一次，并设上限",
  ],
  [
    "每日收益按复利结算",
    "长期曲线很漂亮",
    "存量指数膨胀，数月内经济崩坏",
    "改成单利或带上限的分段收益",
  ],
  [
    "付费专属乘算层",
    "付费感强、短期收入高",
    "免费与付费差距指数张开",
    "付费买时间与外观，不买乘算层",
  ],
  [
    "战力按幂函数增长",
    "升级很有成就感",
    "边界爆炸，后期数值超出承载",
    "base + step×等级 + 分段封顶",
  ],
];

function AdditiveView() {
  return (
    <div className="rf-mode-body">
      <CurveChart danger={false} />
      <div className="rf-mode-text">
        <aside className="callout callout-tip">
          <div className="callout-title">
            <span className="callout-badge">建议</span>
            <strong>加法优先：每级都是同样的一小步</strong>
          </div>
          <p>
            每级只多给 0.1 倍，40 级后也只到 5
            倍左右。曲线是一条直线：今天加一个新系统，只影响往后的固定一小段，
            看得懂、算得准、出了问题也能退回去。
          </p>
        </aside>
        <table className="rf-mini-table">
          <thead>
            <tr>
              <th>等级</th>
              <th>L6</th>
              <th>L20</th>
              <th>L40</th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td>1 + 0.1×n</td>
              <td>1.6×</td>
              <td>3.0×</td>
              <td>5.0×</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  );
}

function MultiplicativeView() {
  return (
    <div className="rf-mode-body">
      <CurveChart danger />
      <div className="rf-mode-text">
        <aside className="callout callout-danger">
          <div className="callout-title">
            <span className="callout-badge">危险</span>
            <strong>乘法叠加：每次都在上一轮的结果上再乘一次</strong>
          </div>
          <p>
            同样写着「每级 +10%」，1.1ⁿ 在 <strong>L6 ≈ 1.77×</strong>、<strong>L20 ≈ 6.7×</strong>
            、<strong>L40 ≈ 45×</strong>。40 级后变成 45
            倍——老内容被碾压、经济会失控、谁都不敢再动数值。
          </p>
        </aside>
        <table className="rf-mini-table rf-mini-table-danger">
          <thead>
            <tr>
              <th>乘法叠加踩坑</th>
              <th>表面收益</th>
              <th>实际后果</th>
              <th>怎么修正</th>
            </tr>
          </thead>
          <tbody>
            {pitfallRows.map((row, index) => (
              <tr key={index}>
                {row.map((cell, cellIndex) => (
                  <td key={cellIndex}>{cell}</td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}

export function SystemMapDiagram() {
  const [danger, setDanger] = useState(false);

  const edges = useMemo(
    () =>
      systemCausality.map(({ source, target, label }) =>
        makeEdge(source, target, label, danger ? "#ef4444" : "#0f766e"),
      ),
    [danger],
  );

  return (
    <div className="rf-section">
      <div className={`rf-modebar${danger ? " rf-modebar-danger" : ""}`}>
        <span className="rf-modebar-title">换一种搭法看看：</span>
        <label className="rf-switch">
          <span className={!danger ? "rf-switch-opt rf-switch-on" : "rf-switch-opt"}>
            加法优先（推荐）
          </span>
          <input
            type="checkbox"
            checked={danger}
            onChange={(event) => setDanger(event.target.checked)}
            aria-label="切换加法优先或乘法叠加"
          />
          <span className="rf-switch-track" aria-hidden="true">
            <span className="rf-switch-thumb" />
          </span>
          <span
            className={danger ? "rf-switch-opt rf-switch-on rf-switch-opt-danger" : "rf-switch-opt"}
          >
            乘法叠加（危险）
          </span>
        </label>
      </div>
      <DiagramCanvas
        nodes={systemNodes}
        edges={edges}
        idleText="👆 点一下任意环节，看它怎么影响下一个。按上面的开关，还能对比两种搭法有什么不同。"
        miniMapColor={danger ? "#ef4444" : "#0f766e"}
      />
      {danger ? <MultiplicativeView /> : <AdditiveView />}
    </div>
  );
}

export function FlowBlock({ diagram }: { diagram: FlowDiagram }) {
  if (diagram === "value-flow") {
    return <ValueFlowDiagram />;
  }

  return <SystemMapDiagram />;
}
