# Development — protoc-gen-sanitize

## Prerequisites

- Go ≥ 1.22 (module floor 1.18).
- `protoc` (Protocol Buffers compiler).
  - Ubuntu: `sudo apt install protobuf-compiler`
  - macOS: `brew install protobuf`
- `golangci-lint` ≥ v2 — install via `/lint-go-install` or
  `go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest`.
- `govulncheck` — install via `/govulncheck-install` or
  `go install golang.org/x/vuln/cmd/govulncheck@latest`.
- `gitleaks` — `sudo apt install gitleaks` or
  `curl -sSfL https://raw.githubusercontent.com/gitleaks/gitleaks/master/install.sh | sh`.
- `syft` (for SBOM) —
  `curl -sSfL https://raw.githubusercontent.com/anchore/syft/main/install.sh | sh -s -- -b ~/.local/bin`.
- `go-licenses` (for license-check) —
  `go install github.com/google/go-licenses@latest`.
- `pre-commit` (optional, recommended) — `pipx install pre-commit`.

## First-time setup

```bash
git clone git@github.com:Intrinsec/protoc-gen-sanitize.git
cd protoc-gen-sanitize
pre-commit install        # gitleaks hook
make distclean
make test                 # codegen + build + test, end-to-end
```

## Daily commands

| Command | What it does |
|---|---|
| `make generate` | Regenerate `tests/*.pb.go` and `tests/*.pb.sanitize.go` |
| `make test` | `generate` → `go test` with coverage |
| `make lint` | `generate` → `golangci-lint run ./...` |
| `make vuln` | `govulncheck ./...` |
| `make vendor-check` | Verify `vendor/` matches `go.mod` |
| `make sbom` | CycloneDX SBOM → `sbom.cdx.json` |
| `make license-check` | Allowlist gate over Go modules |
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

```bash
DEBUG_PG_SAN=1 make test
```

## Releasing

Releases are tag-driven on `v*` tags. CI runs GoReleaser (`.goreleaser.yaml`)
producing:

- `protoc-gen-sanitize_<version>_<os>_<arch>.tar.gz` ×4 (linux/darwin × amd64/arm64).
- `SHA256SUMS` (and `SHA256SUMS.sig` from cosign keyless signing).
- `sbom_<os>_<arch>.cdx.json` per archive.

### Verifying a release

```bash
cosign verify-blob \
  --signature SHA256SUMS.sig \
  --certificate-identity-regexp 'https://github.com/Intrinsec/protoc-gen-sanitize/.*' \
  --certificate-oidc-issuer 'https://token.actions.githubusercontent.com' \
  SHA256SUMS

sha256sum -c SHA256SUMS
```

## Project layout

See [AGENTS.md](../AGENTS.md) section **Repository structure**.

## Standards

[AGENTS.md](../AGENTS.md) is the single source of truth for tier, type, and
mandatory tooling. iagen-dev correction plans live under
[docs/superpowers/plans/](superpowers/plans/).
