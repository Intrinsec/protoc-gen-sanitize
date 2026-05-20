# Onboard — License compliance (SPDX allowlist + CI gate)

## Goal

Block disallowed licenses from entering the dependency tree by gating CI on
a `syft`-generated SPDX inventory against an explicit allowlist.

## Context

`AGENTS.md` section **License compliance** lists allowed SPDX IDs:
`Apache-2.0`, `MIT`, `BSD-2-Clause`, `BSD-3-Clause`, `ISC`, `MPL-2.0`. No CI
check exists today.

## File Structure

- Create: `.license-allowlist.txt`.
- Modify: `Makefile` (add `license-check` target).
- CI integration in `*-onboard-ci-pipeline.md`.

## Tasks

### Task 1: Inventory current licenses

- [ ] Step 1: `syft dir:. -o spdx-json | jq '[.packages[].licenseDeclared] | unique'`
      → list of distinct SPDX IDs currently in the tree.
- [ ] Step 2: Compare to allowlist. Any unknown ID → research, then either
      add to allowlist (with rationale in a comment) or remove the dep.

### Task 2: Create allowlist

- [ ] Step 1: Write
      `/home/sml/Work/protoc-pluggins/protoc-gen-sanitize/.license-allowlist.txt`:

      ```
      Apache-2.0
      BSD-2-Clause
      BSD-3-Clause
      ISC
      MIT
      MPL-2.0
      ```

### Task 3: Makefile target

- [ ] Step 1: Append to `Makefile`:

      ```makefile
      .PHONY: license-check
      license-check:
      	@syft dir:. -o spdx-json \
      	  | jq -r '.packages[] | "\(.name)\t\(.licenseDeclared)"' \
      	  | awk -F'\t' 'NR==FNR{ok[$$1]=1;next} {if(!ok[$$2]){print "DISALLOWED:",$$0;bad=1}} END{exit bad?1:0}' \
      	    .license-allowlist.txt -
      ```

- [ ] Step 2: `make license-check` → exit 0. If exit 1, every disallowed line
      printed must be triaged (add to allowlist with rationale or drop the
      dep).

### Task 4: CI gate

- [ ] Handled in `*-onboard-ci-pipeline.md` (`license-check` job).

## Verification (end-to-end)

- [ ] `make license-check` exits 0.
- [ ] CI fails when introducing a dep with a disallowed license (test by
      temporarily adding a GPL-only module on a throwaway branch).

## Cross-references

- Plan: `2026-05-20-onboard-sbom.md` — uses `syft` for SBOM; this plan uses
  `go-licenses` instead because syft's Go-module license detection is
  unreliable (missed `LICENSE.md`-only modules like bluemonday).
- Plan: `2026-05-20-onboard-ci-pipeline.md`.
- Plan: `2026-05-20-onboard-dep-policy.md` — Renovate must respect the
  allowlist (use `packageRules.matchFiles` to deny on disallowed license).

## Run history

```
date: 2026-05-20
runner: claude (executing-plans)
tool: switched syft+jq → go-licenses (better Go-module support)
build tag: codegenruntime (used by codegenruntime.go to anchor bluemonday)
result: clean — all detected licenses in allowlist
        (Apache-2.0, BSD-3-Clause, MIT)
```
