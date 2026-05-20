# Source-relative output for protoc-gen-sanitize

## Goal

Make the plugin compute its output path from `pgsgo.Context.OutputPath`
instead of `f.InputPath().BaseName()`, so multi-directory proto trees
generate correctly without collisions.

Redo of [PR #2](https://github.com/Intrinsec/protoc-gen-sanitize/pull/2)
by @cobbinma against current master. Co-author credit preserved in commit
trailer.

## Context

`sanitizer.go:117` currently does:

```go
name := f.InputPath().BaseName() + ".pb.sanitize.go"
```

`BaseName()` drops the directory, so `foo/a.proto` and `bar/a.proto` both
produce `a.pb.sanitize.go` — second write overwrites first.

`pgsgo.Context.OutputPath(file)` returns
`<input-dir>/<base>.pb.go`, so combined with `SetExt(".sanitize.go")` the
new filename is `<input-dir>/<base>.pb.sanitize.go` — namespaced and
collision-free.

Generic recipe lives at `docs/source-relative-output-spec.md`.

## File structure

- Modify: `sanitizer.go`
- Modify: `Makefile`
- Modify: `.gitignore`
- Create: `tests/sub/entity_sub.proto`
- Create: `tests/sub/sub_test.go`

## Tasks

### Task 1 — Code: source-relative output path

- [ ] `sanitizer.go:117`:

  ```diff
  - name := f.InputPath().BaseName() + ".pb.sanitize.go"
  + name := p.ctx.OutputPath(f).SetExt(".sanitize.go")
  ```

- [ ] `sanitizer.go:121`:

  ```diff
  - p.AddGeneratorTemplateFile(name, p.tpl, f)
  + p.AddGeneratorTemplateFile(name.String(), p.tpl, f)
  ```

- [ ] `go build -mod=vendor ./...` exits 0.

### Task 2 — Build: point protoc at the source root

- [ ] `Makefile` `generate` target:

  ```diff
  - @protoc -I . --plugin=protoc-gen-$(NAME)=$(shell pwd)/bin/protoc-gen-$(NAME) --$(NAME)_out=tests tests/*.proto
  + @protoc -I . --plugin=protoc-gen-$(NAME)=$(shell pwd)/bin/protoc-gen-$(NAME) --$(NAME)_out=.     $(PROTO_FIXTURES)
  ```

  with `PROTO_FIXTURES := $(shell find tests -name '*.proto')` defined
  near the top of the file so the `protoc-gen-go` invocation can reuse it.

- [ ] `protoc-gen-go` invocation likewise uses `$(PROTO_FIXTURES)` and
      `--go_out=.`.

### Task 3 — Fixture: multi-package regression catcher

- [ ] Create `tests/sub/entity_sub.proto`:

  ```protobuf
  syntax = "proto3";

  package sub;

  import "sanitize/sanitize.proto";

  option go_package = "./tests/sub;sub";

  message SubEntity {
      string name = 1 [
          (sanitize.rules) = {
              kind: TEXT,
              trim: true
          }
      ];
  }
  ```

- [ ] Create `tests/sub/sub_test.go`:

  ```go
  package sub

  import "testing"

  func TestSubEntitySanitize(t *testing.T) {
      e := &SubEntity{Name: " <b>hi</b> "}
      e.Sanitize()
      if e.Name != "hi" {
          t.Fatalf("Sanitize: name = %q, want %q", e.Name, "hi")
      }
  }
  ```

- [ ] `make test` (in CI; no local protoc) produces
      `tests/sub/entity_sub.pb.go` and `tests/sub/entity_sub.pb.sanitize.go`,
      and the `tests/sub` package's test passes.

### Task 4 — `.gitignore`: cover nested generated files

- [ ] Replace flat globs with recursive globs:

  ```diff
  - tests/*.pb.go
  - tests/*.pb.sanitize.go
  + tests/**/*.pb.go
  + tests/**/*.pb.sanitize.go
  ```

  (Git treats `**` only between slashes; the top-level pattern still
  matches `tests/entity.pb.go` because `tests/**/x` matches
  `tests/x`. Tested in advance by running `git check-ignore` against
  both depths.)

### Task 5 — Commit, push, close PR #2

- [ ] Single commit on `feat/source-relative-output`:

  ```
  ✨ Source-relative output paths

  ...
  Co-authored-by: Matthew Cobbing <cobbinma@users.noreply.github.com>
  ```

- [ ] Open PR or merge to master per local convention.
- [ ] Comment on PR #2 referencing the new commit and close.

## Verification (end-to-end)

- [ ] `go build -mod=vendor ./...` — exit 0 locally.
- [ ] `golangci-lint run ./...` — exit 0 (after `make generate` runs in CI).
- [ ] CI `test` job — `make test` exits 0 with both top-level and
      `tests/sub` packages passing.
- [ ] `tests/sub/entity_sub.pb.sanitize.go` lands inside `tests/sub/`,
      not at the repo root or under `tests/`.

## Cross-references

- Generic spec: `docs/source-relative-output-spec.md` (applies to any
  protoc-gen-* plugin built on protoc-gen-star).
- Original PR: https://github.com/Intrinsec/protoc-gen-sanitize/pull/2
- `pgsgo.Context.OutputPath`:
  `vendor/github.com/lyft/protoc-gen-star/lang/go/package.go:49`
- `pgs.FilePath.SetExt`:
  `vendor/github.com/lyft/protoc-gen-star/name.go:171`
