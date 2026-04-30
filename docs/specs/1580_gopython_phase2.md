---
title: gopython phase 2 - platform os time threads locks
status: draft
created: 2026-04-30
project: gopython
phase: 2
source_root: $HOME/github/python/cpython/Python
target_repo: $HOME/github/tamnd/gopython
release: v0.2.0
---

## gopython phase 2

## Goal

Phase 2 ports CPython 3.14 `Python/` files that sit between the helper layer
and the interpreter runtime: file utilities, path configuration, time
conversion, thread primitives, lock primitives, dynamic loading, and profiling
trampolines.

The pull request series for this phase must keep every port traceable to the
matching CPython source file. It must not introduce parser, compiler, bytecode,
or Python execution shortcuts.

## Scope

Phase 2 covers exactly these CPython files:

| CPython file | Lines | Target package | Parity level |
| --- | ---: | --- | --- |
| `dup2.c` | 37 | `pyos` | Platform |
| `fileutils.c` | 3112 | `pyos` | Platform |
| `pathconfig.c` | 471 | `pyos` | Platform |
| `pytime.c` | 1356 | `pytime` | Exact for represented time APIs |
| `thread.c` | 301 | `pythread` | Structural |
| `thread_nt.h` | 532 | `pythread` | Platform |
| `thread_pthread.h` | 1002 | `pythread` | Platform |
| `thread_pthread_stubs.h` | 191 | `pythread` | Platform |
| `condvar.h` | 315 | `pythread` | Platform |
| `lock.c` | 638 | `pythread` | Exact for lock state machine |
| `parking_lot.c` | 442 | `pythread` | Structural |
| `critical_section.c` | 205 | `pythread` | Structural |
| `brc.c` | 208 | `pythread` | Structural |
| `qsbr.c` | 290 | `pythread` | Structural |
| `stackrefs.c` | 219 | `pythread` | Structural |
| `dynamic_annotations.c` | 154 | `pythread` | Structural no-op hooks |
| `dynload_stub.c` | 9 | `pydynload` | Platform |
| `dynload_shlib.c` | 113 | `pydynload` | Platform |
| `dynload_hpux.c` | 76 | `pydynload` | Platform |
| `dynload_win.c` | 329 | `pydynload` | Platform |
| `importdl.c` | 520 | `pydynload` | Structural |
| `emscripten_signal.c` | 54 | `pyos` | Platform no-op outside wasm |
| `emscripten_syscalls.c` | 319 | `pyos` | Platform no-op outside wasm |
| `emscripten_trampoline.c` | 118 | `pyos` | Platform no-op outside wasm |
| `emscripten_trampoline_inner.c` | 38 | `pyos` | Platform no-op outside wasm |
| `asm_trampoline.S` | 59 | `pyperf` | Structural |
| `perf_trampoline.c` | 634 | `pyperf` | Structural |
| `perf_jit_trampoline.c` | 1363 | `pyperf` | Structural |

## Package Layout

Phase 2 adds top-level packages only. The repository must not add an
`internal/` subtree.

```text
pyos/
  dup2.go
  dup2_unix.go
  dup2_windows.go
  fileutils.go
  fileutils_test.go
  pathconfig.go
  pathconfig_test.go
  emscripten.go

pytime/
  time.go
  time_test.go

pythread/
  thread.go
  thread_test.go
  lock.go
  lock_test.go
  parking_lot.go
  parking_lot_test.go
  critical_section.go
  critical_section_test.go
  qsbr.go
  qsbr_test.go
  stackrefs.go
  stackrefs_test.go
  dynamic_annotations.go

pydynload/
  dynload.go
  dynload_unix.go
  dynload_windows.go
  importdl.go
  importdl_test.go

pyperf/
  trampoline.go
  trampoline_test.go
```

## Porting Rules

- Every exported behavior must point back to a CPython branch or public
  contract.
- Go build tags replace CPython platform files when behavior differs by
  operating system.
- Unsupported host capabilities return explicit errors or documented no-op
  results that match CPython's disabled-platform branches.
