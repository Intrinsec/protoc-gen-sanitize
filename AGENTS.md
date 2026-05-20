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

- Go 1.18 minimum (raise floor to current stable when bumping deps).
- Avoid runtime reflection where compile-time type checking is possible.
- Generated test files (`tests/*.pb.go`, `tests/*.pb.sanitize.go`) must be
  produced by `make test` (or the CI codegen step) **before** running
  `go test ./...` or `golangci-lint run ./...` — both fail with `undefined:
  EntityN` otherwise. This is by design (we do not commit generated test
  artifacts), so every workflow that lints or tests must run codegen first.

## Workflow Skills

Mandatory skills to invoke before/during work on this repo:

- `superpowers:brainstorming` — for any new feature, option, or behavior change.
- `superpowers:writing-plans` — for any multi-step change. Plans live under
  `docs/superpowers/plans/`.
- `superpowers:test-driven-development` — write or update `tests/*.proto` +
  `entity_test.go` expectations before touching `sanitizer.go`.
- `superpowers:systematic-debugging` — when a generated file is wrong, do not
  patch the generator blindly; reproduce the codegen output against a minimal
  `.proto`, then trace the template.
- `superpowers:verification-before-completion` — `make test` must pass (lint +
  generate + go test) before claiming done. `golangci-lint run ./...` must be
  clean against generated code in place.
- `superpowers:requesting-code-review` — before merge to `master`.
- `caveman-commit` — commit messages.
- `caveman-review` — PR review comments.

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

## Vulnerability scanning

- `govulncheck ./...` must report zero **called** vulnerabilities.
- Uncalled CVEs in transitive modules are tracked but do not block merge —
  bump via Renovate when patch available.
- Wire to CI as a non-skippable job (see `docs/superpowers/plans/...
  -onboard-vuln-scanning.md`).

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

Carve-outs are honored by `dev-update-project`, `dev-arch-review`, and other
iagen-dev skills — they will not be re-suggested unless the user removes the row.
