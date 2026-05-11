# Contributing to glog

## Test matrix (CI)

| Area | Packages / paths |
|------|------------------|
| Root module | `go vet ./...`, `go test ./... -race` |
| Submodule | `tests/` same commands |
| Benchmarks (optional) | `go test -bench=. -benchmem -run=^$ ./...` (root; can be slow) |

## Local checks (same as CI)

From the repository root:

```bash
go vet ./...
go test ./... -race -count=1
```

The `tests/` directory is a separate Go module; run it explicitly:

```bash
cd tests && go vet ./... && go test ./... -race -count=1
```

Optional benchmarks:

```bash
go test -bench=. -benchmem ./...
```

## Architecture refactor roadmap (phase 3+)

These are larger follow-ups; keep changes in small PRs.

1. **Handler pipeline** — Centralize `filter → format → sink` in one helper (or small type) so `FileHandler`, `StreamHandler`, `StdoutHandler`, and `SyslogHandler` only differ in the final write path. `handler/pipeline.go` (`applyFilter`) is the starting point.
2. **Package boundaries** — Split the top-level `log` package into a thin public API vs `internal/` for engine, defaults, and handler wiring to reduce coupling and test surface.
3. **File rotator** — Split `handler/file_rotator.go` into focused units: open/path helpers, size vs time trigger strategies, and retention listing/deletion, sharing `collectRotatedBackupPaths`-style utilities.

## Compatibility notes

- JSON field for trace identifiers is serialized as `trace_id` (see `message.Record`). Update consumers that still expect the old misspelled key.
