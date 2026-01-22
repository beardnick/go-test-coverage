# Beautiful Go Coverage

Generate a modern HTML coverage report directly from a Go coverprofile without converting the `go tool cover` output. The report is emitted as a single, self-contained HTML file with embedded styles and scripts.

Inspired by `https://github.com/gha-common/go-beautiful-html-coverage`.

## Usage

1. Generate a coverprofile:

```bash
go test ./... -coverprofile=coverage.out
```

2. Create the HTML report:

```bash
go run ./cmd/beautiful-coverage -out coverage.html
```

## Flags

- `-profile`: path to the coverprofile file (default `coverage.out`).
- `-out`: output HTML file (default `coverage.html`).
- `-root`: root directory used to resolve source file paths (default: profile directory).
- `-title`: report title (default `Go Coverage Report`).

## Coverage Algorithm

```mermaid
flowchart TD
    Start([Start]) --> ParseProfiles[Parse coverprofile]
    ParseProfiles --> LoadModule[Load module info from go.mod]
    LoadModule --> EachProfile{For each file profile}
    EachProfile --> CountStmts[Sum total + covered statements per block]
    CountStmts --> ResolvePath[Resolve source path using root/module]
    ResolvePath --> ReadSource{Read source file}
    ReadSource -->|missing| MarkMissing[Mark file missing and keep totals]
    ReadSource -->|found| LineStates[Mark line states from blocks]
    LineStates --> ClassifyLines[Classify each line: covered/missed/partial/not-tracked]
    ClassifyLines --> BuildFile[Build file report: percent, class, lines]
    MarkMissing --> Accumulate
    BuildFile --> Accumulate[Accumulate totals across files]
    Accumulate --> Totals[Compute total percent + class]
    Totals --> Tree[Build file tree from relative paths]
    Tree --> End([Report ready])
```
