# Spec — Source-relative output for protoc-gen-* plugins (built on protoc-gen-star)

Generic migration spec. Re-apply to any custom protoc plugin written against
`github.com/lyft/protoc-gen-star` whose output path currently flattens to
`<basename>.<ext>` and therefore collides on multi-package proto layouts.

Original idea: [Intrinsec/protoc-gen-sanitize#2](https://github.com/Intrinsec/protoc-gen-sanitize/pull/2)
by @cobbinma (2023-10-31). This spec is the redo recipe.

## Symptom you are fixing

```
$ tree proto
proto/
├── foo/a.proto      # package foo.v1
└── bar/a.proto      # package bar.v1

$ protoc --custom_out=gen proto/foo/a.proto proto/bar/a.proto
# Both files emit "a.custom.go" → second invocation overwrites first.
```

Root cause: plugin builds output filename from
`f.InputPath().BaseName()` (which strips the directory component) plus a
fixed suffix.

## Fix in one sentence

Compute the output filename via `pgsgo.Context.OutputPath(file).SetExt(<ext>).String()`
instead of `f.InputPath().BaseName() + <ext>`. Tell `protoc` to use the repo
root (or whatever directory contains the source tree) as the plugin output
base instead of a flat output directory.

## The two code edits

### 1. Plugin Go source — replace the filename computation

Find the place that constructs the output filename. Typical shape:

```go
// Before
name := f.InputPath().BaseName() + ".<ext>.go"
p.AddGeneratorTemplateFile(name, p.tpl, f)
```

Replace with:

```go
// After
name := p.ctx.OutputPath(f).SetExt(".<ext>.go")
p.AddGeneratorTemplateFile(name.String(), p.tpl, f)
```

Notes:

- `p.ctx` is the `pgsgo.Context` value normally stored on the module during
  `InitContext`. If your plugin does not import `pgsgo`, add:

  ```go
  import pgsgo "github.com/lyft/protoc-gen-star/lang/go"

  // In your module struct:
  ctx pgsgo.Context

  // In InitContext:
  p.ctx = pgsgo.InitContext(c.Parameters())
  ```

- `pgsgo.Context.OutputPath(file)` returns the input proto path with the
  extension swapped for `.pb.go`. For `proto/foo/a.proto` you get
  `proto/foo/a.pb.go`.

- `SetExt(".<ext>.go")` strips only the trailing `.go` (via `BaseName()`),
  leaving any `.pb` infix. Example: `a.pb.go` → `a.pb.<ext>.go`. If the
  plugin's existing suffix did **not** include `.pb`, drop the `.pb` infix
  the same way by adjusting the extension argument.

- `AddGeneratorTemplateFile` expects a `string`. `pgs.FilePath` has a
  `.String()` method.

### 2. Build invocation — set the output base to the source root

```diff
- protoc --<plugin>_out=<flat-dir>  proto/**/*.proto
+ protoc --<plugin>_out=.           proto/**/*.proto
```

The plugin now returns the **full** relative path (e.g.
`proto/foo/a.pb.<ext>.go`). The `--<plugin>_out` value is the base directory
protoc joins with that path — must therefore be the source root, not a flat
landing directory, or you double-prefix.

If your build system invokes protoc from a subdirectory, adjust `-I` and the
`--<plugin>_out` value accordingly so the relative path round-trips.

## Test fixture you must add

Without a multi-directory fixture, the change reads as a no-op for any
existing single-directory proto tree. Add the smallest possible regression
catcher:

```
testdata/
├── pkg_a/
│   └── thing.proto      # package pkg_a; option go_package = "./testdata/pkg_a;pkg_a";
└── pkg_b/
    └── thing.proto      # package pkg_b; option go_package = "./testdata/pkg_b;pkg_b";
```

Each `thing.proto` should contain exactly one message that exercises **one**
of your plugin's behaviors. After generation, both directories must contain
`thing.pb.<ext>.go` and the resulting Go packages must compile together
under `go build ./...`. Before the fix, the second invocation overwrites
the first or fails — that's the regression.

Add a `_test.go` per package that imports and instantiates the generated
type, asserting one observable side-effect of your plugin. CI runs both,
proving the collision is gone.

## .gitignore / build artefact patterns

If your project gitignores generated files by their flat name
(`tests/*.pb.<ext>.go`), bump the glob to `tests/**/*.pb.<ext>.go` so the
new subdirectory outputs stay ignored. Same for any clean targets in the
Makefile.

## Rollout / compatibility

- **Filename suffix is preserved** as long as you keep `.pb` in the new
  `SetExt` argument (`.SetExt(".pb.<ext>.go")` for plugins that previously
  emitted `<base>.pb.<ext>.go`). Downstream consumers continue to import
  the same filenames.

- **Single-package consumers see no behavior change** — same input, same
  output path, same file.

- **Multi-package consumers** newly get correct, namespaced outputs.
  Migration on the consumer side: delete the previous flat outputs and
  regenerate.

- **Edition 2024 / protoc 28+** support is independent of this change; this
  spec works with any protoc version that `protoc-gen-star` itself
  supports.

## Verification checklist

Before merging the redo:

- [ ] `protoc --<plugin>_out=. <existing-fixtures>` produces the same set of
      output files as before (sanity check, no regression).
- [ ] `protoc --<plugin>_out=. <new-multi-pkg-fixtures>` produces one output
      file per input proto, with the directory layout mirroring the input
      tree.
- [ ] Existing test suite still passes (output files land at the same
      paths).
- [ ] New multi-package test suite passes.
- [ ] `golangci-lint run ./...` clean.
- [ ] `govulncheck ./...` clean (no new dep introduced — this change is
      pure refactor against the same `protoc-gen-star` API).

## Why this is worth the churn

- Eliminates a silent data-loss bug for multi-package consumers.
- Aligns the plugin with the de-facto protoc-gen-go convention
  (`paths=source_relative` semantics by default).
- Tiny change (~5 lines of plugin Go source plus invocation flag).
- No new dependencies, no API break for the plugin itself, no filename
  break for downstream when `.pb` infix preserved.

## Reference implementation

This spec was applied to `protoc-gen-sanitize` in commit `<filled-in
during commit>`. Diff and test fixtures live at
`docs/superpowers/plans/2026-05-20-source-relative-output.md` in this repo.
Read alongside if you want a worked example.
