# Release Engineering

This document describes how to build, release, and verify DevOpsGen.

## Prerequisites

- Go 1.24.2 or later
- Git

## Build

```bash
make build
```

Produces a `devopsgen` binary for the current platform.

## Cross-platform release

Uses [GoReleaser](https://goreleaser.com/) (`.goreleaser.yaml`):

```bash
goreleaser release --clean
```

Targets: Linux, macOS (Intel + Apple Silicon), Windows (amd64).
Version, commit, and build date are injected via ldflags:

```bash
-X devopsgen.dev/devopsgen/internal/cli.version={{.Version}}
-X devopsgen.dev/devopsgen/internal/cli.commit={{.Commit}}
-X devopsgen.dev/devopsgen/internal/cli.date={{.Date}}
```

## Versioning

Semver (https://semver.org). Development builds report `dev`.

## Benchmark

Core engine benchmark (NFR-01):

```
go test -bench=BenchmarkGenerate -benchtime=10x ./internal/generator/
```

Reference machine (2026-10-16): Intel Core i7-14650HX, Go 1.24.2,
windows/amd64.

| Benchmark              | ns/op | B/op   | allocs/op |
|---|---|---|---|
| BenchmarkGenerate     | 54810 | 49504  | 582       |
| BenchmarkGenerate_Node| 22330 | 29200  | 281       |

NFR-01 bounds: <=20 artifacts, <=256 KiB total output, <1s rendering.
Measured output is ~50 KiB, well within bounds.

## CI

`.github/workflows/ci.yml` runs lint, test with an 80% coverage floor,
multi-platform builds, smoke tests, and end-to-end generation + validation.