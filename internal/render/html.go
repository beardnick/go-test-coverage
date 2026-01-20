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
  <style>
    :root {
      color-scheme: dark;
      --bg: #0b1120;
      --panel: #0f172a;
      --panel-border: #1f2937;
      --text: #e5e7eb;
      --muted: #94a3b8;
      --accent: #38bdf8;
      --covered: #22c55e;
      --missed: #ef4444;
      --partial: #f59e0b;
      --not-tracked: #475569;
    }

    * {
      box-sizing: border-box;
    }

    body {
      margin: 0;
      font-family: "Inter", "Segoe UI", system-ui, sans-serif;
      background: radial-gradient(circle at top, #1f2937 0%, #0b1120 55%);
      color: var(--text);
    }

    a {
      color: inherit;
      text-decoration: none;
    }

    .container {
      max-width: 1200px;
      margin: 0 auto;
      padding: 32px 24px 64px;
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
      font-size: 32px;
      letter-spacing: -0.02em;
    }

    .page-header p {
      margin: 6px 0 0;
      color: var(--muted);
    }

    .summary-inline {
      display: flex;
      flex-wrap: wrap;
      gap: 16px;
      background: rgba(15, 23, 42, 0.65);
      border: 1px solid var(--panel-border);
      padding: 12px 16px;
      border-radius: 16px;
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
      font-size: 20px;
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
      border-radius: 16px;
      display: flex;
      flex-direction: column;
      gap: 12px;
      box-shadow: 0 18px 38px rgba(15, 23, 42, 0.4);
    }

    .card .label {
      color: var(--muted);
      font-size: 12px;
      letter-spacing: 0.08em;
      text-transform: uppercase;
    }

    .card .value {
      font-size: 28px;
      font-weight: 600;
    }

    .progress {
      width: 100%;
      height: 8px;
      background: #1f2937;
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
      background: linear-gradient(90deg, #22c55e, #16a34a);
    }

    .bar.medium {
      background: linear-gradient(90deg, #f59e0b, #f97316);
    }

    .bar.low {
      background: linear-gradient(90deg, #f97316, #ef4444);
    }

    .bar.none {
      background: var(--not-tracked);
    }

    .file-table {
      width: 100%;
      border-collapse: collapse;
      margin-bottom: 32px;
      background: var(--panel);
      border-radius: 16px;
      overflow: hidden;
      border: 1px solid var(--panel-border);
    }

    .file-table th,
    .file-table td {
      text-align: left;
      padding: 12px 16px;
    }

    .file-table th {
      font-size: 12px;
      color: var(--muted);
      text-transform: uppercase;
      letter-spacing: 0.08em;
      background: #0b1224;
    }

    .file-table tr + tr td {
      border-top: 1px solid #1f2937;
    }

    .file-name {
      font-weight: 600;
    }

    .viewer {
      background: rgba(15, 23, 42, 0.65);
      border: 1px solid var(--panel-border);
      border-radius: 20px;
      overflow: hidden;
    }

    .viewer-bar {
      display: flex;
      flex-wrap: wrap;
      gap: 16px;
      align-items: center;
      padding: 16px 20px;
      background: rgba(11, 18, 36, 0.9);
      border-bottom: 1px solid var(--panel-border);
      position: sticky;
      top: 0;
      z-index: 2;
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
      background: #0f172a;
      color: var(--text);
      border: 1px solid #1f2937;
      padding: 8px 10px;
      border-radius: 10px;
      font-size: 14px;
    }

    .legend {
      display: flex;
      flex-wrap: wrap;
      gap: 10px;
      align-items: center;
    }

    .legend-item {
      padding: 4px 10px;
      border-radius: 999px;
      font-size: 12px;
      font-weight: 600;
      background: rgba(148, 163, 184, 0.14);
      color: var(--muted);
    }

    .legend-item.not-tracked {
      background: rgba(148, 163, 184, 0.18);
      color: #cbd5f5;
    }

    .legend-item.missed {
      background: rgba(239, 68, 68, 0.18);
      color: #fecaca;
    }

    .legend-item.partial {
      background: rgba(245, 158, 11, 0.18);
      color: #fde68a;
    }

    .legend-item.covered {
      background: rgba(34, 197, 94, 0.18);
      color: #bbf7d0;
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
      padding: 0 20px 20px;
    }

    .file-section {
      display: none;
      margin-bottom: 24px;
      padding: 20px 0 8px;
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
      font-size: 18px;
      word-break: break-all;
    }

    .pill {
      padding: 4px 10px;
      border-radius: 999px;
      font-size: 13px;
      font-weight: 600;
      background: rgba(148, 163, 184, 0.2);
    }

    .pill.high {
      background: rgba(34, 197, 94, 0.15);
      color: var(--covered);
    }

    .pill.medium {
      background: rgba(245, 158, 11, 0.15);
      color: var(--partial);
    }

    .pill.low {
      background: rgba(239, 68, 68, 0.15);
      color: var(--missed);
    }

    .pill.none {
      background: rgba(148, 163, 184, 0.1);
      color: var(--muted);
    }

    .code-table {
      width: 100%;
      border-collapse: collapse;
      font-family: "Fira Code", "JetBrains Mono", ui-monospace, SFMono-Regular, monospace;
      font-size: 13px;
      margin-top: 16px;
      border-radius: 12px;
      overflow: hidden;
      background: #0b1224;
    }

    .code-table td {
      padding: 2px 12px;
      vertical-align: top;
    }

    .code-table .line-no {
      width: 60px;
      text-align: right;
      color: var(--muted);
      border-right: 1px solid #1e293b;
      user-select: none;
    }

    .code-table .code {
      white-space: pre;
    }

    .code-table tr.covered td.code {
      background: rgba(34, 197, 94, 0.18);
    }

    .code-table tr.missed td.code {
      background: rgba(239, 68, 68, 0.2);
    }

    .code-table tr.partial td.code {
      background: rgba(245, 158, 11, 0.2);
    }

    .code-table tr.not-tracked td.code {
      background: rgba(51, 65, 85, 0.3);
    }

    .code-table tr.covered .line-no {
      color: #86efac;
    }

    .code-table tr.missed .line-no {
      color: #fca5a5;
    }

    .code-table tr.partial .line-no {
      color: #fde68a;
    }

    .missing {
      padding: 12px;
      border-radius: 12px;
      background: rgba(248, 113, 113, 0.12);
      color: #fecaca;
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

    <table class="file-table">
      <thead>
        <tr>
          <th>File</th>
          <th>Coverage</th>
          <th>Statements</th>
        </tr>
      </thead>
      <tbody>
        {{range .Files}}
        <tr>
          <td class="file-name"><a href="#{{.Anchor}}">{{.Name}}</a></td>
          <td><span class="pill {{.CoverageClass}}">{{.CoveragePercent}}</span></td>
          <td>{{.CoveredStmts}} / {{.TotalStmts}}</td>
        </tr>
        {{end}}
      </tbody>
    </table>

    <section class="viewer">
      <div class="viewer-bar">
        <div class="file-picker">
          <label for="file-select">File</label>
          <select id="file-select">
            {{range .Files}}
            <option value="{{.Anchor}}">{{.Name}} ({{.CoveragePercent}})</option>
            {{end}}
          </select>
        </div>
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
                <td class="code"><code>{{.Code}}</code></td>
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
  <script>
    const select = document.getElementById('file-select');
    const sections = Array.from(document.querySelectorAll('.file-section'));
    const filters = document.querySelectorAll('[data-filter]');

    function hasSection(anchor) {
      return sections.some((section) => section.id === anchor);
    }

    function activate(anchor, updateHash) {
      sections.forEach((section) => {
        section.classList.toggle('active', section.id === anchor);
      });
      if (updateHash) {
        history.replaceState(null, '', '#' + anchor);
      }
    }

    function syncFromHash() {
      const hash = window.location.hash.replace('#', '');
      if (hash && hasSection(hash)) {
        select.value = hash;
        activate(hash, false);
        return;
      }
      if (sections.length > 0) {
        select.value = sections[0].id;
        activate(sections[0].id, false);
      }
    }

    if (select && sections.length > 0) {
      syncFromHash();
      select.addEventListener('change', (event) => {
        activate(event.target.value, true);
      });
      window.addEventListener('hashchange', syncFromHash);
    }

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
