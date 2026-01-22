package render

import (
	"html/template"
	"io"

	"github.com/beardnick/go-test-coverage/internal/report"
)

const reportTemplate = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>{{.Title}}</title>
  <style id="highlight-dark">{{.HighlightDarkCSS}}</style>
  <style id="highlight-light" disabled>{{.HighlightLightCSS}}</style>
  <style>
    :root {
      color-scheme: dark;
      --bg: #1f2937;
      --panel: #273449;
      --panel-border: #3b4758;
      --text: #e2e8f0;
      --muted: #b0bac6;
      --accent: #58a6ff;
      --covered: #3fb950;
      --missed: #f85149;
      --partial: #d29922;
      --not-tracked: #6e7681;
      --code-bg: #1b2330;
      --sidebar-bg: #1b2330;
      --header-bg: #273449;
      --progress-track: #334155;
      --input-bg: #1b2330;
      --code-line-bg: #1b2330;
      --legend-not-tracked: #c9d1d9;
      --legend-missed: #fca5a5;
      --legend-partial: #f5d481;
      --legend-covered: #7ee787;
      --partial-range: rgba(248, 81, 73, 0.35);
    }

    body.theme-light {
      color-scheme: light;
      --bg: #f8fafc;
      --panel: #ffffff;
      --panel-border: #e2e8f0;
      --text: #0f172a;
      --muted: #64748b;
      --accent: #2563eb;
      --covered: #16a34a;
      --missed: #dc2626;
      --partial: #d97706;
      --not-tracked: #94a3b8;
      --code-bg: #f1f5f9;
      --sidebar-bg: #eef2f7;
      --header-bg: #f1f5f9;
      --progress-track: #e2e8f0;
      --input-bg: #ffffff;
      --code-line-bg: #e7edf4;
      --legend-not-tracked: #475569;
      --legend-missed: #b91c1c;
      --legend-partial: #b45309;
      --legend-covered: #15803d;
      --partial-range: rgba(220, 38, 38, 0.22);
    }

    * {
      box-sizing: border-box;
    }

    body {
      margin: 0;
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", "Noto Sans", Helvetica, Arial, sans-serif;
      font-size: 14px;
      line-height: 1.5;
      background: var(--bg);
      color: var(--text);
    }

    a {
      color: var(--accent);
      text-decoration: none;
    }

    a:hover {
      text-decoration: underline;
    }

    .app {
      display: flex;
      min-height: 100vh;
      background: var(--bg);
    }

    .sidebar {
      width: 280px;
      background: var(--sidebar-bg);
      border-right: 1px solid var(--panel-border);
      display: flex;
      flex-direction: column;
    }

    .sidebar-header {
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 12px;
      padding: 16px 18px 12px;
      font-size: 12px;
      letter-spacing: 0.12em;
      text-transform: uppercase;
      color: var(--muted);
      border-bottom: 1px solid var(--panel-border);
    }

    .sidebar-toggle {
      border: 1px solid var(--panel-border);
      background: transparent;
      color: var(--muted);
      font-size: 11px;
      padding: 4px 10px;
      border-radius: 999px;
      cursor: pointer;
      text-transform: none;
      letter-spacing: normal;
    }

    .sidebar-toggle:hover {
      background: rgba(88, 166, 255, 0.08);
      color: var(--text);
    }

    .sidebar-toggle:focus-visible {
      outline: 2px solid var(--accent);
      outline-offset: 2px;
    }

    .file-tree {
      list-style: none;
      margin: 0;
      padding: 8px 0;
      overflow: auto;
      flex: 1;
    }

    .file-node {
      margin: 0;
      padding: 0;
    }

    .tree-dir details {
      padding: 0;
    }

    .tree-dir summary {
      list-style: none;
      display: flex;
      align-items: center;
      gap: 6px;
      padding: 6px 16px 6px 12px;
      cursor: pointer;
      color: var(--muted);
      font-size: 12px;
      user-select: none;
    }

    .tree-dir summary::-webkit-details-marker {
      display: none;
    }

    .tree-arrow {
      display: inline-flex;
      width: 12px;
      height: 12px;
      align-items: center;
      justify-content: center;
      transition: transform 0.2s ease;
      color: var(--muted);
    }

    .tree-arrow::before {
      content: '▸';
      font-size: 12px;
    }

    .tree-dir details[open] .tree-arrow {
      transform: rotate(90deg);
    }

    .tree-label {
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
      flex: 1;
    }

    .tree-coverage {
      margin-left: auto;
      font-size: 10px;
      font-weight: 600;
      padding: 2px 8px;
      border-radius: 999px;
      border: 1px solid var(--panel-border);
      color: var(--muted);
      white-space: nowrap;
    }

    .tree-coverage.high {
      color: var(--covered);
      border-color: rgba(63, 185, 80, 0.5);
    }

    .tree-coverage.medium {
      color: var(--partial);
      border-color: rgba(210, 153, 34, 0.5);
    }

    .tree-coverage.low {
      color: var(--missed);
      border-color: rgba(248, 81, 73, 0.5);
    }

    .tree-coverage.none {
      color: var(--muted);
      border-color: rgba(110, 118, 129, 0.5);
    }

    .tree-children {
      list-style: none;
      margin: 0;
      padding: 0 0 0 14px;
    }

    .file-node button {
      width: 100%;
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 10px;
      padding: 6px 16px 6px 22px;
      background: transparent;
      border: none;
      color: var(--text);
      cursor: pointer;
      text-align: left;
      border-left: 3px solid transparent;
      font-family: inherit;
      font-size: 13px;
    }

    .file-node button:hover {
      background: rgba(88, 166, 255, 0.08);
    }

    .file-node.active button {
      background: rgba(88, 166, 255, 0.15);
      border-left-color: var(--accent);
    }

    .file-label {
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
      flex: 1;
    }

    .file-coverage {
      font-size: 11px;
      font-weight: 600;
      padding: 2px 8px;
      border-radius: 999px;
      border: 1px solid var(--panel-border);
      color: var(--muted);
      white-space: nowrap;
    }

    .file-coverage.high {
      color: var(--covered);
      border-color: rgba(63, 185, 80, 0.5);
    }

    .file-coverage.medium {
      color: var(--partial);
      border-color: rgba(210, 153, 34, 0.5);
    }

    .file-coverage.low {
      color: var(--missed);
      border-color: rgba(248, 81, 73, 0.5);
    }

    .file-coverage.none {
      color: var(--muted);
      border-color: rgba(110, 118, 129, 0.5);
    }

    .main {
      flex: 1;
      min-width: 0;
    }

    .container {
      max-width: none;
      margin: 0;
      padding: 24px 32px 48px;
    }

    .page-actions {
      display: flex;
      align-items: center;
      justify-content: flex-start;
      margin-bottom: 12px;
    }

    .page-header {
      display: flex;
      flex-wrap: wrap;
      justify-content: space-between;
      align-items: flex-end;
      gap: 16px;
      margin-bottom: 24px;
    }

    .page-header h1 {
      margin: 0;
      font-size: 26px;
      font-weight: 600;
      letter-spacing: -0.01em;
    }

    .page-header p {
      margin: 6px 0 0;
      color: var(--muted);
    }

    .summary-inline {
      display: flex;
      flex-wrap: wrap;
      gap: 16px;
      background: var(--panel);
      border: 1px solid var(--panel-border);
      padding: 12px 16px;
      border-radius: 8px;
    }

    .summary-item {
      display: flex;
      flex-direction: column;
      gap: 4px;
      min-width: 140px;
    }

    .summary-item .label {
      font-size: 11px;
      letter-spacing: 0.12em;
      text-transform: uppercase;
      color: var(--muted);
    }

    .summary-item .value {
      font-size: 18px;
      font-weight: 600;
    }

    .summary-grid {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
      gap: 16px;
      margin-bottom: 28px;
    }

    .card {
      background: var(--panel);
      border: 1px solid var(--panel-border);
      padding: 16px;
      border-radius: 8px;
      display: flex;
      flex-direction: column;
      gap: 10px;
    }

    .card .label {
      color: var(--muted);
      font-size: 11px;
      letter-spacing: 0.08em;
      text-transform: uppercase;
    }

    .card .value {
      font-size: 24px;
      font-weight: 600;
    }

    .progress {
      width: 100%;
      height: 6px;
      background: var(--progress-track);
      border-radius: 999px;
      overflow: hidden;
    }

    .progress .bar {
      height: 100%;
      border-radius: inherit;
    }

    .high {
      color: var(--covered);
    }

    .medium {
      color: var(--partial);
    }

    .low {
      color: var(--missed);
    }

    .none {
      color: var(--muted);
    }

    .bar.high {
      background: var(--covered);
    }

    .bar.medium {
      background: var(--partial);
    }

    .bar.low {
      background: var(--missed);
    }

    .bar.none {
      background: var(--not-tracked);
    }

    .file-table {
      width: 100%;
      border-collapse: separate;
      border-spacing: 0;
      margin-bottom: 32px;
      background: var(--panel);
      border-radius: 8px;
      overflow: hidden;
      border: 1px solid var(--panel-border);
    }

    .file-table th,
    .file-table td {
      text-align: left;
      padding: 10px 16px;
    }

    .file-table th {
      font-size: 11px;
      color: var(--muted);
      text-transform: uppercase;
      letter-spacing: 0.08em;
      background: var(--header-bg);
      border-bottom: 1px solid var(--panel-border);
    }

    .file-table tr + tr td {
      border-top: 1px solid var(--panel-border);
    }

    .file-name {
      font-weight: 600;
    }

    .viewer {
      background: var(--panel);
      border: 1px solid var(--panel-border);
      border-radius: 8px;
      overflow: hidden;
    }

    .viewer-bar {
      display: flex;
      flex-wrap: wrap;
      gap: 16px;
      align-items: center;
      padding: 12px 16px;
      background: var(--header-bg);
      border-bottom: 1px solid var(--panel-border);
      position: sticky;
      top: 0;
      z-index: 2;
    }

    .current-file {
      font-size: 13px;
      font-weight: 600;
      padding: 4px 10px;
      border-radius: 6px;
      border: 1px solid var(--panel-border);
      background: var(--input-bg);
      color: var(--text);
      max-width: 420px;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }

    .file-picker {
      display: flex;
      flex-direction: column;
      gap: 6px;
      min-width: 260px;
    }

    .file-picker label {
      font-size: 11px;
      letter-spacing: 0.12em;
      text-transform: uppercase;
      color: var(--muted);
    }

    .file-picker select {
      background: var(--input-bg);
      color: var(--text);
      border: 1px solid var(--panel-border);
      padding: 6px 10px;
      border-radius: 6px;
      font-size: 13px;
    }

    .file-picker select:focus {
      outline: 2px solid var(--accent);
      outline-offset: 1px;
    }

    .legend {
      display: flex;
      flex-wrap: wrap;
      gap: 8px;
      align-items: center;
    }

    .legend-item {
      padding: 2px 10px;
      border-radius: 999px;
      font-size: 12px;
      font-weight: 600;
      border: 1px solid var(--panel-border);
      color: var(--muted);
      background: transparent;
    }

    .legend-item.not-tracked {
      border-color: rgba(110, 118, 129, 0.6);
      color: var(--legend-not-tracked);
    }

    .legend-item.missed {
      border-color: rgba(248, 81, 73, 0.6);
      color: var(--legend-missed);
    }

    .legend-item.partial {
      border-color: rgba(210, 153, 34, 0.7);
      color: var(--legend-partial);
    }

    .legend-item.covered {
      border-color: rgba(63, 185, 80, 0.7);
      color: var(--legend-covered);
    }

    .viewer-actions {
      display: flex;
      flex-wrap: wrap;
      gap: 12px;
      align-items: center;
      margin-left: auto;
    }

    .filters {
      display: flex;
      flex-wrap: wrap;
      gap: 12px;
      align-items: center;
    }

    .filter {
      display: flex;
      align-items: center;
      gap: 6px;
      font-size: 12px;
      color: var(--muted);
    }

    .filter input {
      accent-color: var(--accent);
    }

    .theme-toggle {
      border: 1px solid var(--panel-border);
      background: transparent;
      color: var(--text);
      font-size: 12px;
      padding: 6px 10px;
      border-radius: 999px;
      cursor: pointer;
    }

    .theme-toggle:hover {
      background: rgba(88, 166, 255, 0.08);
    }

    .theme-toggle:focus-visible {
      outline: 2px solid var(--accent);
      outline-offset: 2px;
    }

    .viewer-body {
      padding: 0 16px 16px;
    }

    .file-section {
      display: none;
      margin-bottom: 24px;
      padding: 16px 0 8px;
    }

    .file-section.active {
      display: block;
    }

    .file-header {
      display: flex;
      flex-wrap: wrap;
      justify-content: space-between;
      align-items: center;
      gap: 12px;
      margin-bottom: 12px;
    }

    .file-header h2 {
      margin: 0;
      font-size: 14px;
      font-weight: 600;
      word-break: break-all;
    }

    .pill {
      padding: 2px 10px;
      border-radius: 999px;
      font-size: 12px;
      font-weight: 600;
      border: 1px solid transparent;
    }

    .pill.high {
      background: rgba(63, 185, 80, 0.15);
      color: var(--covered);
      border-color: rgba(63, 185, 80, 0.4);
    }

    .pill.medium {
      background: rgba(210, 153, 34, 0.15);
      color: var(--partial);
      border-color: rgba(210, 153, 34, 0.4);
    }

    .pill.low {
      background: rgba(248, 81, 73, 0.15);
      color: var(--missed);
      border-color: rgba(248, 81, 73, 0.4);
    }

    .pill.none {
      background: rgba(110, 118, 129, 0.12);
      color: var(--muted);
      border-color: rgba(110, 118, 129, 0.4);
    }

    .code-table {
      width: 100%;
      border-collapse: collapse;
      font-family: SFMono-Regular, Consolas, "Liberation Mono", Menlo, monospace;
      font-size: 12.5px;
      margin-top: 12px;
      border-radius: 6px;
      overflow: hidden;
      background: var(--code-bg);
      border: 1px solid var(--panel-border);
      tab-size: 4;
      -moz-tab-size: 4;
    }

    .code-table td {
      padding: 0 12px;
      vertical-align: top;
      line-height: 20px;
    }

    .code-table .line-no {
      width: 60px;
      text-align: right;
      color: var(--muted);
      border-right: 1px solid var(--panel-border);
      background: var(--code-line-bg);
      user-select: none;
    }

    .code-table .code {
      white-space: pre;
    }

    .code-table .code code.hljs {
      display: block;
      padding: 0;
      background: transparent;
    }

    .code-table tr.covered td.code {
      background: rgba(63, 185, 80, 0.16);
    }

    .code-table tr.missed td.code {
      background: rgba(248, 81, 73, 0.16);
    }

    .code-table tr.partial td.code {
      background: rgba(210, 153, 34, 0.16);
    }

    .partial-range {
      background: var(--partial-range);
      border-radius: 3px;
    }

    .code-table tr.not-tracked td.code {
      background: rgba(110, 118, 129, 0.08);
    }

    .code-table tr.covered .line-no {
      color: var(--covered);
    }

    .code-table tr.missed .line-no {
      color: var(--missed);
    }

    .code-table tr.partial .line-no {
      color: var(--partial);
    }

    .missing {
      padding: 12px;
      border-radius: 6px;
      background: rgba(248, 81, 73, 0.08);
      color: #fca5a5;
      border: 1px solid rgba(248, 81, 73, 0.35);
      margin-top: 12px;
    }

    .footer {
      color: var(--muted);
      font-size: 12px;
      margin-top: 24px;
    }

    body.hide-not-tracked tr.not-tracked {
      display: none;
    }

    body.hide-missed tr.missed {
      display: none;
    }

    body.hide-partial tr.partial {
      display: none;
    }

    body.hide-covered tr.covered {
      display: none;
    }
  </style>
