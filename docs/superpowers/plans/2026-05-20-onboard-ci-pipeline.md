# Onboard — CI pipeline (lint + vendor-check + test + vuln + SBOM + release)

## Goal

Stand up a CI pipeline for protoc-gen-sanitize: lint, vendor-check, test with
coverage, govulncheck, SBOM, secret scan, and a tag-triggered release stage.

## Context

The repo has no CI config (no `.gitlab-ci.yml`, no `.github/workflows/`). Tier
B Go CLI requires a full pipeline (see `AGENTS.md` section **CI / CD**).

The host (GitLab vs GitHub) is **not detected automatically**. If unsure, ask
the user before emitting `.gitlab-ci.yml`. The `isec-iagen_gitlab-cicd-go`
skill targets GitLab; for GitHub Actions, mirror the same job set in
`.github/workflows/ci.yml`.

## File Structure

- Create: `.gitlab-ci.yml` OR `.github/workflows/ci.yml` (one or the other,
  not both).
- Touch: `Makefile` (CI calls only `make <target>`, no inline tool invocations
  in the YAML).

## Tasks

### Task 1: Pick the host

- [ ] Step 1: `git remote -v` → confirm whether the upstream is GitLab or
      GitHub.
- [ ] Step 2: Record the choice in `AGENTS.md` section **CI / CD**.

### Task 2: Bootstrap the pipeline (GitLab branch)

If host = GitLab:

- [ ] Step 1: Invoke `/isec-iagen_gitlab-cicd-go` to generate `.gitlab-ci.yml`
      with the standard tier-B job set:
      - `lint` → `make lint`
      - `vendor-check` → `make vendor-check`
      - `test` → `make test`, store `tests/cover.out` as artifact
      - `vuln` → `make vuln`
      - `sbom` → CycloneDX (depends on `*-onboard-sbom.md`)
      - `secret-scan` → gitleaks (depends on `*-onboard-secret-scanning.md`)
      - `license-check` → syft (depends on `*-onboard-license-compliance.md`)
      - `release` → on tag `v*`, build matrix linux/darwin × amd64/arm64,
        sign with cosign, attach `SHA256SUMS`.
- [ ] Step 2: Pipeline green on a feature branch before merging to `master`.

### Task 3: Bootstrap the pipeline (GitHub branch)

If host = GitHub, write `.github/workflows/ci.yml` mirroring the GitLab job
set. Each job runs `make <target>` so behavior matches local. Reuse
`actions/setup-go@v5`, `golangci/golangci-lint-action@v6`, `anchore/sbom-action`,
`gitleaks/gitleaks-action`, `sigstore/cosign-installer`.

(Exact YAML deferred to follow-up — the standard skill is GitLab-first.)

### Task 4: Branch protection

- [ ] Step 1: Mark `master` as protected and require all CI jobs green
      before merge (GitLab "Merge request approvals" / GitHub "Branch
      protection rules").

## Verification (end-to-end)

- [ ] CI run on a feature branch shows all jobs green.
- [ ] Tag a `v0.0.0-onboard-test` and verify the release job produces signed
      binaries + checksums (revert/delete the tag after).
- [ ] `2026-05-20-onboard-precheck.md` Task 2 re-runs cleanly inside the CI
      container.

## Cross-references

- Skill: `isec-iagen_gitlab-cicd-go`.
- Plans: `*-onboard-linting.md`, `*-onboard-vendoring.md`,
  `*-onboard-testing.md`, `*-onboard-vuln-scanning.md`, `*-onboard-sbom.md`,
  `*-onboard-secret-scanning.md`, `*-onboard-license-compliance.md`,
  `*-onboard-release-signing.md`.
