# Onboard — Vendoring (commit vendor/ for reproducible builds)

## Goal

Add `vendor/` so tier B builds are reproducible without network access and
match the iagen-dev standard for Go projects.

## Context

`AGENTS.md` section **Vendoring** requires tier B Go projects to commit
`vendor/`. The repo currently lacks it. CI must subsequently build with
`-mod=vendor`.

`go.mod` declares Go 1.18 with old deps (2022 vintage). Bumping deps is a
**separate** plan (`...-onboard-dep-policy.md`); this plan vendors the current
state to capture a stable baseline first.

## File Structure

- Modify: `.gitignore` (ensure `vendor/` is NOT ignored — current `.gitignore`
  is clean on that point).
- Create: `vendor/` (large, committed).
- Modify: `Makefile` — build/install targets use `-mod=vendor`.

## Tasks

### Task 1: Tidy then vendor

- [ ] Step 1: `go mod tidy` → exit 0. Commit any `go.mod` / `go.sum` changes.
- [ ] Step 2: `go mod vendor` → exit 0. Produces `vendor/modules.txt` and the
      full dependency tree.
- [ ] Step 3: `git add vendor/` → confirm size reasonable (< ~50 MB for this
      dependency set). If unexpectedly large, investigate before committing.

### Task 2: Switch build to `-mod=vendor`

- [ ] Step 1: Edit `Makefile`:
      - Replace `go install -v .` (line 21) with
        `go install -mod=vendor -v .`
      - Replace `GOBIN=$(shell pwd)/bin go install .` (line 35) with
        `GOBIN=$(shell pwd)/bin go install -mod=vendor .`
      - Replace `cd tests && go test -v .` with
        `cd tests && go test -mod=vendor -v .`
- [ ] Step 2: `make distclean && make test` → exit 0 (full rebuild from
      vendored deps).

### Task 3: Guardrail

- [ ] Step 1: Add to `Makefile`:

      ```makefile
      .PHONY: vendor-check
      vendor-check:
      	@go mod tidy
      	@go mod vendor
      	@git diff --exit-code -- go.mod go.sum vendor/ \
      		|| (echo "vendor/ out of date — run 'go mod tidy && go mod vendor'"; exit 1)
      ```

- [ ] Step 2: Add `vendor-check` to the CI pipeline (see CI plan).

## Verification (end-to-end)

- [ ] `vendor/modules.txt` present.
- [ ] `go build -mod=vendor ./...` exits 0 with `GOFLAGS=-mod=vendor`
      and no network.
- [ ] `make vendor-check` clean.
- [ ] `2026-05-20-onboard-precheck.md` Task 2 re-runs cleanly with vendor in
      place.

## Cross-references

- Plan: `2026-05-20-onboard-dep-policy.md` — handles bumping deps; this plan
  freezes the current state.
- Plan: `2026-05-20-onboard-ci-pipeline.md` — uses `vendor-check` job.
