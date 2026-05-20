# AGENTS.md — protoc-gen-sanitize

Project tier and type drive every rule below.

| Key | Value |
|-----|-------|
| TIER | B — shared non-critical (developer build tool, no SLA, used across Intrinsec Go projects) |
| TYPE | 2 — Go CLI (protoc plugin invoked at build time) |
| API_STYLE | n/a (codegen plugin, not an API server) |
| Go version | 1.18 (declared in `go.mod`) |
| Module path | `github.com/intrinsec/protoc-gen-sanitize` |

## Purpose

`protoc-gen-sanitize` is a `protoc` plugin (built on `protoc-gen-star`) that
generates sanitization methods on Go structs produced from `.proto` files. The
plugin reads sanitization options declared in `sanitize/sanitize.proto`, walks
the message tree, and emits `*.pb.sanitize.go` files alongside the standard
`*.pb.go` outputs.

A `strict` mode (`--sanitize_opt=strict`) makes the plugin fail when a string
field has no explicit sanitization option, so CI/CD pipelines that depend on
`protoc-gen-sanitize` can prevent unsanitized fields from slipping through.

## Language

Default agent response: English, even if user writes French.
French response only when user explicitly asks French.

Caveman compression **mandatory** for all conversational responses (default level: `full`).
Code blocks, commit messages, PR descriptions, security warnings, irreversible-action
confirmations stay normal prose (Caveman auto-clarity rules). Do not disable Caveman unless
user says "stop caveman" or "normal mode".

HARD RULE: all code, comments, identifiers, doc strings, commit messages, ADRs, technical
docs in English — every project type, regardless of team spoken language.
User-facing strings + UI copy exempt — match audience language.

**Project specifics:**

- Go 1.18 minimum (raise floor to current stable when bumping deps).
- Avoid runtime reflection where compile-time type checking is possible.
- Generated test files (`tests/*.pb.go`, `tests/*.pb.sanitize.go`) must be
  produced by `make test` (or the CI codegen step) **before** running
  `go test ./...` or `golangci-lint run ./...` — both fail with `undefined:
  EntityN` otherwise. This is by design (we do not commit generated test
  artifacts), so every workflow that lints or tests must run codegen first.

## Workflow Skills (mandatory)

Every agent session in this repo must load + apply these skill packs:

- **superpowers** — process discipline (`brainstorming`, `writing-plans`, `executing-plans`,
  `test-driven-development`, `systematic-debugging`, `verification-before-completion`,
  `requesting-code-review`).
- **caveman** — response compression (see Language section).

Pack missing? Install per iagen-dev `INSTALL.md` before work.

### Session start gate

Before any response, clarification, repository inspection, shell command, or file edit:
run `superpowers:using-superpowers` first, then run `caveman` so compression is active
for every response. Use `superpowers:using-superpowers` to decide which additional
skills apply, then follow the selected skill workflows.

### Plan-writing mandatory before non-trivial implementation

Any feature, refactor, bugfix touching more than one function, or agent cannot reason in
one pass:

1. Run `superpowers:brainstorming` — clarify intent + requirements.
2. Run `superpowers:writing-plans` — persist plan at `docs/superpowers/plans/<short-name>.md`
   (commit to git).
3. Execute via `superpowers:executing-plans` (single-session) or
   `superpowers:subagent-driven-development` (parallelisable steps).
4. Gate completion with `superpowers:verification-before-completion` — no "done" claim
   without evidence (test output, lint output, build output).

**Trivial edits exception:** typos, single-line config tweaks, self-evident one-liners
skip steps 1–3 but still verify before claiming done.

### Bug fixes go through systematic-debugging

Any bug, failing test, unexpected behaviour → `superpowers:systematic-debugging` first.
No symptom patching without root cause.

### Code review before merge

Before merge or PR for non-trivial work: run `superpowers:requesting-code-review`.

### Project-specific skill mapping

- `superpowers:test-driven-development` — write or update `tests/*.proto` +
  `entity_test.go` expectations before touching `sanitizer.go`.
- `caveman-commit` — commit messages.
- `caveman-review` — PR review comments.

## Code Quality

After modifying any Go file: run `golangci-lint run ./...` before marking work complete.
Fix all lint errors, re-run until clean. Lint errors = task not done.
`gofmt` non-negotiable — zero diff allowed. Run `gofmt -w .` if in doubt.

In this repo, `make lint` runs `make generate` first so the linter has the codegen
output in place — see `## Linting` below for the codegen-first ordering rationale.

## Generated Code

Generated sources are **read-only**. Never hand-edit files produced by a code generator:

- `protoc` / `buf` outputs for gRPC + Protobuf (typically `*.pb.go`, `*_grpc.pb.go`,
  often under `gen/`, `pb/`, or `proto/`)
- `mockgen` / `moq` mocks
- `sqlc`, `ent`, `gqlgen`, `wire_gen.go`, `oapi-codegen`, `swag` outputs
- any file with a `// Code generated ... DO NOT EDIT.` header

