# oneof dispatch, WKT skip, pgs v2 migration — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix the `protoc-gen-sanitize` plugin so generated `Sanitize()` compiles for messages containing `oneof` and well-known-type (WKT) fields, and migrate the underlying framework from the dormant `lyft/protoc-gen-star` v0 to the active `/v2`.

**Architecture:** The plugin is a `protoc-gen-star` (pgs) module. `sanitizer.go` holds a `text/template` (`sanitizeTpl`) plus Go helper funcs registered into it; `main.go` wires the module + post-processors. We migrate imports to pgs `/v2` first (mechanical, verified API-parity), then add a guard so `.Sanitize()` is only emitted for messages this plugin actually generates a method for (skips WKT/external/`disable_file`), then restructure the template to dispatch oneof members through a type switch over the protoc-gen-go wrapper structs.

**Tech Stack:** Go 1.25, `github.com/lyft/protoc-gen-star/v2 v2.0.4`, `google.golang.org/protobuf`, `github.com/microcosm-cc/bluemonday`, `protoc` + `protoc-gen-go`.

**Pre-flight facts (verified during design, do not re-litigate):**
- pgs v2.0.4 has full API parity for everything used: `Execute(map[string]File, map[string]Package)`, `ModuleBase`, `AddGeneratorTemplateFile`, `Extension`, `Name`, `OutputPath`, `Message.NonOneOfFields()`, `Message.RealOneOfs()`, `OneOf.Fields()`, `Context.OneofOption()`, `Field.Type().Embed()` / `.Element().Embed()`, `Message.BuildTarget()`.
- v2 ships `pgsgo.GoImports()` byte-identical to the repo's `goimports_post_processor.go` (PR#96 upstreamed) → that file is deleted in Task 1.
- Bug 3 (`paths=source_relative`) already works on HEAD; Task 4 only adds a regression guard.
- Oneof members cannot be `repeated` (proto constraint), so oneof handling needs no loop logic.

---

## File Structure

- `go.mod` / `go.sum` / `vendor/` — dependency: swap pgs v0.6.2 → /v2 v2.0.4.
- `main.go` — import-path swap; use `pgsgo.GoImports()`.
- `goimports_post_processor.go` — **deleted** (superseded by `pgsgo.GoImports()`).
- `codegenruntime.go` — add `timestamppb` to the codegen-runtime dep pins so `vendor/` carries it for generated test code.
- `sanitizer.go` — import-path swap; add `embeddedMessage`, `fileDisabled`, `hasSanitizeMethod`, `oneofSanitizer`; guard the `MessageT` branch; restructure `sanitizeTpl`.
- `tests/wkt.proto` — new fixture (Bug 2).
- `tests/oneof.proto` — new fixture (Bug 1).
- `tests/wkt_test.go`, `tests/oneof_test.go` — new runtime tests.
- `Makefile` — `PROTOC_INCLUDE` var; add WKT include to `generate`; `test-source-relative` smoke target wired into `test`.
- `CHANGELOG.md` — `## [1.0.0]` block.

---

## Task 1: Migrate to protoc-gen-star /v2

**Files:**
- Modify: `main.go`
- Modify: `sanitizer.go:12-14` (import block)
- Delete: `goimports_post_processor.go`
- Modify: `go.mod`, `go.sum`, `vendor/` (via tooling)

- [ ] **Step 1: Swap imports in `main.go`**

Replace the import block and the post-processor registration:

```go
import (
	pgs "github.com/lyft/protoc-gen-star/v2"
	pgsgo "github.com/lyft/protoc-gen-star/v2/lang/go"
)

// version and commit are populated at build time by goreleaser via
// `-ldflags "-X main.version=... -X main.commit=..."`. Unused at the
// protoc protocol level (plugins have no `--version` flag), but kept so
// `strings <binary> | grep main.version` reveals the build provenance.
var (
	version = "dev"
	commit  = "none"
)

func main() {
	_, _ = version, commit

	sanitizeModule := Sanitize()

	pgs.Init(pgs.DebugEnv("DEBUG_PG_SAN")).
		RegisterModule(sanitizeModule).
		RegisterPostProcessor(pgsgo.GoFmt()).
		RegisterPostProcessor(pgsgo.GoImports()).
		Render()
	sanitizeModule.ExitCheck()
}
```