- Tests land in the same commit as each ported file.
- No fixture-specific branches are allowed.
- No broad TODO may hide missing CPython behavior.
- No `internal/` directory is allowed.

## Validation Plan

Required local checks before the Phase 2 PR is marked ready:

```sh
go test ./...
go vet ./...
staticcheck ./...
npx markdownlint-cli2 "**/*.md" "$HOME/notes/Spec/1500/1580_gopython_phase2.md"
go run ./tools/sourceinventory --root "$HOME/github/python/cpython/Python" --strict
```

Required release checks before `v0.2.0`:

```sh
for goos in linux darwin windows; do
  for goarch in amd64 arm64; do
    GOOS="$goos" GOARCH="$goarch" go build ./cmd/gopython
  done
done
```

## Public Behavior Target

Phase 2 still does not execute Python source. Public behavior added in this
phase is limited to reusable runtime substrate:

- CPython-style file descriptor duplication and file system encoding helpers.
- Path configuration storage and normalization.
- Nanosecond time conversion with CPython rounding modes.
- Thread creation, thread-local state, mutexes, event locks, condition
  variables, and parking behavior.
- Critical-section, biased-reference-counting, QSBR, and stack-reference
  primitives needed by later interpreter state work.
- Dynamic loader entry points with platform-specific success and failure
  behavior.
- Perf trampoline metadata used later by the VM and JIT-facing layers.

## Migration Order

The implementation should use small commits in this order:

1. Add this spec and keep the PR in draft.
2. Port `dup2.c`.
3. Port `fileutils.c`.
4. Port `pathconfig.c`.
5. Port `pytime.c`.
6. Port `thread.c` and platform thread headers.
7. Port `condvar.h`.
8. Port `lock.c`.
9. Port `parking_lot.c`.
10. Port `critical_section.c`.
11. Port `brc.c`, `qsbr.c`, and `stackrefs.c`.
12. Port `dynamic_annotations.c`.
13. Port dynamic loader files.
14. Port Emscripten support files.
15. Port perf trampoline files.
16. Add release notes for `v0.2.0`.
17. Mark the spec `ready`, mark the PR ready, merge only after CI is green,
    then tag `v0.2.0`.

## File Sections

### `dup2.c`

`dup2.c` wraps platform descriptor duplication. CPython needs the wrapper
because Windows and POSIX differ in descriptor inheritance, error reporting,
and retry behavior.

The Go port will expose a `pyos.Dup2` helper with Unix and Windows build-tag
implementations. Unix uses the host syscall layer and preserves the destination
descriptor contract. Windows keeps a separate implementation because Windows
file handles and C runtime descriptors are not the same object.

Tests must cover replacing an existing descriptor, duplicating onto the same
descriptor when the host allows it, bad descriptors, and platform-specific
errors. Tests must close all descriptors they create.

### `fileutils.c`

`fileutils.c` is the large CPython file system boundary. It owns byte and wide
path conversion, locale and file system encoding decisions, descriptor flags,
file status helpers, path comparison, realpath behavior, and safe wrappers
around low-level file APIs.

The Go port will split this behavior inside `pyos` while keeping one spec
section and one source mapping for the CPython file. Path conversion must
preserve CPython's distinction between bytes, unicode, decode failure, encode
failure, surrogate escape, and null-byte rejection where those concepts are
observable before the Python object model exists.

The first Go surface should use typed path values instead of plain strings
when encoding state matters. Later object-model work can convert Python `str`
and `bytes` objects into these values without changing the file utility core.

Tests must cover ASCII paths, non-ASCII UTF-8 paths, invalid byte sequences,
embedded null rejection, missing files, regular files, directories, symlinks
where available, close-on-exec flag handling where available, and platform
differences for absolute path rules.

### `pathconfig.c`

`pathconfig.c` stores and computes interpreter path configuration: executable,
prefix, exec-prefix, module search path, standard library directory, and
related global path fields.

