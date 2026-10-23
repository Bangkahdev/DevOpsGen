# DevOpsGen

Developer-oriented DevOps scaffolding tool. DevOpsGen generates deterministic
Dockerfiles, GitHub Actions workflows, and Kubernetes manifests from a small,
explicit configuration.

> Documentation: **[devopsgen.bangkah.tech](https://devopsgen.bangkah.tech/)** —
> screenshots of every command, the interactive wizard, flags reference and
> exit codes. Build it with `npm run build` inside `docs/web`.

## Features

- Interactive TUI wizard and non-interactive headless modes.
- Stack detection for Go, Node.js, and Python.
- Deterministic output for the same binary version and configuration.
- Safe filesystem behavior: no-overwrite by default, dry-run, explicit force,
  backup, and best-effort rollback.
- Credential-free core engine: no network access or registry credentials
  required for generation or internal validation.

## Supported stacks

- Go (`binary`)
- Node.js: Express, Next.js, Vite, NestJS, generic
- Python: FastAPI, Django, Flask, generic

## Installation

```bash
go install devopsgen.dev/devopsgen@latest
```

Or build from source:

```bash
make build
```

## Quick start

Non-interactive:

```bash
devopsgen init --stack go --ci github-actions --registry ghcr \
  --image bangkah/my-app --port 8080
```

Interactive (when a terminal is available and configuration is incomplete):

```bash
devopsgen init
```

## Commands

| Command | Description |
|---|---|
| `init` | Generate DevOps configuration. |
| `detect` | Detect candidate project stack (read-only). |
| `validate` | Validate generated artifacts. |
| `list` | List supported stacks, app types, registries, and templates. |
| `version` | Print version information. |

## Exit codes

| Code | Meaning |
|---:|---|
| 0 | Success. |
| 1 | General error, I/O, or rollback failure. |
| 2 | Invalid input, flag, or configuration. |
| 3 | File conflict without explicit overwrite. |
| 4 | Artifact validation failure. |
| 5 | Ambiguous or unsupported stack. |

## Project layout

```text
devopsgen/
├── cmd/devopsgen/        # entry point
├── internal/
│   ├── cli/              # Cobra command tree
│   ├── wizard/           # Bubble Tea TUI
│   ├── detector/         # stack detection
│   ├── config/           # models, enums, resolution, validation
│   ├── templates/        # embedded template loader
│   ├── generator/        # template rendering
│   ├── output/           # in-memory artifact representation
│   ├── validator/        # artifact validation
│   └── filesystem/       # path safety, plan, backup, commit, rollback
├── templates/            # source templates (embedded at build)
├── tests/
├── docs/
├── go.mod
├── Makefile
└── README.md
```

---

Bagian dari gudang project publik **[Bangkah](https://bangkah.tech)** — dokumentasi dan alat developer di satu domain.
