import { useEffect, useMemo, useState } from "react";
import type { ReactNode } from "react";
import { getPage, groupedPages, neighbours, pages } from "./content";
import type { Block, DocPage, Tone } from "./types";

const toneLabels: Record<Tone, string> = {
  note: "说明",
  tip: "建议",
  warn: "注意",
  danger: "重要",
};

function headingId(index: number): string {
  return `heading-${index}`;
}

/** 正文里支持极少量的行内标记：**加粗** 与 `行内代码`。 */
function renderInline(text: string): ReactNode[] {
  const parts = text.split(/(\*\*[^*]+\*\*|`[^`]+`)/g);

  return parts.map((part, i) => {
    if (part.startsWith("**") && part.endsWith("**") && part.length > 4) {
      return <strong key={i}>{part.slice(2, -2)}</strong>;
    }

    if (part.startsWith("`") && part.endsWith("`") && part.length > 2) {
      return (
        <code key={i} className="inline-code">
          {part.slice(1, -1)}
        </code>
      );
    }

    return part;
  });
}

function CopyButton({ source }: { source: string }) {
  const [copied, setCopied] = useState(false);

  function handleCopy() {
    void navigator.clipboard?.writeText(source).then(() => {
      setCopied(true);
      window.setTimeout(() => setCopied(false), 1500);
    });
  }

  return (
    <button type="button" className="copy-btn" onClick={handleCopy}>
      {copied ? "已复制" : "复制"}
    </button>
  );
}

function BlockView({ block, index }: { block: Block; index: number }) {
  switch (block.type) {
    case "h2":
      return (
        <h2 id={headingId(index)} data-heading className="doc-h2">
          {renderInline(block.text)}
        </h2>
      );
    case "h3":
      return (
        <h3 id={headingId(index)} data-heading className="doc-h3">
          {renderInline(block.text)}
        </h3>
      );
    case "p":
      return <p className="doc-p">{renderInline(block.text)}</p>;
    case "ul":
      return (
        <ul className="doc-list">
          {block.items.map((item, i) => (
            <li key={i}>{renderInline(item)}</li>
          ))}
        </ul>
      );
    case "ol":
      return (
        <ol className="doc-list">
          {block.items.map((item, i) => (
            <li key={i}>{renderInline(item)}</li>
          ))}
        </ol>
      );
    case "code":
      return (
        <figure className="code-block">
          <div className="code-head">
            <span className="code-lang">{block.lang}</span>
            {block.caption ? <span className="code-caption">{block.caption}</span> : null}
            <CopyButton source={block.source} />
          </div>
          <pre>
            <code>{block.source}</code>
          </pre>
        </figure>
      );
    case "table":
      return (
        <figure className="table-block">
          <div className="table-scroll">
            <table>
              <thead>
                <tr>
                  {block.head.map((cell, i) => (
                    <th key={i}>{renderInline(cell)}</th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {block.rows.map((row, ri) => (
                  <tr key={ri}>
                    {row.map((cell, ci) => (
                      <td key={ci}>{renderInline(cell)}</td>
                    ))}
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          {block.caption ? <figcaption>{block.caption}</figcaption> : null}
        </figure>
      );
    case "callout":
      return (
        <aside className={`callout callout-${block.tone}`}>
          <div className="callout-title">
            <span className="callout-badge">{toneLabels[block.tone]}</span>
            <strong>{renderInline(block.title)}</strong>
          </div>
          <p>{renderInline(block.text)}</p>
        </aside>
      );
    case "formula":
      return (
        <div className="formula">
          <code>{block.expression}</code>
          {block.note ? <p>{renderInline(block.note)}</p> : null}
        </div>
      );
    case "steps":
      return (
        <ol className="steps">
          {block.items.map((item, i) => (
            <li key={i}>
              <span className="step-num">{i + 1}</span>
              <div>
                <strong>{renderInline(item.title)}</strong>
                <p>{renderInline(item.text)}</p>
              </div>
            </li>
          ))}
        </ol>
      );
    case "quote":
      return (
        <blockquote className="doc-quote">
          <p>{renderInline(block.text)}</p>
          {block.attribution ? <cite>{renderInline(block.attribution)}</cite> : null}
        </blockquote>
      );
  }
}

function Toc({ page, active }: { page: DocPage; active: string }) {
  const headings = page.blocks
    .map((block, index) => ({ block, index }))
    .filter((entry) => entry.block.type === "h2" || entry.block.type === "h3");

  if (headings.length === 0) {
    return null;
  }

  return (
    <nav className="toc" aria-label="本页目录">
      <div className="toc-title">本页目录</div>
      <ul>
        {headings.map(({ block, index }) => {
          const text = block.type === "h2" || block.type === "h3" ? block.text : "";
          const id = headingId(index);

          return (
            <li key={id} className={block.type === "h3" ? "toc-sub" : ""}>
              <a
                href={`#${id}`}
                className={active === id ? "toc-active" : ""}
                onClick={(event) => {
                  event.preventDefault();
                  document.getElementById(id)?.scrollIntoView({ behavior: "smooth" });
                  window.history.replaceState(null, "", `#${id}`);
                }}
              >
                {text}
              </a>
            </li>
          );
        })}
      </ul>
    </nav>
  );
}

