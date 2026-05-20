# Onboard — SBOM (CycloneDX SBOM in CI + release)

## Goal

Generate a CycloneDX SBOM (`sbom.cdx.json`) for every CI run on `master` and
attach it to each tagged release artifact.

## Context

`AGENTS.md` section **Releases** lists `sbom.cdx.json` as a required release
artifact. Repo has no SBOM tooling configured.

## File Structure

- Modify: `Makefile` (add `sbom` target).
- Modify: CI config (add `sbom` job — handled in `*-onboard-ci-pipeline.md`).
- Add release artifact (handled in `*-onboard-release-signing.md`).

## Tasks

### Task 1: Add `sbom` Makefile target using `syft`

- [ ] Step 1: Confirm `syft` available (`syft version`). If not, install
      via `curl -sSfL https://raw.githubusercontent.com/anchore/syft/main/install.sh | sh -s -- -b /usr/local/bin`.
- [ ] Step 2: Append to `Makefile`:

      ```makefile
      .PHONY: sbom
      sbom:
      	@syft dir:. -o cyclonedx-json=sbom.cdx.json
      ```

- [ ] Step 3: Add `sbom.cdx.json` to `.gitignore` (it's a build artifact, not
      tracked).

### Task 2: CI integration

- [ ] Step 1: In the CI pipeline (`*-onboard-ci-pipeline.md`), add a `sbom`
      job that runs `make sbom` and stores `sbom.cdx.json` as an artifact.
- [ ] Step 2: For tag jobs (`v*`), include `sbom.cdx.json` in the release
      assets.

### Task 3: Verify

- [ ] Step 1: `make sbom` exits 0.
- [ ] Step 2: `jq '.bomFormat' sbom.cdx.json` returns `"CycloneDX"`.

## Verification (end-to-end)

- [ ] CI artifact contains `sbom.cdx.json` for every pipeline run.
- [ ] Each release contains `sbom.cdx.json` next to the binaries.

## Cross-references

- Plan: `2026-05-20-onboard-ci-pipeline.md` — CI job that calls `make sbom`.
- Plan: `2026-05-20-onboard-release-signing.md` — releases bundle the SBOM.
- Plan: `2026-05-20-onboard-license-compliance.md` — also uses `syft` output.
