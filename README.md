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
