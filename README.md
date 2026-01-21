# Beautiful Go Coverage

Generate a modern HTML coverage report directly from a Go coverprofile without converting the `go tool cover` output.

## Usage

1. Generate a coverprofile:

```bash
go test ./... -coverprofile=coverage.out
```

2. Create the HTML report:

```bash
go run ./cmd/beautiful-coverage -profile coverage.out -out coverage.html
```

## Flags

- `-profile`: path to the coverprofile file (required).
- `-out`: output HTML file (default `coverage.html`).
- `-root`: root directory used to resolve source file paths (default `.`).
- `-title`: report title (default `Go Coverage Report`).
- `-assets`: directory name or path to write local CSS/JS assets (default `assets`).

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
