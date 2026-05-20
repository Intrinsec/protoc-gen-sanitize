# Onboard — Secret scanning (gitleaks pre-commit + CI)

## Goal

Detect committed secrets before they hit `master` via gitleaks running both
in a pre-commit hook and in CI, with an explicit allowlist file.

## Context

`AGENTS.md` section **Secret scanning** requires gitleaks in pre-commit and CI
with a baseline `.gitleaks.toml`. Repo has neither.

The repo is a protoc plugin — unlikely to hold secrets — but the gate is still
mandatory for tier B (prevents accidental commits during experiments).

## File Structure

- Create: `.gitleaks.toml` (allowlist + rule overrides).
- Create: `.pre-commit-config.yaml` (or extend if exists).
- CI integration handled in `*-onboard-ci-pipeline.md`.

## Tasks

### Task 1: Baseline scan

- [ ] Step 1: `gitleaks detect --no-banner --redact -v` → exit 0 expected.
      If non-zero, **STOP** and review findings. Genuine secrets must be
      rotated and history scrubbed before continuing.

### Task 2: Configure `.gitleaks.toml`

- [ ] Step 1: Create `/home/sml/Work/protoc-pluggins/protoc-gen-sanitize/.gitleaks.toml`:

      ```toml
      [extend]
      useDefault = true

      [allowlist]
      description = "protoc-gen-sanitize allowlist"
      paths = [
        '''sanitize/sanitize\.pb\.go''',
        '''tests/.*\.pb\.go''',
        '''tests/.*\.pb\.sanitize\.go''',
        '''vendor/.*''',
      ]
      ```

### Task 3: Pre-commit hook

- [ ] Step 1: Create `.pre-commit-config.yaml`:

      ```yaml
      repos:
        - repo: https://github.com/gitleaks/gitleaks
          rev: v8.18.4
          hooks:
            - id: gitleaks
      ```

- [ ] Step 2: `pre-commit install` (once per clone — document in
      `docs/DEVELOPMENT.md`).

### Task 4: CI integration

- [ ] Handled in `*-onboard-ci-pipeline.md` (`secret-scan` job).

## Verification (end-to-end)

- [ ] `gitleaks detect --no-banner --config .gitleaks.toml` → exit 0.
- [ ] Pre-commit hook fires when staging a test secret (e.g. add
      `AWS_SECRET_ACCESS_KEY=AKIA...` to a scratch file → commit blocked).
- [ ] CI fails on a branch that introduces a secret.

## Cross-references

- Plan: `2026-05-20-onboard-ci-pipeline.md`.
- Plan: `2026-05-20-onboard-dev-guide.md` — mention `pre-commit install`.