function readHash(): string {
  const raw = window.location.hash.replace(/^#\/?/, "");

  return getPage(raw) ? raw : pages[0].id;
}

export default function App() {
  const [pageId, setPageId] = useState<string>(readHash);
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const [activeHeading, setActiveHeading] = useState("");

  const page = getPage(pageId) ?? pages[0];
  const groups = useMemo(() => groupedPages(), []);
  const { prev, next } = useMemo(() => neighbours(page.id), [page.id]);

  useEffect(() => {
    function onHashChange() {
      const id = readHash();
      setPageId(id);
      setSidebarOpen(false);
      window.scrollTo(0, 0);
      setActiveHeading("");
    }

    window.addEventListener("hashchange", onHashChange);
    return () => window.removeEventListener("hashchange", onHashChange);
  }, []);

  useEffect(() => {
    function onScroll() {
      const headings = Array.from(document.querySelectorAll<HTMLElement>("[data-heading]"));
      let current = "";

      for (const heading of headings) {
        if (heading.getBoundingClientRect().top <= 140) {
          current = heading.id;
        }
      }

      setActiveHeading(current);
    }

    window.addEventListener("scroll", onScroll, { passive: true });
    return () => window.removeEventListener("scroll", onScroll);
  }, [pageId]);

  function navigate(id: string) {
    window.location.hash = `#/${id}`;
  }

  return (
    <div className="app-shell">
      <header className="topbar">
        <button
          type="button"
          className="menu-toggle"
          aria-label="切换侧边栏"
          onClick={() => setSidebarOpen((open) => !open)}
        >
          ☰
        </button>
        <button type="button" className="brand" onClick={() => navigate(pages[0].id)}>
          <span className="brand-mark">道</span>
          <span>
            <strong>大道归一</strong>
            <small>游戏数值与长线运营手册</small>
          </span>
        </button>
        <span className="topbar-tag">companion docs</span>
      </header>

      <div className="layout">
        <aside className={`sidebar ${sidebarOpen ? "sidebar-open" : ""}`}>
          <div className="sidebar-inner">
            {groups.map((group) => (
              <div className="nav-group" key={group.section}>
                <div className="nav-group-title">{group.section}</div>
                <ul>
                  {group.pages.map((item) => (
                    <li key={item.id}>
                      <button
                        type="button"
                        className={`nav-link ${item.id === page.id ? "nav-active" : ""}`}
                        onClick={() => navigate(item.id)}
                      >
                        {item.nav}
                      </button>
                    </li>
                  ))}
                </ul>
              </div>
            ))}
            <div className="sidebar-foot">数值是体检报告，不是命运判词。</div>
          </div>
        </aside>

        {sidebarOpen ? (
          <button
            type="button"
            className="sidebar-backdrop"
            aria-label="关闭侧边栏"
            onClick={() => setSidebarOpen(false)}
          />
        ) : null}

        <main className="main">
          <article className="doc">
            <div className="doc-breadcrumb">
              <span>{page.section}</span>
              <span className="crumb-sep">/</span>
              <span>{page.nav}</span>
              <span className="reading-time">约 {page.readingTime}</span>
            </div>
            <h1 className="doc-title">{page.title}</h1>
            <p className="doc-subtitle">{page.subtitle}</p>

            <div className="doc-body">
              {page.blocks.map((block, index) => (
                <BlockView key={index} block={block} index={index} />
              ))}
            </div>

            <nav className="page-nav" aria-label="页面导航">
              {prev ? (
                <button type="button" className="page-nav-btn" onClick={() => navigate(prev.id)}>
                  <small>← 上一篇</small>
                  <span>{prev.nav}</span>
                </button>
              ) : (
                <span />
              )}
              {next ? (
                <button
                  type="button"
                  className="page-nav-btn page-nav-next"
                  onClick={() => navigate(next.id)}
                >
                  <small>下一篇 →</small>
                  <span>{next.nav}</span>
                </button>
              ) : (
                <span />
              )}
            </nav>
          </article>

          <Toc page={page} active={activeHeading} />
        </main>
      </div>
    </div>
  );
}
