---
title: gopython phase 1 - core helper ports
status: ready
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
| `pyhash.c` | `pycore` | Exact |
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

## Port Notes

### `README`

The CPython file is a one-line description of the `Python/` directory. The Go
port stores the exact line in `pycore.PythonDirectoryReadme`. This keeps the
inventory visible without adding runtime behavior.

Tests assert the string byte-for-byte, including the trailing newline.

### `config_common.h`

The C helpers read required keys from a Python dict and report a missing key or
invalid type. Phase 1 does not have Python objects yet, so the Go port uses
`map[string]any` as the direct structural stand-in.

`ConfigDictGet` returns the stored value or the same missing-key message shape.
`ConfigDictInvalidType` returns the invalid-type message. Later object-model
work can replace `map[string]any` with a real Python dict without changing the
caller contract.

Tests cover existing keys, missing keys, and invalid-type errors.

### `pyctype.c`

CPython defines locale-independent byte classification tables and case maps.
The Go port builds the same 256-entry classification behavior at init time.
Only ASCII letters, digits, hex digits, and whitespace receive flags. Bytes
above `0x7f` keep zero flags and map to themselves.

`Charmask`, `IsLower`, `IsUpper`, `IsAlpha`, `IsDigit`, `IsXDigit`,
`IsAlnum`, `IsSpace`, `ToLower`, and `ToUpper` mirror the C macros.

Tests cover all ASCII class ranges, punctuation negatives, high-byte identity,
and signed-byte masking with `Charmask(-1)`.

### `pystrcmp.c`

CPython compares NUL-terminated byte strings with ASCII-only lowercasing. The
Go port treats the end of a Go string as the implicit NUL and stops at explicit
`\x00` bytes, matching the C loop.

`MyStrICmp` and `MyStrNICmp` preserve CPython return semantics by returning the
difference between lowercased bytes, not only `-1`, `0`, or `1`.

Tests cover equal case-insensitive strings, ordering, size-limited comparison,
zero-size comparison, and embedded NUL termination.

### `pystrhex.c`

CPython hex-encodes bytes with optional grouping separators. Positive grouping
places separators from the right. Negative grouping places separators from the
left. Groups larger than the input disable separators.

The Go port keeps those branches in `strHexImpl`. Public helpers return either
`string` or `[]byte`, matching the C distinction between unicode and bytes
return values.

Tests cover no separator, positive grouping, negative grouping, oversize
grouping, byte return values, and separator validation.

### `pyhash.c`

CPython keeps hash helpers in one file so equal numeric values keep equal
hashes across types. The Go port implements double hashing, pointer hashing,
SipHash13 buffer hashing, keyed hashing, and hash function metadata.

`HashDouble` follows CPython's finite-float reduction modulo
`2**_PyHASH_BITS - 1`, infinity handling, NaN fallback, sign handling, and the
reserved `-1` to `-2` rewrite. `HashBuffer` keeps the empty-buffer zero case
and the reserved `-1` rewrite. `KeyedHash` ports `_Py_KeyedHash` using
SipHash13 with the second key lane set to zero.

Tests cover finite floats, signed `-1`, infinities, NaN fallback, pointer hash
reserved-value handling, empty buffers, a fixed zero-secret SipHash13 result, a
fixed keyed result, and function metadata.

### `mysnprintf.c`

CPython wraps `vsnprintf` so the last byte is always NUL and the return value
is the full formatted length. Go does not expose C varargs, so Phase 1 maps the
formatting boundary to `fmt.Sprintf` while preserving buffer truncation and
termination rules.

`PyOSSnprintf` panics on an empty buffer because the C API asserts that the
buffer is non-null and size is positive.

Tests cover fitting writes, truncated writes, final NUL termination, returned
length, and the empty-buffer assertion.

### `mystrtoul.c`

CPython implements base-aware unsigned and signed long parsing without locale
dependencies. The Go port uses 64-bit `long` and `unsigned long` semantics,
matching the checked-in CPython 3.14 source configuration used for this port.

`PyOSStrtoul` preserves whitespace skipping, explicit base handling, `0x`,
`0o`, and `0b` prefix rules, invalid-prefix end offsets, leading-zero behavior,
digit limits, overflow spooling, and `ERANGE` behavior through `ErrRange`.

`PyOSStrtol` preserves sign handling, `LONG_MIN`, `LONG_MAX`, and overflow
mapping.

Tests cover base detection, invalid bases, invalid prefixes, leading-zero
behavior, unsigned overflow, signed min and max, and signed overflow.

### `pymath.c`

The CPython file only defines x87 control-word helpers when a specific x86
assembly feature is enabled. Go does not expose that FPU control word portably,
and Phase 1 does not need it.

The port records this as structural behavior. No public x87 API is exposed
until a later platform-specific requirement appears.

### `pyfpe.c`

CPython keeps stable-ABI floating-point exception symbols after removing
`--with-fpectl`. The live behavior is only `PyFPE_dummy`, which returns `1.0`.

The Go port exposes `PyFPEdummy` with the same return behavior. The obsolete
setjmp storage is intentionally not modeled because Go has no stable ABI
extension surface in Phase 1.

