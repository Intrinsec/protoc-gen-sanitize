# Spec — oneof dispatch, WKT skip, pgs v2 migration

**Origin:** SIRP CR (Intrinsec / ia-gen-lab), 2026-05-29. Observed on `v0.0.15`.
**Target tag:** `v1.0.0` (framework major bump + behavioral fixes).

## Summary of findings (pre-design investigation)

Reproduced against current `master` (pgs `v0.6.2`):

| Bug | Status at HEAD | Action |
|-----|----------------|--------|
| 1 — oneof members referenced as direct fields | **broken** (reproduced) | fix codegen |
| 2 — `.Sanitize()` emitted on WKT pointers | **broken** (reproduced) | fix codegen |
| 3 — `paths=source_relative` flattens output | **already fixed** | add regression guard only |

Bug 3 was resolved before this CR by two independent prior changes already on
`master`: (a) the `pgsgo.Context.OutputPath(f).SetExt(...)` refactor documented
in `docs/source-relative-output-spec.md`, and (b) the pgs `v0.6.0 → v0.6.2`
sustainment bump. The CR observed `v0.0.15`, which predates both. Verified:
`paths=source_relative` on `sirp/v1/event.proto` emits
`sirp/v1/event.pb.sanitize.go` correctly.

## Scope decision

Fix Bugs 1 + 2, **and** migrate `lyft/protoc-gen-star v0.6.2` →
`lyft/protoc-gen-star/v2 v2.0.4`. Rationale: the v0 line is dormant (last tag
2022-12-13); v2 is the active line (2025-03-17). v2 brings **no** functional
gain for these bugs (they are template-logic, framework-agnostic) and is **not**
a stability upgrade (v2 carries the same "API UNSTABLE" README banner) — the
sole benefit is sitting on the maintained major. Accepted as deliberate
modernization churn.

## Bug 1 — oneof-aware dispatch

`protoc-gen-go` emits, per oneof, a wrapper interface field (`m.Value`), a getter
(`m.GetValue()`), and one wrapper struct per case (`Observable_Account`, …). The
oneof members are **not** direct fields on the parent. Current template iterates
`Message.Fields()` (which includes oneof members) and emits `m.Account.Sanitize()`
— uncompilable.

### Fix

Template iterates two disjoint sets instead of `Fields()`:

1. `Message.NonOneOfFields()` — flat sanitize, unchanged behavior.
2. `Message.RealOneOfs()` — excludes proto3-optional synthetic oneofs — emits a
   type switch per oneof.

Per oneof: accessor `m.Get<OneofName>()`, wrapper type name from
`pgsgo.Context.OneofOption(field)`:

```go
switch v := m.GetValue().(type) {
case *Observable_Account:
    if v.Account != nil {
        v.Account.Sanitize()
    }
case *Observable_Host:
    if v.Host != nil {
        v.Host.Sanitize()
    }
}
```

Per member field, by Go type:

- message + in-module (see Bug 2) → `if v.X != nil { v.X.Sanitize() }`
- message + WKT/external → case omitted
- string + sanitize rules → `v.X = textSanitize.Sanitize(v.X)` (+ `strings.TrimSpace(v.X)` when `trim`)
- string + `disable_field` → case omitted

Oneof members cannot be `repeated` (proto constraint) — no loop logic.
A case with no work is omitted. A type switch with no non-empty cases is omitted
entirely.

The `initializer` helper continues to scan `Message.Fields()` (all fields,
including oneof members) so `htmlSanitize` / `textSanitize` locals are declared
whenever any oneof string member needs them.

## Bug 2 — skip `.Sanitize()` on types without the method

`*timestamppb.Timestamp` and other WKT/external pointers have no `Sanitize()`
method. The plugin emits one only for messages it generates: those that are
**build targets** AND whose file is not `disable_file`.

### Fix

```go
func (p *SanitizeModule) hasSanitizeMethod(m pgs.Message) bool {
    return m.BuildTarget() && !p.fileDisabled(m.File())
}
```

`fileDisabled` reads the `sanitize.E_DisableFile` file extension. In the
`MessageT` branch of `sanitizer()` and inside oneof message-member handling: if
`embed := f.Type().Embed(); !hasSanitizeMethod(embed)` → skip (return `""`).
Applies to singular and repeated message fields alike.

`disable_message` messages keep an emitted (empty-bodied) `Sanitize()`, so they
remain callable — not skipped. Only WKT/external (`BuildTarget()==false`) and
`disable_file` targets are skipped. This matches the CR's robust option: "emit
`.Sanitize()` only on fields whose target has its own `Sanitize()`."

## Bug 3 — regression guard only

No code change. Add a Makefile smoke step (run by `make test`): invoke the plugin
with `--sanitize_opt=paths=source_relative` on a nested fixture into a temp dir,
assert output lands at the source-relative path (`tests/sub/entity_sub.pb.sanitize.go`),
fail otherwise.

## pgs v2 migration

- `go.mod`: `github.com/lyft/protoc-gen-star v0.6.2` → `github.com/lyft/protoc-gen-star/v2 v2.0.4`.
- Import-path swaps in `main.go`, `sanitizer.go`, `goimports_post_processor.go`:
  `pgs "github.com/lyft/protoc-gen-star/v2"`, `pgsgo "github.com/lyft/protoc-gen-star/v2/lang/go"`.
- Verify v2 API parity for: `pgs.Init`, `ModuleBase`, `Execute(map[string]pgs.File, map[string]pgs.Package)`,
  `Artifact` types (`GeneratorTemplateFile` etc.), `Extension`, `Name`, `OutputPath`,
  `NonOneOfFields`/`RealOneOfs`/`OneofOption`. Fix breaks as found.
- Re-check `goimports_post_processor.go` (lifted from pgs PR #96): confirm still
  needed on v2 (v2 may ship a goimports post-processor) and compiles.
- `go mod tidy && go mod vendor`; `govulncheck ./...` clean.

## Tests

- `tests/oneof.proto` — `Observable` oneof with message members (`Account`,
  `Host`) + a oneof carrying a `string` member with `kind:TEXT` + a oneof member
  of WKT type → exercises every Bug 1 path and the Bug 2 skip.
- `tests/wkt.proto` — `Event { google.protobuf.Timestamp detection_time = 1; }`
  → Bug 2; generated file must compile with no `.Sanitize()` on the WKT.
- `tests/oneof_test.go` — runtime assertions: message member sanitized, string
  member sanitized in place via wrapper, WKT untouched, nil oneof safe.
- WKT fixtures require the google well-known-type include path — Makefile
  `generate` step gains a portable `-I` resolution for it.

## Versioning / docs

- `CHANGELOG.md`: new `## [1.0.0]` block. Entries: Fixed (oneof, WKT), Changed
  (`**BREAKING:**` import-path major bump for library importers — note the
  `/v2` path). Plugin *binary* consumers are unaffected.
- README: no surface change expected; verify install/example still accurate.
- Tagging is a human action — spec prepares CHANGELOG only.

## Verification gates (per AGENTS.md)

- `make generate` — codegen succeeds on all fixtures.
- `make test` — all unit tests pass, incl. new oneof/wkt cases + Bug 3 smoke.
- `golangci-lint run ./...` clean.
- `govulncheck ./...` zero called vulns.
- `go build ./...` on generated fixtures.
