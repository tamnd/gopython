# gopython

gopython is an idiomatic Go port of CPython 3.14's `Python/` directory.

The target is the compiler, runtime, and VM layer that lets Python source move
through tokenizing, parsing, AST construction, symbol analysis, bytecode
generation, and execution with CPython-compatible behavior.

## Status

This repository is at phase 0.

The project shell, CI, release workflow, and source inventory are being set up
first. There is no Python interpreter in this repository yet. The first release
tag is `v0.0.0` and means scaffold only.

## Scope

The source inventory tracks all 118 top-level files from:

```text
$HOME/github/python/cpython/Python
```

Every later migration gets its own pull request, spec, tests, and source
mapping. A migration is not complete until the behavior can be traced back to
CPython 3.14 source.

## Rules

- No `internal/` package tree.
- No cgo dependency for normal builds.
- No fixture-only branches.
- No skipped CPython behavior marked as complete.
- One migration unit per pull request after phase 0.
- Tests must ship with each migration.

## Commands

```sh
go test ./...
go run ./tools/sourceinventory --list
go run ./cmd/gopython --version
```

To compare the checked-in inventory with a local CPython checkout:

```sh
go run ./tools/sourceinventory --root "$HOME/github/python/cpython/Python" --strict
```

## Repository Metadata

Description:

```text
Idiomatic Go port of CPython 3.14's Python/ compiler, runtime, and VM layer.
```

Topics:

```text
go golang python cpython python314 interpreter compiler bytecode
virtual-machine parser
```
