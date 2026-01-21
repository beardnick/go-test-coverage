package render

import (
	"html/template"
	"io"

	"go-test-coverage/internal/report"
)

const reportTemplate = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>{{.Title}}</title>
  <link rel="stylesheet" href="{{.AssetsPath}}/highlight/github-dark.min.css" media="(prefers-color-scheme: dark)">
  <link rel="stylesheet" href="{{.AssetsPath}}/highlight/github.min.css" media="(prefers-color-scheme: light)">
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
      background: #1b2330;
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
      background: #334155;
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
      background: #273449;
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
      background: #273449;
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
      background: #1b2330;
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
      background: #1b2330;
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
      color: #c9d1d9;
    }

    .legend-item.missed {
      border-color: rgba(248, 81, 73, 0.6);
      color: #fca5a5;
    }

    .legend-item.partial {
      border-color: rgba(210, 153, 34, 0.7);
      color: #f5d481;
    }

    .legend-item.covered {
      border-color: rgba(63, 185, 80, 0.7);
      color: #7ee787;
    }

    .filters {
      display: flex;
      flex-wrap: wrap;
      gap: 12px;
      align-items: center;
      margin-left: auto;
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
      background: #1b2330;
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

    .code-table tr.not-tracked td.code {
      background: rgba(110, 118, 129, 0.08);
    }

    .code-table tr.covered .line-no {
      color: #7ee787;
    }

    .code-table tr.missed .line-no {
      color: #fca5a5;
    }

    .code-table tr.partial .line-no {
      color: #f5d481;
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
          <details open>
            <summary>
              <span class="tree-arrow"></span>
              <span class="tree-label">{{.Name}}</span>
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
        <button type="button" class="sidebar-toggle" id="toggle-tree" aria-expanded="true">Collapse all</button>
      </div>
      <ul class="file-tree">
        {{template "tree" .Tree}}
      </ul>
    </aside>
    <main class="main">
      <div class="container">
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
        <div class="filters">
          <label class="filter"><input type="checkbox" data-filter="not-tracked" checked> not tracked</label>
          <label class="filter"><input type="checkbox" data-filter="missed" checked> not covered</label>
          <label class="filter"><input type="checkbox" data-filter="partial" checked> partial</label>
          <label class="filter"><input type="checkbox" data-filter="covered" checked> covered</label>
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
                <td class="code"><code class="hljs language-go">{{.Code}}</code></td>
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
  <script src="{{.AssetsPath}}/highlight/highlight.min.js"></script>
  <script src="{{.AssetsPath}}/highlight/go.min.js"></script>
  <script>
    const sections = Array.from(document.querySelectorAll('.file-section'));
    const filters = document.querySelectorAll('[data-filter]');
    const fileNodes = Array.from(document.querySelectorAll('.file-node'));
    const treeDetails = Array.from(document.querySelectorAll('.tree-dir details'));
    const treeToggle = document.getElementById('toggle-tree');
    const currentFile = document.getElementById('current-file');
    const codeBlocks = document.querySelectorAll('.code-table code');

    if (window.hljs) {
      codeBlocks.forEach((block) => {
        hljs.highlightElement(block);
      });
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
      let parent = node.parentElement;
      while (parent) {
        if (parent.tagName === 'DETAILS') {
          parent.open = true;
        }
        parent = parent.parentElement;
      }
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
	tmpl, err := template.New("report").Parse(reportTemplate)
	if err != nil {
		return err
	}
	return tmpl.Execute(writer, reportData)
}