Tests assert the return value.

### `getcopyright.c`

The C file returns a fixed manually maintained copyright string. The Go port
stores the same text as a package constant and returns it from
`PyGetCopyright`.

Tests assert representative stable content.

### `getcompiler.c`

CPython returns a compiler identification string selected by preprocessor
macros. The Go port has no C compiler in normal builds, so it returns the Go
toolchain identity as `[Go <runtime.Version()>]`.

This keeps the same role as CPython's API: a short build-tool identifier.

Tests compare the value with `runtime.Version()`.

### `getplatform.c`

CPython returns the configured `PLATFORM` macro. The Go port maps that role to
`runtime.GOOS`, which is the build platform identifier available in normal Go
builds.

Tests compare the value with `runtime.GOOS`.

### `getversion.c`

CPython builds a cached version string from `PY_VERSION`, build info, and the
compiler string. The Go port keeps the same constants from `patchlevel.h`,
including `PY_VERSION_HEX`, and lazily builds `PyGetVersion`.

`PyGetBuildInfo` uses CPython's fallback values from `getbuildinfo.c`:
`main, Jan 01 1970, 00:00:00`.

Tests cover `PY_VERSION`, `PY_VERSION_HEX`, build info, and version prefix.

### `stdlib_module_names.h`

The CPython file is generated data for `sys.stdlib_module_names`. The Go port
checks in the same sorted list as `pycore.StdlibModuleNames` and builds a set
for membership checks.

Tests assert the exact count, strict sortedness, known positive modules, and
known non-stdlib negatives.

### `suggestions.c`

CPython uses a bounded Levenshtein edit cost for attribute and name
suggestions. The Go port keeps the same constants, substitution cost, common
affix trimming, max string size rule, quick-fail rule, one-row dynamic
programming buffer, and best-candidate selection rule.

`UTF8EditCost` operates on UTF-8 bytes like the C source. `CalculateSuggestion`
uses `[]string` in place of Python lists until the object model exists.

Tests cover exact matches, case-only edits, substitutions, insertions,
max-cost early failure, candidate-list limits, skipped exact candidates, and
far-match rejection.

### `uniqueid.c`

CPython allocates one-based unique IDs from an interpreter-local table and
returns released entries through a freelist. The Go port keeps the same table
and freelist model in `UniqueIDPool`.

`Assign` grows the table from zero to eight entries, then doubles capacity.
`Release` pushes the released ID to the freelist, so reuse follows the same
LIFO behavior as CPython. `Finalize` clears the table and freelist.

Tests cover first allocation, object lookup, release, ID reuse, resize from the
minimum pool size, and finalization.

### `object_stack.c`

CPython stores objects in linked chunks of 254 entries so each chunk has a
power-of-two size. The Go port keeps the chunk size and linked-list structure
with a generic `ObjectStack[T]`.

`Push`, `Pop`, `Size`, `Merge`, and `Clear` follow the C header and source
logic. `Merge` appends the destination chain to the bottom of the source chain,
then moves the source head into the destination and clears the source.

Tests cover empty pop, chunk rollover, LIFO order, merge order, source clearing,
and clear behavior.

### `index_pool.c`

CPython allocates monotonically increasing indices and reuses freed indices via
a min-heap. The Go port keeps that heap behavior in `IndexPool`.

`AllocIndex` returns the smallest available freed index, or the next new index
when the heap is empty. `FreeIndex` adds an index to the heap. Both operations
increment `TLBCGeneration`, matching the C generation counter updates.

Tests cover sequential allocation, free-order independent reuse in sorted
order, generation increments, and `Fini` reset behavior.

### `hashtable.c`

CPython's `_Py_hashtable_t` uses a power-of-two bucket array and singly linked
entries. The Go port keeps that structure in `HashTable` instead of using a Go
map, because insertion order inside buckets, rehash thresholds, steal behavior,
and destroy callbacks are part of the source logic.

`Set` prepends entries to buckets and grows when load is above `0.50`. `Steal`
removes an entry without calling destroy callbacks and shrinks through the same
rehash path when load falls below `0.10`. `Clear` calls destroy callbacks,
empties every bucket, and rehashes back to the minimum bucket count.

Tests cover set, get, duplicate-key rejection, steal, rehash growth, foreach
iteration, foreach early stop, destroy callbacks, clear, destroy, and bucket
rounding.

### `pyarena.c`

CPython's arena owns linked raw-memory blocks and a list of Python objects that
stay alive until the arena is freed. The Go port keeps the linked block model
with `Arena`, `arenaBlock`, and block-local offsets.

`Malloc` rounds each allocation up to the same 8-byte alignment. If the current
block cannot fit the request, it links a new block. Oversized requests allocate
a one-off block sized to the rounded request, matching CPython's branch.
`AddObject` records objects until `Free`, which clears blocks and retained
objects and makes later arena use invalid.

Tests cover aligned allocation sizes, default block use, oversized block
allocation, object retention, free cleanup, and use-after-free rejection.

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
