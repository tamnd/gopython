---
title: gopython phase 1 - core helper ports
status: draft
created: 2026-04-30
project: gopython
phase: 1
source_root: $HOME/github/python/cpython/Python
target_repo: $HOME/github/tamnd/gopython
release: v0.1.0
---

## gopython phase 1

## Goal

Phase 1 ports the dependency-light helper layer from CPython 3.14 `Python/` to
Go. The pull request must remain limited to these files and their tests.

No parser, compiler, VM, object model, or interpreter behavior is introduced in
this phase.

## Scope

Phase 1 covers exactly these CPython files:

| CPython file | Target package | Parity level |
| --- | --- | --- |
| `README` | `pycore` docs | Exact documentation inventory |
| `config_common.h` | `pycore` | Structural |
| `pyctype.c` | `pycore` | Exact |
| `pystrcmp.c` | `pycore` | Exact |
| `mysnprintf.c` | `pycore` | Exact for Go formatting boundary |
| `mystrtoul.c` | `pycore` | Exact for unsigned and signed long parsing |
| `pystrhex.c` | `pycore` | Exact |
| `pyhash.c` | `pycore` | Exact helpers, structural keyed hash |
| `pymath.c` | `pycore` | Exact |
| `pyfpe.c` | `pycore` | Structural no-op |
| `getcopyright.c` | `pycore` | Exact |
| `getcompiler.c` | `pycore` | Exact for Go build metadata |
| `getplatform.c` | `pycore` | Exact for Go platform mapping |
| `getversion.c` | `pycore` | Exact for configured version string |
| `stdlib_module_names.h` | `pycore` | Generated data parity |
| `suggestions.c` | `pycore` | Exact |
| `uniqueid.c` | `pycore` | Exact |
| `object_stack.c` | `pycore` | Exact |
| `index_pool.c` | `pycore` | Exact |
| `hashtable.c` | `pycore` | Exact |
| `pyarena.c` | `pycore` | Exact for allocation lifetime semantics |

## Package Layout

Phase 1 creates a top-level `pycore` package. No `internal/` tree is allowed.

```text
pycore/
  config.go
  ctype.go
  ctype_test.go
  hash.go
  hash_test.go
  hashtable.go
  hashtable_test.go
  hex.go
  hex_test.go
  index_pool.go
  index_pool_test.go
  object_stack.go
  object_stack_test.go
  parseint.go
  parseint_test.go
  platform.go
  platform_test.go
  snprintf.go
  snprintf_test.go
  stdlib_modules.go
  stdlib_modules_test.go
  suggestions.go
  suggestions_test.go
  uniqueid.go
  uniqueid_test.go
  arena.go
  arena_test.go
```

## Porting Rules

- Keep the same algorithmic branches as CPython.
- Replace C macros with typed constants or small functions.
- Replace C pointer ownership with Go values while preserving visible lifetime
  behavior.
- Do not add interpreter shortcuts.
- Do not hard-code fixture behavior.
- Do not mark a structural dependency as exact.
- Each source file gets tests in the same commit as the port.

## Commit Plan

1. Spec commit.
2. `pyctype.c`, `pystrcmp.c`, `pystrhex.c`.
3. `mystrtoul.c`, `mysnprintf.c`.
4. Version, platform, copyright, compiler, math, FPE, config.
5. `stdlib_module_names.h` as generated Go data.
6. `suggestions.c`.
7. `uniqueid.c`, `object_stack.c`, `index_pool.c`.
8. `hashtable.c`.
9. `pyarena.c`.
10. Release documentation for `v0.1.0`.

## Validation

Every implementation commit must pass:

```sh
go test ./...
go vet ./...
staticcheck ./...
npx markdownlint-cli2 "**/*.md"
go run ./tools/sourceinventory --root "$HOME/github/python/cpython/Python" --strict
```

Before release `v0.1.0`, also run the release cross-build locally:

```sh
targets="linux/amd64 linux/arm64 darwin/amd64 darwin/arm64"
targets="$targets windows/amd64 windows/arm64"
for target in $targets; do
  GOOS=${target%/*} GOARCH=${target#*/} go build ./cmd/gopython
done
```

## Release

After the Phase 1 PR is merged, create tag `v0.1.0`.

Release notes live in:

```text
docs/releases/v0.1.0.md
```

The release must include binaries and checksums from the release workflow.
