# Onboard — Developer guide (docs/DEVELOPMENT.md)

## Goal

Write `docs/DEVELOPMENT.md` so a new contributor can clone, build, generate,
test, lint, and submit a change without reading the source.

## Context

`AGENTS.md` references the standard iagen-dev developer guide. Repo currently
has only `README.md` (28 lines, parameters + tests pointer). The `Makefile`
encodes the workflow but is not human documentation.

## File Structure

- Create: `docs/DEVELOPMENT.md`.

## Tasks

### Task 1: Write `docs/DEVELOPMENT.md`

- [ ] Step 1: Save the following content to
      `/home/sml/Work/protoc-pluggins/protoc-gen-sanitize/docs/DEVELOPMENT.md`:

      ```markdown
      # Development — protoc-gen-sanitize

      ## Prerequisites

      - Go ≥ 1.22 (module floor 1.18).
      - `protoc` (Protocol Buffers compiler).
      - `golangci-lint` ≥ v2 — install via `/lint-go-install` or
        `go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest`.
      - `govulncheck` — install via `/govulncheck-install` or
        `go install golang.org/x/vuln/cmd/govulncheck@latest`.

      ## First-time setup

      ```bash
      git clone <repo>
      cd protoc-gen-sanitize
      make distclean    # remove any stale generated artifacts
      make test         # codegen + build + test, end-to-end
      ```

      ## Daily commands

      | Command | What it does |
      |---|---|
      | `make generate` | Regenerate `tests/*.pb.go` and `tests/*.pb.sanitize.go` |
      | `make test` | `generate` → `go test` with coverage |
      | `make lint` | `generate` → `golangci-lint run ./...` |
      | `make vuln` | `govulncheck ./...` |
      | `make vendor-check` | Verify `vendor/` matches `go.mod` |
      | `make distclean` | Remove `bin/`, generated files, generated proto |

      ## Adding a new sanitization option

      1. Add the option to `sanitize/sanitize.proto`.
      2. `make generate` (regenerates `sanitize/sanitize.pb.go`).
      3. Add a fixture under `tests/<feature>.proto`.
      4. Add a test case to `tests/entity_test.go`.
      5. Implement codegen in `sanitizer.go` (template + walking logic).
      6. `make test` until green.
      7. `make lint && make vuln` clean.
      8. Commit via `/caveman-commit`.

      ## Debugging the plugin

      Set `DEBUG_PG_SAN=1` in the environment before invoking `protoc`; the
      plugin (see `main.go`) wires it into `protoc-gen-star`'s debug logger.
      Read the README's **Debug** section for the VS Code launch setup
      (current state: best-effort).

      ## Project layout

      See `AGENTS.md` section **Repository structure**.
      ```

- [ ] Step 2: Verify the file renders cleanly (`mdformat --check docs/DEVELOPMENT.md`
      if `mdformat` available, otherwise visual check).

## Verification

- [ ] `docs/DEVELOPMENT.md` exists.
- [ ] `README.md` links to it under a new `## Development` section:
      `See [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md).`

## Cross-references

- `AGENTS.md` — single source of truth for tier/type/standards.
- Plan: `2026-05-20-onboard-linting.md` — `make lint` target.
- Plan: `2026-05-20-onboard-vendoring.md` — `make vendor-check`.
- Plan: `2026-05-20-onboard-testing.md` — `make test` + coverage.