- [ ] **Step 2: Swap imports in `sanitizer.go`**

Change lines 12-14 from:

```go
	"github.com/intrinsec/protoc-gen-sanitize/sanitize"
	pgs "github.com/lyft/protoc-gen-star"
	pgsgo "github.com/lyft/protoc-gen-star/lang/go"
```

to:

```go
	"github.com/intrinsec/protoc-gen-sanitize/sanitize"
	pgs "github.com/lyft/protoc-gen-star/v2"
	pgsgo "github.com/lyft/protoc-gen-star/v2/lang/go"
```

- [ ] **Step 3: Delete the now-redundant post-processor**

Run:

```bash
git rm goimports_post_processor.go
```

(`pgsgo.GoImports()` from v2 is identical; the local copy lifted from pgs PR#96 is no longer needed.)

- [ ] **Step 4: Update module graph and vendor**

Run (the repo sets `GOFLAGS=-mod=vendor`; override to `-mod=mod` so tidy can rewrite):

```bash
GOFLAGS=-mod=mod go get github.com/lyft/protoc-gen-star/v2@v2.0.4
GOFLAGS=-mod=mod go mod tidy
GOFLAGS=-mod=mod go mod vendor
```

Expected: `go.mod` now requires `github.com/lyft/protoc-gen-star/v2 v2.0.4`; the old `github.com/lyft/protoc-gen-star v0.6.2` and any now-unused `github.com/golang/protobuf` direct require are dropped or demoted; `vendor/github.com/lyft/protoc-gen-star/v2/` exists.

- [ ] **Step 5: Build the plugin (compile check)**

Run:

```bash
make distclean && make build
```

Expected: builds `bin/protoc-gen-go` and `bin/protoc-gen-sanitize` with no compile errors.

- [ ] **Step 6: Run the existing suite — behavior unchanged on v2**

Run:

```bash
make test
```

Expected: all existing tests in `tests/` and `tests/sub/` PASS. Generated output (`tests/entity.pb.sanitize.go` etc.) is unchanged in shape from before the migration.

- [ ] **Step 7: Lint + vuln**

Run:

```bash
make lint
GOFLAGS=-mod=vendor govulncheck ./...
```

Expected: `golangci-lint` clean; `govulncheck` reports zero **called** vulnerabilities.

- [ ] **Step 8: Commit**

```bash
git add -A
git commit -m "build: migrate to protoc-gen-star/v2

Swap lyft/protoc-gen-star v0.6.2 (dormant since 2022) for /v2 v2.0.4.
v2 ships pgsgo.GoImports() identical to the local post-processor lifted
from pgs PR#96, so goimports_post_processor.go is removed. No behavior
change; existing fixtures regenerate identically."
```

---

## Task 2: Skip `.Sanitize()` on types without the method (Bug 2)

**Files:**
- Create: `tests/wkt.proto`
- Create: `tests/wkt_test.go`
- Modify: `sanitizer.go` (add helpers; guard `MessageT` branch)
- Modify: `codegenruntime.go` (pin `timestamppb` for vendoring)
- Modify: `Makefile` (`PROTOC_INCLUDE`, include path on `generate`)

- [ ] **Step 1: Add the WKT fixture**

Create `tests/wkt.proto`:

```proto
// Copyright Example 2026

syntax = "proto3";

package test;

import "google/protobuf/timestamp.proto";

option go_package = "./tests;test";

message Event {
    google.protobuf.Timestamp detection_time = 1;
}
```

- [ ] **Step 2: Add the well-known-type include path to the Makefile**

In `Makefile`, after the `GO_IMPORT` definition (around line 16), add:

```makefile
# protoc bundles the well-known types under <protoc>/../include. Fixtures that
# import google/protobuf/*.proto need this on the include path.
PROTOC_INCLUDE := $(shell dirname $(shell which protoc))/../include
```

Then update the `generate` target's two `protoc` invocations (around lines 43-44) to add `-I $(PROTOC_INCLUDE)`:

```makefile
.PHONY: generate
generate: bin/protoc-gen-go bin/protoc-gen-$(NAME)
	@protoc -I . -I $(PROTOC_INCLUDE) --plugin=protoc-gen-go=$(shell pwd)/bin/protoc-gen-go --go_out="." $(PROTO_FIXTURES)
	@protoc -I . -I $(PROTOC_INCLUDE) --plugin=protoc-gen-$(NAME)=$(shell pwd)/bin/protoc-gen-$(NAME) --$(NAME)_out=. $(PROTO_FIXTURES)
```

- [ ] **Step 3: Pin `timestamppb` so `vendor/` carries it for generated test code**

In `codegenruntime.go`, add the import (generated `tests/*.pb.go` for the WKT fixtures import `timestamppb`, but they don't exist at `go mod vendor` time — pinning here keeps the package in `vendor/`):

```go
import (
	_ "github.com/microcosm-cc/bluemonday"
	_ "google.golang.org/protobuf/types/known/timestamppb"
)
```

Then re-vendor:

```bash
GOFLAGS=-mod=mod go mod vendor
```

- [ ] **Step 4: Write the failing test**

Create `tests/wkt_test.go`:

```go
package test

import (
	"testing"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestEventSanitize_WKTUntouched(t *testing.T) {
	e := &Event{DetectionTime: timestamppb.New(time.Unix(123, 0))}
	e.Sanitize() // must compile (no .Sanitize() on the WKT) and not mutate it
	if e.DetectionTime == nil || e.DetectionTime.GetSeconds() != 123 {
		t.Fatalf("WKT field mutated or cleared: %v", e.DetectionTime)
	}
}
```

- [ ] **Step 5: Run to verify it fails (RED)**

Run:

```bash
make generate && (cd tests && GOFLAGS=-mod=vendor go build .)
```

Expected: FAIL — generated `tests/wkt.pb.sanitize.go` contains `m.DetectionTime.Sanitize()`, so the `tests` package fails to compile with `m.DetectionTime.Sanitize undefined (type *timestamppb.Timestamp has no field or method Sanitize)`.

- [ ] **Step 6: Add the guard helpers in `sanitizer.go`**

Add these helpers (place them just above `func (p *SanitizeModule) sanitizer`):

```go
// embeddedMessage returns the message a field embeds — the element message for
// repeated fields, the message itself for singular embeds — or nil if the field
// is not a message (or is a map value).
func embeddedMessage(f pgs.Field) pgs.Message {
	ft := f.Type()
	if ft.IsRepeated() {
		return ft.Element().Embed()
	}
	return ft.Embed()
}

// fileDisabled reports whether a file carries the sanitize.disable_file option.
func (p *SanitizeModule) fileDisabled(f pgs.File) bool {
	var disable bool
	ok, err := f.Extension(sanitize.E_DisableFile, &disable)
	return ok && err == nil && disable
}

// hasSanitizeMethod reports whether this plugin emits a Sanitize() method for m.
// A method is emitted only for build-target messages whose file is not
// disable_file'd. Well-known types and other external messages are never build
// targets, so .Sanitize() must not be called on them.
func (p *SanitizeModule) hasSanitizeMethod(m pgs.Message) bool {
	if m == nil {
		return false
	}
	return m.BuildTarget() && !p.fileDisabled(m.File())
}
```

- [ ] **Step 7: Guard the `MessageT` branch of `sanitizer`**

In `sanitizer`, change the `case pgs.MessageT` branch from:

```go
	case pgs.MessageT:
		return p.buildSanitizeCall(f, string(name), "", false)
```

to:

```go
	case pgs.MessageT:
		if !p.hasSanitizeMethod(embeddedMessage(f)) {
			return ""
		}
		return p.buildSanitizeCall(f, string(name), "", false)
```

- [ ] **Step 8: Run to verify GREEN**

Run:

```bash
make test
```

Expected: PASS, including `TestEventSanitize_WKTUntouched`. Inspect `tests/wkt.pb.sanitize.go` — `Event.Sanitize()` has only the `if m == nil { return }` body, no `.Sanitize()` on `DetectionTime`.

- [ ] **Step 9: Lint**

Run:

```bash
make lint
```

Expected: clean.

- [ ] **Step 10: Commit**

```bash
git add -A
git commit -m "fix: skip Sanitize() on well-known and external message fields

Generated code called .Sanitize() on *timestamppb.Timestamp and any
other message the plugin does not generate a method for, which does not
compile. Emit the call only when the target is a build-target message
whose file is not disable_file'd (i.e. has its own Sanitize())."
```

---

## Task 3: oneof-aware dispatch (Bug 1)

**Files:**
- Create: `tests/oneof.proto`
- Create: `tests/oneof_test.go`
- Modify: `sanitizer.go` (add `oneofSanitizer`; register func; restructure template)

- [ ] **Step 1: Add the oneof fixture**

Create `tests/oneof.proto` (exercises message members, an in-place string member, and a WKT member that must be skipped):

```proto
// Copyright Example 2026

syntax = "proto3";

package test;

import "sanitize/sanitize.proto";
import "google/protobuf/timestamp.proto";

option go_package = "./tests;test";

message Account {
    string name = 1 [
        (sanitize.rules) = {
            kind: TEXT,
            trim: true
        }
    ];
    string domain = 2;
}

message Host {
    string name = 1 [
        (sanitize.rules) = {
            kind: TEXT,
            trim: true
        }
    ];
}

message Observable {
    oneof value {
        Account account = 1;
        Host host = 2;
        string raw = 3 [
            (sanitize.rules) = {
                kind: TEXT,
                trim: true
            }
        ];
        google.protobuf.Timestamp seen_at = 4;
    }
}
```

- [ ] **Step 2: Write the failing test**

Create `tests/oneof_test.go`:

```go
package test

import (
	"testing"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestObservableSanitize(t *testing.T) {
	// message member: dispatched to Account.Sanitize()
	o := &Observable{Value: &Observable_Account{Account: &Account{Name: " <b>acct</b> "}}}
	o.Sanitize()
	if got := o.GetAccount().GetName(); got != "acct" {
		t.Fatalf("account name = %q, want %q", got, "acct")
	}

	// string member: sanitized in place on the wrapper
	o2 := &Observable{Value: &Observable_Raw{Raw: " <i>raw</i> "}}
	o2.Sanitize()
	if got := o2.GetRaw(); got != "raw" {
		t.Fatalf("raw = %q, want %q", got, "raw")
	}

	// WKT member: left untouched, no .Sanitize() emitted, no panic
	o3 := &Observable{Value: &Observable_SeenAt{SeenAt: timestamppb.New(time.Unix(7, 0))}}
	o3.Sanitize()
	if o3.GetSeenAt().GetSeconds() != 7 {
		t.Fatalf("seen_at mutated: %v", o3.GetSeenAt())
	}

	// nil oneof: safe
	(&Observable{}).Sanitize()
}
```

- [ ] **Step 3: Run to verify it fails (RED)**

Run:

```bash
make generate && (cd tests && GOFLAGS=-mod=vendor go build .)
```

Expected: FAIL — generated `tests/oneof.pb.sanitize.go` references `m.Account`, `m.Host`, etc. (non-existent direct fields), so `tests` fails to compile with `m.Account undefined (type *Observable has no field or method Account)`.

- [ ] **Step 4: Add `oneofSanitizer` to `sanitizer.go`**

Add this method (place it just above `func (p *SanitizeModule) sanitizer`, alongside the Task 2 helpers):

```go
// oneofSanitizer emits a type switch that dispatches sanitization to the active
// member of a (real) oneof. protoc-gen-go represents oneof members as wrapper
// structs (e.g. *Observable_Account) reached via m.Get<Oneof>(); they are not
// direct fields on the parent message. Members may be messages (call Sanitize)
// or strings with rules (sanitize in place on the wrapper). Members that are
// disabled, ruleless strings, or types without a Sanitize() method (WKT/
// external) contribute no case. An all-empty switch is omitted entirely.
func (p *SanitizeModule) oneofSanitizer(o pgs.OneOf) string {
	var cases []string

	for _, f := range o.Fields() {
		name := p.ctx.Name(f)

		var disableField bool
		if ok, err := f.Extension(sanitize.E_DisableField, &disableField); ok && err == nil && disableField {
			continue
		}

		access := fmt.Sprintf("v.%s", name)
		var body string

		switch f.Type().ProtoType() {
		case pgs.StringT:
			var rules sanitize.FieldRules
			ok, err := f.Extension(sanitize.E_Rules, &rules)
			if err != nil {
				p.Logf(
					"%v:%d: Error can't retrieve rules extension for message %s with error: %s",
					f.File().Name(),
					f.SourceCodeInfo().Location().Span[0]+1,
					f.FullyQualifiedName(),
					err,
				)
				p.hasErrors = true
				continue
			}
			if !ok {
				continue
			}
			var kind string
			switch rules.Kind {
			case sanitize.Sanitization_HTML:
				kind = "html"
			case sanitize.Sanitization_TEXT:
				kind = "text"
			default:
				continue
			}
			lines := []string{fmt.Sprintf("%s = %sSanitize.Sanitize(%s)", access, kind, access)}
			if rules.GetTrim() {
				lines = append(lines, fmt.Sprintf("%[1]s = strings.TrimSpace(%[1]s)", access))
			}
			body = strings.Join(lines, "\n")
		case pgs.MessageT:
			if !p.hasSanitizeMethod(embeddedMessage(f)) {
				continue
			}
			body = fmt.Sprintf("if %[1]s != nil {\n%[1]s.Sanitize()\n}", access)
		default:
			continue
		}

		cases = append(cases, fmt.Sprintf("case *%s:\n%s", p.ctx.OneofOption(f), body))
	}

	if len(cases) == 0 {
		return ""
	}

	return fmt.Sprintf("switch v := m.Get%s().(type) {\n%s\n}", p.ctx.Name(o), strings.Join(cases, "\n"))
}
```

(Formatting need not be tidy — `pgsgo.GoFmt()` + `pgsgo.GoImports()` post-process the output.)

- [ ] **Step 5: Register the template func**

In `InitContext`, add `"oneofSanitizer": p.oneofSanitizer,` to the `Funcs` map:

```go
	tpl := template.New("Sanitize").Funcs(map[string]interface{}{
		"package":           p.ctx.PackageName,
		"name":              p.ctx.Name,
		"sanitizer":         p.sanitizer,
		"oneofSanitizer":    p.oneofSanitizer,
		"initializer":       p.initializer,
		"leadingCommenter":  p.leadingCommenter,
		"isDisabledMessage": p.isDisabledMessage,
		"checkNoSanitize":   p.checkNoSanitize,
	})
```

- [ ] **Step 6: Restructure the template**

In `sanitizeTpl`, replace the single field loop in the enabled-message branch. Change:

```
	{{ initializer . }}

{{ range .Fields }}
    {{ sanitizer . }}
{{ end }}
{{ end }}
```

to:

```
	{{ initializer . }}

{{ range .NonOneOfFields }}
    {{ sanitizer . }}
{{ end }}
{{ range .RealOneOfs }}
    {{ oneofSanitizer . }}
{{ end }}
{{ end }}
```

(The `disable_message` branch keeps `{{ range .Fields }}{{ checkNoSanitize . }}{{ end }}` unchanged. `initializer` still scans `Message.Fields()`, which includes oneof members, so `textSanitize`/`htmlSanitize` locals are declared whenever a oneof string member needs them.)

- [ ] **Step 7: Run to verify GREEN**

Run:

```bash
make test
```

Expected: PASS, including `TestObservableSanitize`. Inspect `tests/oneof.pb.sanitize.go` — `Observable.Sanitize()` contains a `switch v := m.GetValue().(type)` with cases for `*Observable_Account`, `*Observable_Host`, `*Observable_Raw`, and **no** case for `seen_at` (the Timestamp WKT).

- [ ] **Step 8: Lint**

Run:

```bash
make lint
```

Expected: clean.

- [ ] **Step 9: Commit**

```bash
git add -A
git commit -m "fix: dispatch oneof members through a type switch

Generated code referenced oneof members as direct fields on the parent
(m.Account), which does not compile. Emit a type switch over
m.Get<Oneof>() with one case per wrapper struct, dispatching .Sanitize()
to message members and in-place sanitization to string members. WKT/
external and disabled members are skipped; an empty switch is omitted."
```

---

## Task 4: source_relative regression guard (Bug 3)

**Files:**
- Modify: `Makefile` (add `test-source-relative`; wire into `test`)

- [ ] **Step 1: Add the smoke target**

In `Makefile`, add a `test-source-relative` target and make `test` depend on it. Replace the existing `test` target:

```makefile
.PHONY: test
test: generate test-source-relative
	@cat tests/entity.pb.$(NAME).go
	@cd tests && go test -mod=vendor -v -coverprofile=cover.out -covermode=atomic .
	@go tool cover -func=tests/cover.out | tail -1

.PHONY: test-source-relative
test-source-relative: bin/protoc-gen-$(NAME)
	@tmp=$$(mktemp -d); \
	protoc -I . -I $(PROTOC_INCLUDE) \
	  --plugin=protoc-gen-$(NAME)=$(shell pwd)/bin/protoc-gen-$(NAME) \
	  --$(NAME)_out=$$tmp --$(NAME)_opt=paths=source_relative tests/sub/entity_sub.proto; \
	if [ -f $$tmp/tests/sub/entity_sub.pb.$(NAME).go ]; then \
	  echo "source_relative OK: nested path tests/sub/ preserved"; \
	  rm -rf $$tmp; \
	else \
	  echo "source_relative FAIL: expected tests/sub/entity_sub.pb.$(NAME).go under $$tmp"; \
	  find $$tmp -type f; rm -rf $$tmp; exit 1; \
	fi
```

- [ ] **Step 2: Run the smoke target in isolation**

Run:

```bash
make test-source-relative
```

Expected: prints `source_relative OK: nested path tests/sub/ preserved` and exits 0. (`tests/sub/entity_sub.proto` with `paths=source_relative` and `-I .` must land output at `tests/sub/entity_sub.pb.sanitize.go`, not flat.)

- [ ] **Step 3: Run the full suite**

Run:

```bash
make test
```

Expected: smoke passes, then all unit tests pass.

- [ ] **Step 4: Commit**

```bash
git add Makefile
git commit -m "test: guard paths=source_relative nested output

Locks the already-fixed source_relative behavior: regenerating
tests/sub/entity_sub.proto with paths=source_relative must preserve the
tests/sub/ directory in the output path rather than flattening it."
```

---

## Task 5: Changelog, docs, final verification

**Files:**
- Modify: `CHANGELOG.md`
- Verify: `README.md`

- [ ] **Step 1: Read the current CHANGELOG top**

Run:

```bash
sed -n '1,30p' CHANGELOG.md
```

Confirm the `## [Unreleased]` block location and the compare-link section at the bottom.

- [ ] **Step 2: Add the 1.0.0 release block**

Rename `## [Unreleased]` to a fresh empty block plus a new `## [1.0.0] - 2026-05-29` block beneath it. The `1.0.0` block content:

```markdown
## [1.0.0] - 2026-05-29

### Fixed
- Generated `Sanitize()` now compiles for messages containing a `oneof`: oneof
  members are reached through a type switch instead of as direct fields.
- Generated `Sanitize()` no longer calls `.Sanitize()` on well-known-type fields
  (e.g. `google.protobuf.Timestamp`) or other messages the plugin does not
  generate a method for, which previously failed to compile.

### Changed
- **BREAKING:** The plugin is now built on `protoc-gen-star/v2`. This affects
  only code that imports this repository as a Go library (the module's internal
  framework changed); the `protoc-gen-sanitize` binary and its plugin options
  are unchanged. Plugin users need no migration.
```

- [ ] **Step 3: Update the compare links at the bottom of `CHANGELOG.md`**

Add/adjust the link references so `[Unreleased]` compares against `v1.0.0` and `[1.0.0]` compares against the previous tag. Match the existing link style in the file (read the bottom of the file first to mirror the exact URL format).

- [ ] **Step 4: Verify README accuracy**

Run:

```bash
sed -n '1,80p' README.md
```

Confirm install steps, the `go install` line, and any usage/option examples still match shipped reality (no behavioral option changed; the only change is internal). Fix any stale references inline if found.

- [ ] **Step 5: Full verification sweep**

Run:

```bash
make distclean && make test && make lint && GOFLAGS=-mod=vendor govulncheck ./...
```

Expected: build + all tests pass (incl. WKT, oneof, source_relative smoke); lint clean; `govulncheck` zero called vulns.

- [ ] **Step 6: Commit**

```bash
git add CHANGELOG.md README.md
git commit -m "docs: changelog for v1.0.0 (oneof + WKT fixes, pgs v2)"
```

- [ ] **Step 7: Hand back for review / tagging**

Do **not** tag. Report completion with the verification output and note that tagging `v1.0.0` is the maintainer's action. Recommend `superpowers:requesting-code-review` before merge per AGENTS.md.

---

## Self-Review

**Spec coverage:**
- Bug 1 (oneof) → Task 3. ✓ (message members, string members, WKT-member skip, nil safety, synthetic-oneof exclusion via `RealOneOfs`).
- Bug 2 (WKT skip) → Task 2. ✓ (`hasSanitizeMethod` = `BuildTarget() && !disable_file`, singular + repeated via `embeddedMessage`).
- Bug 3 (source_relative) → Task 4 regression guard. ✓ (already fixed, per spec).
- pgs v2 migration → Task 1. ✓ (imports, `GoImports` dedupe, vendor, govulncheck).
- Tests (`oneof.proto`, `wkt.proto`, runtime tests, WKT include path) → Tasks 2-3 + Makefile edits. ✓
- Versioning / CHANGELOG / `**BREAKING:**` note → Task 5. ✓

**Placeholder scan:** No TBD/TODO/"handle edge cases". All code blocks complete; all commands have expected output.

**Type consistency:** Helper names consistent across tasks — `embeddedMessage`, `fileDisabled`, `hasSanitizeMethod`, `oneofSanitizer` defined in Task 2/3 and referenced consistently. Template funcs `sanitizer`/`oneofSanitizer`/`initializer` match registration. `pgsgo.GoImports` used in Task 1 matches the deletion rationale. Accessor `m.Get<Oneof>()` and wrapper `p.ctx.OneofOption(f)` match protoc-gen-go output verified in design.

**Note on maps:** message-valued *map* fields (none in fixtures) previously emitted broken `m.X.Sanitize()`; with the Task 2 guard, `embeddedMessage` returns nil for them → skipped (no emission). This is a latent-bug improvement, not a regression. Documented here, not separately tested.