The Go port will define a `pyos.PathConfig` value with explicit setters,
getters, copy behavior, and clear behavior. It must keep CPython's ownership
rules as value-copy rules in Go, so a caller cannot mutate stored path slices
through an alias.

Tests must cover initialization defaults, setting every path field, clearing
configuration, copy isolation, and joining path list entries using the host
separator.

### `pytime.c`

`pytime.c` converts between seconds, nanoseconds, doubles, time specs, time
vals, and monotonic or wall-clock readings. The important CPython behavior is
rounding: floor, ceiling, half-even, and timeout rounding cannot collapse into
ordinary Go conversions.

The Go port will create `pytime` with explicit duration and timestamp types.
Rounding modes must be named and tested directly. Overflow detection must be
preserved rather than relying on silent Go integer wrap.

Tests must cover zero, positive, negative, fractional, maximum, minimum,
overflow, nanosecond precision, monotonic clock availability, wall clock
availability, and timeout conversion. Golden tests should compare pure
conversion results against CPython 3.14 for representative values.

### `thread.c`

`thread.c` is the platform-independent front door for CPython thread support.
It exposes initialization, thread start, stack size configuration, thread IDs,
native IDs, thread-local storage, and shutdown helpers.

The Go port will use `pythread` as the public package. Go goroutines are not
OS threads, so APIs that promise CPython OS-thread identity must document and
test their exact Go mapping. Where CPython requires a native thread ID, the Go
port must either obtain it through platform support or return the matching
unsupported-path error.

Tests must cover initialization idempotence, starting a worker, joining through
a synchronization primitive, thread-local key lifecycle, stack size validation,
and native ID behavior.

### `thread_nt.h`

`thread_nt.h` implements CPython threads on Windows. It handles Windows thread
creation, locks, semaphores, TLS, native IDs, and timeout conversion.

The Go port will use Windows build tags in `pythread`. Windows-specific files
must keep CPython's timeout and failure behavior where Go exposes the required
host calls. Unsupported Windows-only C runtime details must be represented as
documented structural behavior, not silently ignored.

Tests must compile on Windows in CI and should run behavior checks for locks,
TLS, and native IDs on Windows runners when those runners are available.

### `thread_pthread.h`

`thread_pthread.h` implements CPython threads on POSIX hosts with pthreads. It
owns pthread creation, mutexes, condition variables, TLS, stack sizing, and
timeout translation.

The Go port will use Unix build tags in `pythread`. POSIX-only behaviors must
be separated from the generic thread API so Windows builds do not carry Unix
assumptions.

Tests must compile on Linux and macOS, cover lock acquisition and timeout
behavior, and verify that invalid stack sizes fail predictably.

### `thread_pthread_stubs.h`

`thread_pthread_stubs.h` provides stub behavior when pthread support is not
available. CPython keeps it so builds can compile without real thread support.

The Go port will keep an unsupported-thread build path only when a target
needs it. The stub must fail explicitly for operations that require real
threads and must not pretend to provide synchronization.

Tests must validate explicit unsupported errors under the matching build tag.

### `condvar.h`

`condvar.h` defines condition-variable behavior used by CPython locks and
thread coordination. It handles wait, timed wait, signal, broadcast, and
platform-specific fallback behavior.

The Go port will model this with `sync.Cond` plus explicit timeout handling.
The CPython contract around spurious wakeups must be preserved by requiring
callers to wait in predicate loops.

Tests must cover signal waking one waiter, broadcast waking all waiters,
timeout expiration, no lost wakeup under lock, and repeated wait cycles.

### `lock.c`

`lock.c` implements CPython's higher-level lock object over lower-level thread
primitives. It contains important state transitions for unlocked, locked,
waiting, timed acquire, interrupted acquire, release, and error paths.

The Go port will expose a `pythread.Lock` with CPython acquire modes: blocking,
non-blocking, and timed. Release must reject unlocking an unlocked lock. Timed
acquire must preserve deadline behavior, including zero timeout and expired
timeout.

Tests must cover all acquire modes, release errors, fairness-sensitive waiters
without assuming deterministic scheduling, timeout boundaries, and concurrent
contention under the race detector.

