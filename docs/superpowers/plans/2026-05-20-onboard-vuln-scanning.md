# Onboard — Vulnerability scanning (govulncheck wired to CI)

## Goal

Make `govulncheck ./...` a non-skippable CI gate, document the
called-vs-uncalled triage policy, and clear or accept the existing uncalled
CVEs in `go.mod`.

## Context

`AGENTS.md` section **Vulnerability scanning** requires zero **called**
vulnerabilities. Baseline (2026-05-20): 0 called, 3 imported + 7 required
modules uncalled. Local scan is already clean; the gap is policy + CI wiring.

## File Structure

- Create: `docs/superpowers/plans/...-onboard-vuln-scanning.md` (this file).
- Modify: `Makefile` (add `vuln` target).
- Touch (later, in `*-onboard-ci-pipeline.md`): CI job that runs `make vuln`.

## Tasks

### Task 1: Capture uncalled CVE list

- [ ] Step 1: `govulncheck -show verbose ./... > .govulncheck-baseline.txt 2>&1`.
- [ ] Step 2: Manually inspect the output. Any **called** CVE → STOP, file
      a bump-deps issue and patch before continuing. Baseline says 0 called,
      verify it still holds.
- [ ] Step 3: Add `.govulncheck-baseline.txt` to `.gitignore` (it is a
      diagnostic snapshot, not a tracked artifact).

### Task 2: Add `vuln` Makefile target

- [ ] Step 1: Append to `Makefile`:

      ```makefile
      .PHONY: vuln
      vuln:
      	@govulncheck ./...
      ```

- [ ] Step 2: `make vuln` → exit 0.

### Task 3: Document triage in AGENTS.md

- [ ] Step 1: `AGENTS.md` already states the policy ("zero called, uncalled
      tracked via Renovate"). No edit required — verify the text matches the
      Makefile target name.

## Verification (end-to-end)

- [ ] `make vuln` exits 0 locally.
- [ ] CI plan (`...-onboard-ci-pipeline.md`) includes a `vuln` job calling
      `make vuln`.
- [ ] When the dependency-policy plan lands, uncalled CVEs auto-PR on patch
      release.

## Cross-references

- Skill: `isec-iagen_govulncheck-install` if the tool is missing.
- Skill: `isec-iagen_govulncheck` for re-running and triaging results.
- Plan: `2026-05-20-onboard-dep-policy.md` for Renovate auto-bumps.
- Plan: `2026-05-20-onboard-ci-pipeline.md` for the CI hook.
