# Onboard Precheck — protoc-gen-sanitize

## Goal

Re-runnable diagnostic. Run before and after correction plans to verify state.
Records the baseline captured during onboarding on 2026-05-20.

## Baseline (frozen 2026-05-20)

| Check | Result |
|-------|--------|
| Tool: `golangci-lint` | present, v2.10.1 |
| Tool: `govulncheck` | present, v1.2.0 (DB https://vuln.go.dev) |
| Tool: `go` | present, go1.26.2 linux/amd64 (module floor go1.18) |
| Tool: `protoc` | required for `make test`; verify with `protoc --version` |
| `golangci-lint run ./...` | **1 issue** — `typecheck` on `tests/entity_test.go` because generated `tests/*.pb.go` files are absent (Makefile generates them; they are not committed) |
| `govulncheck ./...` | **0 called** vulnerabilities. 3 imported + 7 required-module CVEs uncalled (informational, do not block merge) |
| `go build ./...` | success |
| `go test ./... -count=1 -short` | **build failed** — `tests/entity_test.go` references generated `Entity1..Entity9` not present. `make test` (which regenerates first) is the canonical entry point |
| `.golangci.yml` | missing |
| `vendor/` | missing |
| `.gitlab-ci.yml` / `.github/workflows/` | missing |
| `docs/DEVELOPMENT.md` | missing |
| `renovate.json` / `dependabot.yml` | missing |

## Tasks

### Task 1: Verify tooling installed

- [ ] `golangci-lint --version` → exit 0, version ≥ 2.0
- [ ] `govulncheck -version` → exit 0
- [ ] `go version` → exit 0, ≥ go1.22
- [ ] `protoc --version` → exit 0

### Task 2: Run all mandatory checks

Generate first so lint and tests have inputs:

- [ ] `make test` → exit 0 (this triggers `protoc` codegen for `tests/*.proto`,
      builds the plugin, runs it against fixtures, then runs `go test`).
      Record: pass/fail counts of `go test`.
- [ ] After `make test` has populated `tests/*.pb.go` and `tests/*.pb.sanitize.go`:
      `golangci-lint run ./...` → exit 0 (target: 0 issues).
      Record: issue count and top linter.
- [ ] `govulncheck ./...` → exit 0 (target: 0 called vulnerabilities).
      Record: called CVE count and uncalled counts.
- [ ] `go build ./...` → exit 0.

### Task 3: Report deltas

- [ ] Append the run result + delta vs baseline to this plan file under
      `## Run history`.

## Run history

<!-- Append a block per run -->

```
date: 2026-05-20 (initial baseline)
runner: claude (executing-plans)
golangci-lint issues: 1 (baseline 1) — codegen-first dependency confirmed
govulncheck called: 0 (baseline 0)
govulncheck uncalled: 3 imported + 7 modules (baseline 3+7)
go test pass/fail: build-failed (baseline build-failed) — same root cause
protoc: MISSING on PATH locally — `make test` not exercised in this run.
        Re-run after `apt install protobuf-compiler` (or equivalent).
notes: Baseline frozen as documented.
```

```
date: 2026-05-20 (post-correction gate sweep)
runner: claude (executing-plans)
build: go build -mod=vendor ./...                 → exit 0
vuln:  govulncheck ./...                          → 0 called, 3+6 uncalled
sbom:  make sbom                                  → sbom.cdx.json 23.6 KB
license-check: make license-check (go-licenses)   → exit 0 (Apache-2.0 / BSD-3-Clause / MIT)
gitleaks: gitleaks detect --config .gitleaks.toml → 0 leaks
golangci-lint config verify                       → exit 0 (no output = valid)
DEFERRED (protoc missing locally):
  - make generate (cannot codegen)
  - make test (depends on generate)
  - make lint (depends on generate)
notes: tools.go renamed → codegenruntime.go with `//go:build codegenruntime`
       tag to avoid collision with lyft/protoc-gen-star's tools tag.
       go.mod refreshed via `go mod tidy` (transitive bumps:
       google/go-cmp 0.5.8→0.6.0, golang.org/x/text 0.3.7→0.16.0,
       golang.org/x/tools 0.1.10→0.21.1, bluemonday 1.0.18→1.0.27).
       Runtime verification of these bumps deferred until protoc available.
```