### `parking_lot.c`

`parking_lot.c` is CPython's compact waiter parking implementation. It
coordinates thread suspension, unparking, timeout, and waiter queue behavior.

The Go port will implement the same state machine with channels and mutexes,
not with busy waiting. Queue behavior must be deterministic enough for tests
without relying on goroutine scheduling order.

Tests must cover park, unpark one, unpark all, timeout, cancellation cleanup,
and reuse after a waiter leaves.

### `critical_section.c`

`critical_section.c` manages critical sections used by free-threading work. It
tracks entry, exit, suspension, and resume behavior around object operations.

The Go port will represent critical sections as explicit guard values. Entry
and exit must be paired. Incorrect release order must return a clear error or
panic consistently with the chosen API contract.

Tests must cover nested sections, exit order, suspension, resume, and
interaction with lock acquisition paths.

### `brc.c`

`brc.c` implements biased reference-counting support for free-threaded CPython.
It handles local and shared ownership transitions and merge points.

The Go port will provide the state machinery needed by later object headers,
but it will not attach to real Python objects until the object model phase.
Counters must still preserve CPython transition rules so later integration is
not a rewrite.

Tests must cover local increments, shared increments, merge behavior, overflow
guards, and invalid transitions.

### `qsbr.c`

`qsbr.c` implements quiescent-state based reclamation. CPython uses it to know
when deferred work is safe after participating threads have passed quiescent
states.

The Go port will model epochs, registration, unregistration, advance, and wait
behavior. Waiting must not spin without blocking. Cancellation or timeout must
be explicit in the Go API where the CPython call can block.

Tests must cover epoch advancement, multiple registered threads, unregistering
while waiting, timeout, and safe reclamation ordering.

### `stackrefs.c`

`stackrefs.c` tracks references stored on thread stacks for free-threaded
runtime behavior. It is small but important because later object lifetime code
will rely on its invariants.

The Go port will store stack-reference entries in typed slices or linked
records, whichever maps most directly to CPython's push and pop behavior.

Tests must cover push, pop, scan, empty stack behavior, nested frames, and
cleanup after frame exit.

### `dynamic_annotations.c`

`dynamic_annotations.c` provides sanitizer and race-detector annotation hooks.
In normal builds the functions compile as no-op annotations.

The Go port will expose no-op hooks with names that describe the CPython
annotation intent. The hooks must stay cheap and side-effect free unless a
future build tag wires them to a Go sanitizer-facing implementation.

Tests must verify that each hook can be called and does not mutate observable
state.

### `dynload_stub.c`

`dynload_stub.c` is the dynamic loading implementation for platforms without
loader support. It returns failure rather than pretending extension loading is
available.

The Go port will provide a stub build path in `pydynload` that reports
unsupported dynamic loading.

Tests must cover the unsupported error and stable error text.

### `dynload_shlib.c`

`dynload_shlib.c` loads extension modules from shared libraries on POSIX
systems. It handles path names, loader flags, symbol lookup, and loader errors.

The Go port will keep a POSIX build-tag implementation. Go cannot portably call
all `dlopen` behavior without cgo, so the first implementation must make a
clear choice: either a cgo-backed exact path or an explicit unsupported path
when cgo is disabled. The chosen path must be tested.

Tests must cover missing library, missing symbol, unsupported cgo-disabled
mode, and successful symbol lookup when a safe fixture library is available.

### `dynload_hpux.c`

`dynload_hpux.c` is CPython's HP-UX loader path. It exists for a platform Go
does not normally target.

The Go port will document HP-UX as unsupported unless a Go target exists. The
stub must be explicit and must not affect supported POSIX loaders.

Tests must cover compile-time exclusion and the unsupported result when the
stub build tag is selected.

### `dynload_win.c`

`dynload_win.c` loads extension modules on Windows. It handles DLL path
conversion, Windows loader flags, symbol lookup, and Windows error formatting.

The Go port will use Windows build tags in `pydynload`. It must keep Windows
path handling separate from POSIX loader behavior.