To change generated code: change the source of truth (`.proto`, `.sql`, schema, interface)
then re-run the generator (`buf generate`, `go generate ./...`, `sqlc generate`, etc.).
Commit the regenerated files alongside the source change in the same commit.

**This project's particularity:** `protoc-gen-sanitize` IS a code generator. Two
separate categories of generated code exist:

- `sanitize/sanitize.pb.go` — generated from `sanitize/sanitize.proto`, **committed**
  (consumers import it).
- `tests/*.pb.go`, `tests/*.pb.sanitize.go` — generated by `make test`, **NOT
  committed** (regenerated on every run; tests rely on them being fresh).

Modifying the plugin's codegen templates (in `sanitizer.go`) is a source-of-truth
edit, not a generated-code edit — covered by normal lint + test rules.

## Error Handling

"Crash early, let orchestrator recover" model:
- Transient errors (network, timeout): retry 1–3× exponential backoff, log each retry
  at WARN. Retries exhausted → log ERROR with full context, exit non-zero.
- Structural errors (missing config, unavailable critical dep): crash immediately at
  startup. No retry.
Never swallow errors silently. Every error includes enough context for diagnosis
without accessing running pod.

**For this plugin:** errors during codegen must surface back to `protoc` via the
plugin protocol (return non-zero, write diagnostic to stderr). Strict mode
(`--sanitize_opt=strict`) escalates "missing sanitization option" warnings to
errors so CI/CD pipelines fail fast on unsanitized fields.

## Documentation Coherence

After any meaningful change (feature, bugfix touching public behaviour, API surface,
config schema, CLI flags, deps with user impact): verify `README.md` + `docs/**` still
match shipped reality before mark task done. Out-of-sync doc = task not done.

Pre-release sweep (mandatory before every tag, all release levels):
- README accurate — install steps, quickstart, examples runnable as-is.
- All in-repo doc links + references resolve (no dead anchors, no stale paths).
- Public API docs match shipped surface (endpoints, flags, env vars).
- Migration notes present for breaking changes.
- Screenshots / diagrams reflect current UI + architecture.
- `CHANGELOG.md` matches release scope (see Changelog section).

Release blocked if sweep fails. Doc fix = same MR as code change, never separate
follow-up.

## Changelog

