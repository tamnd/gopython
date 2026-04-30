# Repository Setup

Set the GitHub repository description to:

```text
Idiomatic Go port of CPython 3.14's Python/ compiler, runtime, and VM layer.
```

Set these topics:

```text
go
golang
python
cpython
python314
interpreter
compiler
bytecode
virtual-machine
parser
```

After the phase 0 pull request is merged, create the first tag:

```sh
git tag v0.0.0
git push origin v0.0.0
```

Release title:

```text
gopython v0.0.0 repository scaffold
```

Release notes:

```text
Initial repository scaffold with CI, release workflow, README, and CPython
Python/ source inventory. No CPython runtime or compiler logic is ported in
this tag.
```