Tests must cover missing DLL, missing symbol, path conversion, and successful
fixture loading when a test DLL is built in CI.

### `importdl.c`

`importdl.c` is the common extension-module import layer above platform dynamic
loaders. It computes initialization symbol names, calls the dynamic loader, and
validates loader failures.

The Go port will keep symbol name computation exact for ASCII and non-ASCII
module names. Actual module execution will remain structural until the import
system and object model exist.

Tests must cover symbol name generation, package-qualified names, invalid
module names, loader error propagation, and a structural successful load using
a fake loader.

### `emscripten_signal.c`

`emscripten_signal.c` provides signal behavior for Emscripten builds. It is
not active on normal desktop hosts.

The Go port will isolate this behind wasm or explicit Emscripten-style build
tags. Non-wasm builds must compile without exposing fake signal support.

Tests must cover no-op behavior on non-wasm builds and compile behavior for
the wasm path where CI can build it.

### `emscripten_syscalls.c`

`emscripten_syscalls.c` adapts syscall behavior for Emscripten. It exists
because browser and WASI environments do not match POSIX file and process
rules.

The Go port will keep this behavior separate from normal `pyos` file
utilities. Unsupported operations must return explicit errors.

Tests must cover unsupported process calls, file call forwarding where Go wasm
supports it, and stable error behavior where it does not.

### `emscripten_trampoline.c`

`emscripten_trampoline.c` manages trampoline calls needed by CPython in
Emscripten environments.

The Go port will provide a wasm-only structural implementation. Desktop builds
must not depend on these symbols.

Tests must cover build separation and no-op behavior under the non-wasm stub.

### `emscripten_trampoline_inner.c`

`emscripten_trampoline_inner.c` contains the inner trampoline body used by the
Emscripten trampoline layer.

The Go port will keep this paired with `emscripten_trampoline.c`. Its public
surface should stay minimal so later runtime code can call one trampoline API
instead of depending on inner details.

Tests must cover the call path through the outer trampoline API.

### `asm_trampoline.S`

`asm_trampoline.S` is architecture-specific assembly used by CPython perf
trampoline support.

The Go port will not translate the assembly instruction by instruction in
Phase 2 unless a Go assembly target is required by the surrounding perf port.
Instead, it will expose the structural call boundary used by the C code and
document unsupported architectures clearly.

Tests must cover that the package builds on linux, darwin, and windows for
amd64 and arm64.

### `perf_trampoline.c`

`perf_trampoline.c` writes trampoline metadata so Linux perf can associate
runtime-generated code with Python frames.

The Go port will model metadata records, symbol names, code ranges, and output
file formatting. It must not fake integration with the VM before bytecode and
frames exist.

Tests must cover record encoding, symbol formatting, disabled perf mode,
unsupported platform behavior, and cleanup of temporary metadata files.

### `perf_jit_trampoline.c`

`perf_jit_trampoline.c` emits JIT dump data for perf tooling. It is larger
than `perf_trampoline.c` because it owns file headers, code-load records,
debug-info records, close records, timestamps, and byte ordering.

The Go port will implement the binary writer exactly for the represented
record types. Host endianness, timestamp source, process ID, and record sizes
must be tested directly.

Tests must cover header bytes, code-load records, debug records, close record,
multiple records in one file, disabled mode, unsupported platform behavior,
and deterministic output through injected clock and process metadata.

## Completion Checklist

- Every Phase 2 CPython file has a Go target file or explicit platform stub.
- Every Phase 2 CPython file has a detailed section in this spec.
- Every exported Go symbol has direct unit tests.
- Platform behavior is isolated with build tags instead of runtime guessing
  where the platform changes the API contract.
- Unsupported behavior is explicit and tested.
- `go test ./...` passes.
- `go vet ./...` passes.
- `staticcheck ./...` passes.
- Markdown lint passes for repo docs and the external spec.
- Source inventory still tracks every CPython `Python/` file.
- CI is green before merge.
- `v0.2.0` release publishes Linux, macOS, and Windows binaries with checksums.