Maintain user-friendly `CHANGELOG.md` at repo root. Format: Keep-a-Changelog
(https://keepachangelog.com/en/1.1.0/) + SemVer.

Every user-visible change → entry under `## [Unreleased]` with one type:
`Added` / `Changed` / `Deprecated` / `Removed` / `Fixed` / `Security`.

Wording rules (user-facing, not commit log):
- End-user perspective. No commit hash, no internal module name, no implementation
  detail.
- Bad: "refactor auth middleware to use JWT v2 lib".
- Good: "Sessions survive backend restarts; existing tokens stay valid".
- Breaking changes prefix `**BREAKING:**` + 1-3 line migration note.

Release cut process:
1. Rename `## [Unreleased]` → `## [X.Y.Z] - YYYY-MM-DD`.
2. Create fresh empty `## [Unreleased]` block at top.
3. Tag matches header version exactly (`vX.Y.Z`).
4. Bump compare links at file bottom.

Internal-only changes (refactor with zero user impact, test infra, CI config) skip
CHANGELOG entry — but if in doubt, log under `Changed`.

## Tooling

Required on developer machine and in CI:

- `go` ≥ 1.22 (build), 1.18 module floor.
- `protoc` (binary) — codegen.
- `protoc-gen-go` (vendored via `make bin/protoc-gen-go`).
- `golangci-lint` ≥ v2 — installed via `/lint-go-install`.
- `govulncheck` — installed via `/govulncheck-install`.

## Linting

- `.golangci.yml` lives at repo root. Use `/lint-go-config` to refresh.
- Lint runs **after** `make test` (or its codegen step) so generated files are
  present. Running lint on a fresh checkout without codegen will fail on
  `tests/entity_test.go` with `package test` typecheck errors.
- Zero-issue policy on tracked code (generated `*.pb.*.go` files excluded via
  `issues.exclude-files` or `path-except`).

## Vulnerability Scanning

After modifying `go.mod` / `go.sum`: run `govulncheck ./...` before marking work complete.
(`vendor/` is honored via `GOFLAGS=-mod=vendor` in env; govulncheck has no `-mod` CLI flag.)
Fix called vulns: `go get <module>@<fixed>`, `go mod tidy`, re-vendor if applicable,
re-run until clean. Imported-only vulns: report to user.
Called vulns remaining = task not done.

**Project specifics:**

- `govulncheck ./...` must report zero **called** vulnerabilities.
- Uncalled CVEs in transitive modules are tracked but do not block merge —
  bump via Renovate when patch available.
- Wired to CI as a non-skippable job (`.github/workflows/ci.yml` `vuln` job).

## Vendoring

Tier B Go projects must commit `vendor/`. This repo currently does not — see
correction plan `*-onboard-vendoring.md`.

Once vendored:
- `go mod tidy && go mod vendor` after every dependency change.
- CI must build with `-mod=vendor`.

## Testing

- Unit tests: `tests/entity_test.go` against the live codegen output for every
  `tests/*.proto` fixture.
- New behavior → add a `tests/<feature>.proto` fixture + a test case in
  `entity_test.go` (or a new `_test.go` for clarity).
- `make test` is the canonical local entry point.

## CI / CD

This repo has no CI pipeline yet. Tier B requires at minimum:

- lint
- vet / build
- generate + test
- govulncheck
- SBOM
- secret scan
- release (tag-triggered) producing checksummed + signed binaries.

See `*-onboard-ci-pipeline.md` for the GitLab CI bootstrap (skill:
`gitlab-cicd-go`). If the repo lives on GitHub instead of GitLab, mirror the
job set in `.github/workflows/`.

## Releases

- Tag scheme: `vMAJOR.MINOR.PATCH`, Conventional Commits feeding the changelog.
- Release artifact: `protoc-gen-sanitize` binary for `linux/amd64`,
  `linux/arm64`, `darwin/amd64`, `darwin/arm64`.
- Each artifact accompanied by:
  - `SHA256SUMS` (and `.minisig` or cosign signature — see release-signing
    correction plan).
  - `sbom.cdx.json` (CycloneDX).

## Dependency policy

- Renovate or Dependabot configured. PRs auto-rebase, manual merge.
- `go.mod` currently pinned to **old** versions (2022 era). Plan
  `*-onboard-dep-policy.md` upgrades + sets up the bot.

## Secret scanning

- gitleaks pre-commit hook + CI job. Baseline scan committed to
  `.gitleaks.toml` (allowlist).

## License compliance

- All direct deps must be in the allowlist (`Apache-2.0`, `MIT`, `BSD-2-Clause`,
  `BSD-3-Clause`, `ISC`, `MPL-2.0`). CI gate via `syft`/`fossa`.

## Repository structure

```
.
├── main.go                       # protoc plugin entrypoint
├── sanitizer.go                  # codegen logic (templates)
├── goimports_post_processor.go   # post-processor for generated files
├── sanitize/
│   ├── sanitize.proto            # plugin options schema
│   └── sanitize.pb.go            # generated, committed
├── tests/
│   ├── *.proto                   # fixtures
│   ├── entity_test.go            # depends on generated *.pb.go (not committed)
│   └── *.pb.go, *.pb.sanitize.go # generated by `make test`, NOT committed
├── Makefile
├── AGENTS.md
└── docs/
    └── superpowers/plans/        # iagen-dev correction plans
```

## Carve-outs

The following standard sections are explicitly excluded from this project.
Re-evaluate on tier change or quarterly review.

| Section | Reason | Note |
|---------|--------|------|
| Monitoring / alerts / dashboards | `not-applicable` | Build-time codegen plugin, no runtime to instrument. Process exits after `protoc` invocation. |
| Authentication (Keycloak) | `not-applicable` | No user-facing surface, no auth. |
| Secrets management (Vault AppRole) | `not-applicable` | No runtime secrets. Plugin only reads `.proto` inputs and writes `.go` outputs. |
| Database (CNPG manifest, backups) | `not-applicable` | No persistent state. |
| API contract (REST/gRPC/OpenAPI/proto schema) | `not-applicable` | Plugin consumes `.proto` but does not serve an API. `sanitize/sanitize.proto` is an options schema, not an external contract. |
| Container hardening (distroless, non-root, healthcheck) | `not-applicable` | No `Dockerfile`. Plugin is consumed as a binary on developer/CI hosts via `go install`. Re-evaluate if we ever publish a `protoc-gen-sanitize` OCI image. |
| Mobile (TYPE=6) and frontend (TYPE=3) tooling | `not-applicable` | Pure Go CLI. |
| Local Development (`docker compose up`) | `not-applicable` | Build-time plugin invoked by `protoc`. No services, no DB, no cache to stand up. `make test` covers the entire local loop. |
| Dependency Injection (constructor DI, Wire, Dig) | `not-applicable` | Single-binary plugin with one entrypoint (`main.go`) and one codegen module (`sanitizer.go`). No service graph to wire. |
| Logging (`slog` JSON, structured fields) | `not-applicable` | Plugin runs synchronously inside `protoc` and exits. Diagnostics go to stderr via `protoc-gen-star` (`DEBUG_PG_SAN=1`). No log shipping. |
| Testing & Architecture template (Red-Green-Refactor + edge I/O + interface-at-call-site) | `replaced-by:fixture-driven codegen tests` | Project is a code generator. Tests live in `tests/*.proto` fixtures + `tests/entity_test.go` asserting against the generated output. Standard service-DDD test layering doesn't fit. |
| Project Layout template (`internal/` four-layer split) | `replaced-by:flat-cli-layout` | Plugin total ~3 Go files plus the `sanitize/` proto package. The template's `cmd/internal/domain/usecase/repository/delivery` split would be 100% empty directories. Actual layout documented in `## Repository structure` above. |

Carve-outs are honored by `dev-update-project`, `dev-arch-review`, and other
iagen-dev skills — they will not be re-suggested unless the user removes the row.