</head>
<body>
  {{define "tree"}}
    {{range .}}
      {{if .IsDir}}
        <li class="tree-dir">
          <details>
            <summary>
              <span class="tree-arrow"></span>
              <span class="tree-label">{{.Name}}</span>
              <span class="tree-coverage {{.CoverageClass}}">{{.CoveragePercent}}</span>
            </summary>
            <ul class="tree-children">
              {{template "tree" .Children}}
            </ul>
          </details>
        </li>
      {{else}}
        <li class="file-node" data-anchor="{{.Anchor}}" data-name="{{.RelativePath}}" data-coverage="{{.CoveragePercent}}">
          <button type="button">
            <span class="file-label">{{.Name}}</span>
            <span class="file-coverage {{.CoverageClass}}">{{.CoveragePercent}}</span>
          </button>
        </li>
      {{end}}
    {{end}}
  {{end}}
  <div class="app">
    <aside class="sidebar">
      <div class="sidebar-header">
        <span>Files</span>
        <button type="button" class="sidebar-toggle" id="toggle-tree" aria-expanded="false">Expand all</button>
      </div>
      <ul class="file-tree">
        {{template "tree" .Tree}}
      </ul>
    </aside>
    <main class="main">
      <div class="container">
        <div class="page-actions">
          <button type="button" class="theme-toggle" id="theme-toggle" aria-pressed="false">Light theme</button>
        </div>
        <header class="page-header">
          <div>
            <h1>{{.Title}}</h1>
            <p>Generated {{.GeneratedAt}}</p>
          </div>
          <div class="summary-inline">
            <div class="summary-item">
              <span class="label">Total Coverage</span>
              <span class="value {{.TotalCoverageClass}}">{{.TotalCoveragePercent}}</span>
            </div>
            <div class="summary-item">
              <span class="label">Statements</span>
              <span class="value">{{.CoveredStmts}} / {{.TotalStmts}}</span>
            </div>
            <div class="summary-item">
              <span class="label">Files</span>
              <span class="value">{{.TotalFiles}}</span>
              <span class="label">Missing {{.MissingFiles}}</span>
            </div>
          </div>
        </header>

    <section class="summary-grid">
      <div class="card">
        <div class="label">Total Coverage</div>
        <div class="value {{.TotalCoverageClass}}">{{.TotalCoveragePercent}}</div>
        <div>Covered {{.CoveredStmts}} / {{.TotalStmts}} statements</div>
        <div class="progress">
          <div class="bar {{.TotalCoverageClass}}" style="width: {{.TotalCoveragePercent}};"></div>
        </div>
      </div>
      <div class="card">
        <div class="label">Files</div>
        <div class="value">{{.TotalFiles}}</div>
        <div>Missing sources: {{.MissingFiles}}</div>
      </div>
      <div class="card">
        <div class="label">Legend</div>
        <div>Covered</div>
        <div>Partial</div>
        <div>Not covered</div>
        <div>Not tracked</div>
      </div>
    </section>

    <section class="viewer">
      <div class="viewer-bar">
        <div class="current-file" id="current-file"></div>
        <div class="legend">
          <span class="legend-item not-tracked">not tracked</span>
          <span class="legend-item missed">not covered</span>
          <span class="legend-item partial">partial</span>
          <span class="legend-item covered">covered</span>
        </div>
        <div class="viewer-actions">
          <div class="filters">
            <label class="filter"><input type="checkbox" data-filter="not-tracked" checked> not tracked</label>
            <label class="filter"><input type="checkbox" data-filter="missed" checked> not covered</label>
            <label class="filter"><input type="checkbox" data-filter="partial" checked> partial</label>
            <label class="filter"><input type="checkbox" data-filter="covered" checked> covered</label>
          </div>
        </div>
      </div>
      <div class="viewer-body">
        {{range .Files}}
        <section class="file-section" id="{{.Anchor}}">
          <div class="file-header">
            <h2>{{.Name}}</h2>
            <span class="pill {{.CoverageClass}}">{{.CoveragePercent}}</span>
          </div>
          <div>Covered {{.CoveredStmts}} / {{.TotalStmts}} statements</div>
          <div class="progress" style="margin-top: 8px;">
            <div class="bar {{.CoverageClass}}" style="width: {{.CoveragePercent}};"></div>
          </div>
          {{if .Missing}}
          <div class="missing">{{.MissingDescription}}</div>
          {{else}}
          <table class="code-table">
            <tbody>
              {{range .Lines}}
              <tr class="{{.Class}}">
                <td class="line-no">{{.Number}}</td>
                <td class="code"{{if .Ranges}} data-partial="{{.Ranges}}"{{end}}><code class="hljs language-go">{{.Code}}</code></td>
              </tr>
              {{end}}
            </tbody>
          </table>
          {{end}}
        </section>
        {{end}}
      </div>
    </section>

    <div class="footer">Generated by beautiful-coverage.</div>
      </div>
    </main>
  </div>
  <script>{{.HighlightJS}}</script>
  <script>{{.HighlightGoJS}}</script>
  <script>
    const sections = Array.from(document.querySelectorAll('.file-section'));
    const filters = document.querySelectorAll('[data-filter]');
    const fileNodes = Array.from(document.querySelectorAll('.file-node'));
    const treeDetails = Array.from(document.querySelectorAll('.tree-dir details'));
    const treeToggle = document.getElementById('toggle-tree');
    const currentFile = document.getElementById('current-file');
    const codeBlocks = document.querySelectorAll('.code-table code');
    const themeToggle = document.getElementById('theme-toggle');
    const highlightDark = document.getElementById('highlight-dark');
    const highlightLight = document.getElementById('highlight-light');

    if (window.hljs) {
      codeBlocks.forEach((block) => {
        hljs.highlightElement(block);
      });
    }
    applyPartialRanges();

    function parseRanges(value) {
      if (!value) {
        return [];
      }
      return value.split(',').map((segment) => {
        const parts = segment.split('-').map((part) => Number(part));
        if (parts.length !== 2) {
          return null;
        }
        const start = parts[0];
        const end = parts[1];
        if (!Number.isFinite(start) || !Number.isFinite(end) || end <= start) {
          return null;
        }
        return { start, end };
      }).filter(Boolean);
    }

    function wrapNodeRanges(node, ranges) {
      const text = node.nodeValue || '';
      if (!text || ranges.length === 0) {
        return;
      }
      const fragment = document.createDocumentFragment();
      let cursor = 0;
      ranges.forEach((range) => {
        if (range.start > cursor) {
          fragment.appendChild(document.createTextNode(text.slice(cursor, range.start)));
        }
        const span = document.createElement('span');
        span.className = 'partial-range';
        span.textContent = text.slice(range.start, range.end);
        fragment.appendChild(span);
        cursor = range.end;
      });
      if (cursor < text.length) {
        fragment.appendChild(document.createTextNode(text.slice(cursor)));
      }
      node.replaceWith(fragment);
    }

    function applyPartialRanges() {
      const cells = document.querySelectorAll('td.code[data-partial]');
      cells.forEach((cell) => {
        const ranges = parseRanges(cell.dataset.partial);
        if (ranges.length === 0) {
          return;
        }
        const code = cell.querySelector('code');
        if (!code) {
          return;
        }
        const textNodes = [];
        const walker = document.createTreeWalker(code, NodeFilter.SHOW_TEXT);
        while (walker.nextNode()) {
          textNodes.push(walker.currentNode);
        }
        let offset = 1;
        textNodes.forEach((node) => {
          const length = (node.nodeValue || '').length;
          const nodeStart = offset;
          const nodeEnd = offset + length;
          const nodeRanges = ranges.map((range) => {
            const start = Math.max(range.start, nodeStart);
            const end = Math.min(range.end, nodeEnd);
            if (end <= start) {
              return null;
            }
            return { start: start - nodeStart, end: end - nodeStart };
          }).filter(Boolean);
          if (nodeRanges.length > 0) {
            wrapNodeRanges(node, nodeRanges);
          }
          offset = nodeEnd;
        });
      });
    }

    function applyTheme(theme) {
      const useLight = theme === 'light';
      document.body.classList.toggle('theme-light', useLight);
      if (highlightDark && highlightLight) {
        highlightDark.disabled = useLight;
        highlightLight.disabled = !useLight;
      }
      if (themeToggle) {
        themeToggle.textContent = useLight ? 'Dark theme' : 'Light theme';
        themeToggle.setAttribute('aria-pressed', useLight ? 'true' : 'false');
      }
      try {
        localStorage.setItem('theme', theme);
      } catch (err) {
        // Ignore storage failures (private mode, etc.).
      }
    }

    function initTheme() {
      let theme = 'dark';
      try {
        const stored = localStorage.getItem('theme');
        if (stored === 'light' || stored === 'dark') {
          theme = stored;
        } else if (window.matchMedia && window.matchMedia('(prefers-color-scheme: light)').matches) {
          theme = 'light';
        }
      } catch (err) {
        // Ignore storage failures and fall back to default theme.
      }
      applyTheme(theme);
    }

    function updateTreeToggleLabel() {
      if (!treeToggle) {
        return;
      }
      const allOpen = treeDetails.length > 0 && treeDetails.every((item) => item.open);
      treeToggle.textContent = allOpen ? 'Collapse all' : 'Expand all';
      treeToggle.setAttribute('aria-expanded', allOpen ? 'true' : 'false');
      treeToggle.disabled = treeDetails.length === 0;
    }

    function setAllTree(open) {
      treeDetails.forEach((item) => {
        item.open = open;
      });
    }

    function hasSection(anchor) {
      return sections.some((section) => section.id === anchor);
    }

    function setCurrent(anchor) {
      const node = fileNodes.find((item) => item.dataset.anchor === anchor);
      if (!node) {
        return;
      }
      const name = node.dataset.name || anchor;
      const coverage = node.dataset.coverage;
      if (currentFile) {
        currentFile.textContent = coverage ? name + ' (' + coverage + ')' : name;
      }
      fileNodes.forEach((item) => {
        item.classList.toggle('active', item.dataset.anchor === anchor);
      });
    }

    function activate(anchor, updateHash) {
      sections.forEach((section) => {
        section.classList.toggle('active', section.id === anchor);
      });
      setCurrent(anchor);
      if (updateHash) {
        history.replaceState(null, '', '#' + anchor);
      }
    }

    function syncFromHash() {
      const hash = window.location.hash.replace('#', '');
      if (hash && hasSection(hash)) {
        activate(hash, false);
        return;
      }
      if (sections.length > 0) {
        activate(sections[0].id, false);
      }
    }

    if (sections.length > 0) {
      syncFromHash();
      window.addEventListener('hashchange', syncFromHash);
    }

    if (treeToggle) {
      treeToggle.addEventListener('click', () => {
        const allOpen = treeDetails.length > 0 && treeDetails.every((item) => item.open);
        setAllTree(!allOpen);
        updateTreeToggleLabel();
      });
      treeDetails.forEach((item) => {
        item.addEventListener('toggle', updateTreeToggleLabel);
      });
      updateTreeToggleLabel();
    }

    if (themeToggle) {
      themeToggle.addEventListener('click', () => {
        const isLight = document.body.classList.contains('theme-light');
        applyTheme(isLight ? 'dark' : 'light');
      });
    }

    initTheme();

    fileNodes.forEach((node) => {
      node.addEventListener('click', () => {
        activate(node.dataset.anchor, true);
      });
    });

    filters.forEach((filter) => {
      filter.addEventListener('change', (event) => {
        const key = event.target.getAttribute('data-filter');
        document.body.classList.toggle('hide-' + key, !event.target.checked);
      });
    });
  </script>
</body>
</html>
`

func HTML(writer io.Writer, reportData report.Report) error {
	assets, err := LoadInlineAssets()
	if err != nil {
		return err
	}

	tmpl, err := template.New("report").Parse(reportTemplate)
	if err != nil {
		return err
	}

	data := struct {
		report.Report
		HighlightDarkCSS  template.CSS
		HighlightLightCSS template.CSS
		HighlightJS       template.JS
		HighlightGoJS     template.JS
	}{
		Report:            reportData,
		HighlightDarkCSS:  template.CSS(assets.HighlightDarkCSS),
		HighlightLightCSS: template.CSS(assets.HighlightLightCSS),
		HighlightJS:       template.JS(assets.HighlightJS),
		HighlightGoJS:     template.JS(assets.HighlightGoJS),
	}

	return tmpl.Execute(writer, data)
}
