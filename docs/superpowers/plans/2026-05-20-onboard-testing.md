# Onboard — Testing (codegen-first test workflow + coverage gate)

## Goal

Make `make test` the single canonical test entry point, ensure it generates
required `.pb.go` and `.pb.sanitize.go` fixtures before running `go test`, and
add a coverage gate that does not regress.

## Context

Baseline (2026-05-20):
- `tests/entity_test.go` references `Entity1..Entity9` defined in
  `tests/*.pb.go` produced by `make test`.
- `go test ./...` straight from a clean checkout fails with `undefined:
  EntityN`.
- No coverage measurement today.

## File Structure

- Modify: `Makefile` (already to be touched by `*-onboard-linting.md` and
  `*-onboard-vendoring.md`; this plan layers coverage on top).
- Create: `tests/cover.out` (gitignored).
- Modify: `.gitignore`.

## Tasks

### Task 1: Confirm codegen-first ordering after linting plan applies

- [ ] Precondition: `*-onboard-linting.md` Task 2 done (Makefile has
      `generate` and `test` depends on it).
- [ ] Step 1: From a clean checkout, `make distclean && make test` → exit 0.

### Task 2: Add coverage measurement

- [ ] Step 1: Edit `Makefile` `test` target to capture coverage:

      ```makefile
      .PHONY: test
      test: generate
      	@cd tests && go test -mod=vendor -v -coverprofile=cover.out -covermode=atomic .
      	@go tool cover -func=tests/cover.out | tail -1
      ```

- [ ] Step 2: Append to `.gitignore`:

      ```
      tests/cover.out
      ```

- [ ] Step 3: Run `make test` and record the **total** coverage figure as the
      floor.

### Task 3: Coverage gate in CI

- [ ] Step 1: In the CI plan, the `test` job stores `cover.out` as an
      artifact and parses the total. Below the floor → fail.
- [ ] Step 2: Defer the actual numeric floor decision to the CI plan because
      it depends on what `make test` reports after Task 2.

## Verification (end-to-end)

- [ ] `make distclean && make test` from a clean checkout exits 0 and prints
      a coverage line.
- [ ] `2026-05-20-onboard-precheck.md` Task 2 re-runs cleanly.

## Cross-references

- Plan: `2026-05-20-onboard-linting.md` — owns the `generate` target.
- Plan: `2026-05-20-onboard-ci-pipeline.md` — wires coverage gate.
- Skill: `superpowers:test-driven-development` — fixture-first when adding
  options.
