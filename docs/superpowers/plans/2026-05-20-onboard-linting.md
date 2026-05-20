# Onboard — Linting (.golangci.yml + codegen-first lint workflow)

## Goal

Establish a working `.golangci.yml` and a workflow that lints **after** the
`tests/*.pb.go` and `tests/*.pb.sanitize.go` files have been generated, so the
typecheck linter no longer fails on `undefined: EntityN`.

## Context

`AGENTS.md` section **Linting** requires `.golangci.yml` at repo root and a
zero-issue policy. Baseline (precheck plan, 2026-05-20) recorded 1 issue, all
on `tests/entity_test.go` because generated files are not committed.

The fix has two parts:
1. Add `.golangci.yml` excluding generated files from the noisy linters but
   keeping typecheck.
2. Define a `make lint` target that runs codegen **first**, then lint.

## File Structure

- Create: `.golangci.yml`
- Modify: `Makefile` (add `lint` target, ensure ordering)
- Test: re-run `golangci-lint run ./...` → 0 issues.

## Tasks

### Task 1: Create `.golangci.yml`

- [ ] Step 1: Run `/lint-go-config` and accept the tier-B default, OR write the
      following minimal config:

      ```yaml
      version: "2"

      run:
        timeout: 5m
        modules-download-mode: readonly
        tests: true

      issues:
        max-issues-per-linter: 0
        max-same-issues: 0

      linters:
        default: none
        enable:
          - errcheck
          - govet
          - ineffassign
          - staticcheck
          - unused
          - revive
          - gosec
          - misspell
          - unconvert
          - copyloopvar
        exclusions:
          paths:
            - "sanitize/sanitize.pb.go"
            - "tests/.*\\.pb\\.go$"
            - "tests/.*\\.pb\\.sanitize\\.go$"
          rules:
            - path: "tests/"
              linters:
                - gosec  # fixtures, not production code
      ```

- [ ] Step 2: Save as `/home/sml/Work/protoc-pluggins/protoc-gen-sanitize/.golangci.yml`.
- [ ] Step 3: `golangci-lint config verify` → exit 0.

### Task 2: Add `lint` Makefile target

- [ ] Step 1: Edit `Makefile`, append after the existing `test` target:

      ```makefile
      .PHONY: generate
      generate: bin/protoc-gen-go bin/protoc-gen-$(NAME)
      	@protoc -I . --plugin=protoc-gen-go=$(shell pwd)/bin/protoc-gen-go --go_out="." tests/*.proto
      	@protoc -I . --plugin=protoc-gen-$(NAME)=$(shell pwd)/bin/protoc-gen-$(NAME) --$(NAME)_out=tests tests/*.proto

      .PHONY: lint
      lint: generate
      	@golangci-lint run ./...
      ```

- [ ] Step 2: Refactor `test` target to depend on `generate` instead of inlining
      the `protoc` calls, so `make lint` and `make test` share the codegen step:

      ```makefile
      .PHONY: test
      test: generate
      	@cd tests && go test -v .
      ```

- [ ] Step 3: `make lint` → exit 0.

### Task 3: Verify zero issues

- [ ] `make generate && golangci-lint run ./...` → exit 0, **0** issues.

## Verification (end-to-end)

- [ ] Re-run `2026-05-20-onboard-precheck.md` Task 2 — `golangci-lint run ./...`
      exits 0 after `make generate`.
- [ ] `AGENTS.md` section **Linting** present and matches reality.
- [ ] `.golangci.yml` committed.

## Cross-references

- Skill: `isec-iagen_lint-go-config` for the canonical tier-B config.
- Skill: `isec-iagen_lint-go-install` if `golangci-lint` missing on a fresh
  developer machine.
- Related plan: `2026-05-20-onboard-ci-pipeline.md` — CI must call `make lint`,
  not `golangci-lint` directly.
