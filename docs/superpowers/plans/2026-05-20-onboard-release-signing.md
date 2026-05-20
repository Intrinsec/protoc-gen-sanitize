# Onboard — Release signing (cosign + SHA256SUMS + SLSA provenance)

## Goal

Produce signed, checksummed release binaries for `protoc-gen-sanitize` on
every `v*` tag, with optional SLSA provenance attestation.

## Context

`AGENTS.md` section **Releases** requires:
- Binaries for linux/darwin × amd64/arm64.
- `SHA256SUMS` file.
- cosign signature (or minisig) per artifact.
- `sbom.cdx.json` (from `*-onboard-sbom.md`).

This plan handles signing + checksums + the build matrix; the CI scaffold
itself is in `*-onboard-ci-pipeline.md`.

## File Structure

- Modify: CI release job.
- Possibly add: `.goreleaser.yaml` (recommended — GoReleaser handles matrix +
  checksums + signing in one config).

## Tasks

### Task 1: Decide release tool

- [ ] Step 1: Default to **GoReleaser**. It produces matrix builds, checksums,
      signs with cosign, generates the SBOM (or wraps `syft`), and writes
      release notes. Single source of truth for both GitLab and GitHub.

### Task 2: Add `.goreleaser.yaml`

- [ ] Step 1: Create `/home/sml/Work/protoc-pluggins/protoc-gen-sanitize/.goreleaser.yaml`:

      ```yaml
      version: 2

      before:
        hooks:
          - make generate
          - go mod tidy

      builds:
        - id: protoc-gen-sanitize
          main: ./
          binary: protoc-gen-sanitize
          env:
            - CGO_ENABLED=0
          flags:
            - -mod=vendor
          ldflags:
            - -s -w -X main.version={{.Version}} -X main.commit={{.Commit}}
          goos: [linux, darwin]
          goarch: [amd64, arm64]

      archives:
        - format: tar.gz
          name_template: "{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}"

      checksum:
        name_template: "SHA256SUMS"
        algorithm: sha256

      signs:
        - cmd: cosign
          args:
            - sign-blob
            - --yes
            - --output-signature
            - "${signature}"
            - "${artifact}"
          artifacts: checksum
          signature: "${artifact}.sig"

      sboms:
        - artifacts: archive
          documents:
            - "sbom_{{ .Os }}_{{ .Arch }}.cdx.json"

      release:
        draft: false
        prerelease: auto
      ```

### Task 3: CI release job

- [ ] Step 1: In CI (`*-onboard-ci-pipeline.md`), add a `release` job that
      runs only on tags matching `v*`:
      - install `goreleaser`, `cosign`, `syft`
      - export `COSIGN_EXPERIMENTAL=1` (for keyless OIDC) or load
        `COSIGN_PRIVATE_KEY` from CI secret
      - run `goreleaser release --clean`
- [ ] Step 2: Verify the published release contains:
      `*.tar.gz` ×4, `SHA256SUMS`, `SHA256SUMS.sig`, `sbom_*.cdx.json`.

### Task 4: Document verification

- [ ] Step 1: Add to `docs/DEVELOPMENT.md` (or a new `docs/RELEASE.md`):

      ```markdown
      ## Verifying a release

      ```bash
      cosign verify-blob \
        --signature SHA256SUMS.sig \
        --certificate-identity-regexp '^https://gitlab.example/...' \
        --certificate-oidc-issuer-regexp '.*' \
        SHA256SUMS

      sha256sum -c SHA256SUMS
      ```
      ```

## Verification (end-to-end)

- [ ] Tag a `v0.0.0-onboard-test` on a feature branch (use
      `git tag -d` after).
- [ ] CI release job succeeds.
- [ ] Downloaded archives verify against `SHA256SUMS`.
- [ ] `cosign verify-blob` succeeds.

## Cross-references

- Plan: `2026-05-20-onboard-ci-pipeline.md`.
- Plan: `2026-05-20-onboard-sbom.md`.
- Plan: `2026-05-20-onboard-vendoring.md` (release builds use vendor).
