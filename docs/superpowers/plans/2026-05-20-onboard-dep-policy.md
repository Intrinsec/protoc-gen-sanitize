# Onboard — Dependency policy (Renovate + version bumps)

## Goal

Set up Renovate so dependencies refresh automatically and bring the existing
2022-era `go.mod` deps up to current patch lines.

## Context

`go.mod` (as of 2026-05-20):

```
golang/protobuf v1.5.2
google/go-cmp v0.5.8
lyft/protoc-gen-star v0.6.0
spf13/afero v1.8.2
microcosm-cc/bluemonday v1.0.18
google.golang.org/protobuf v1.28.0
golang.org/x/text v0.3.7
golang.org/x/tools v0.1.10
... (more transitives)
```

All are 3+ years stale. Baseline `govulncheck` shows 10 uncalled CVEs in
required modules — almost certainly cleared by routine bumps.

## File Structure

- Create: `renovate.json` at repo root.
- Modify: `go.mod`, `go.sum`, `vendor/` (post-bump).

## Tasks

### Task 1: Add `renovate.json`

- [ ] Step 1: Create `/home/sml/Work/protoc-pluggins/protoc-gen-sanitize/renovate.json`:

      ```json
      {
        "$schema": "https://docs.renovatebot.com/renovate-schema.json",
        "extends": [
          "config:recommended",
          ":semanticCommits",
          ":semanticCommitTypeAll(deps)",
          ":timezone(Europe/Paris)",
          ":dependencyDashboard"
        ],
        "schedule": ["before 4am on monday"],
        "labels": ["deps"],
        "postUpdateOptions": ["gomodTidy", "gomodUpdateImportPaths"],
        "packageRules": [
          {
            "matchManagers": ["gomod"],
            "matchUpdateTypes": ["patch"],
            "automerge": true
          },
          {
            "matchManagers": ["gomod"],
            "matchUpdateTypes": ["minor"],
            "automerge": false
          },
          {
            "matchManagers": ["gomod"],
            "matchUpdateTypes": ["major"],
            "automerge": false,
            "labels": ["deps", "major"]
          }
        ]
      }
      ```

### Task 2: Initial bulk bump

Done **once** to clear the 2022 backlog. Renovate handles the rest.

- [ ] Step 1: `go get -u ./...` → exit 0.
- [ ] Step 2: `go mod tidy && go mod vendor` → exit 0.
- [ ] Step 3: `make distclean && make test` → exit 0. If failures, bisect
      which module bump broke things and pin or downgrade.
- [ ] Step 4: `make vuln` → 0 called CVEs and ideally 0 uncalled too.
- [ ] Step 5: Commit as a single `deps: bulk bump from 2022 baseline` MR
      (do **not** mix with feature work).

### Task 3: Enable Renovate

- [ ] Step 1: On GitLab — enable `renovate-bot` in project Integrations,
      or install the public bot. On GitHub — install the Renovate app.
- [ ] Step 2: First Renovate run produces a "Dependency Dashboard" issue.
      Verify schedule and patch-automerge behavior.

## Verification (end-to-end)

- [ ] `go.mod` minor versions ≤ 6 months old.
- [ ] `govulncheck` baseline drops to 0/0.
- [ ] Renovate Dashboard issue present and updating weekly.

## Cross-references

- Plan: `2026-05-20-onboard-vendoring.md` — vendoring must precede this.
- Plan: `2026-05-20-onboard-vuln-scanning.md` — CVE count tracked here.
- Plan: `2026-05-20-onboard-ci-pipeline.md` — Renovate MRs must run full CI.
