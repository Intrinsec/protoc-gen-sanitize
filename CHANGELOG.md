# Changelog

All notable user-facing changes to `protoc-gen-sanitize` are documented here.

Format follows [Keep a Changelog 1.1.0](https://keepachangelog.com/en/1.1.0/)
and [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2026-05-20

First tagged release following the iagen-dev onboarding pass.
Predecessors `v0.0.1` … `v0.0.15` were untagged-from-Changelog snapshots;
this entry summarizes the user-visible delta since `v0.0.15`.

### Added

- Source-relative output paths. Generated `.pb.sanitize.go` files now sit
  next to their input `.proto` files (mirroring `protoc-gen-go`'s
  `paths=source_relative` behavior). Multi-directory proto trees that
  previously collided on a shared basename (`foo/a.proto` and
  `bar/a.proto` both producing `a.pb.sanitize.go`) now generate
  collision-free outputs.
- Multi-package regression test fixture (`tests/sub/`).
- `main.version` and `main.commit` build-time variables, populated by
  goreleaser via `-ldflags -X`. Visible with
  `strings <binary> | grep main.version`.

### Changed

- **BREAKING:** Go module floor raised from `1.18` to `1.25.0`. Consumers
  embedding the plugin via `go install` need Go ≥ 1.25 to build it.
  Runtime use through a precompiled binary is unaffected.
- **BREAKING:** When invoking `protoc`, pass `--sanitize_out=<source-root>`
  (typically `.`) instead of `--sanitize_out=<flat-output-dir>`. The
  plugin now returns the full relative path, so a flat output directory
  causes a double-prefix. Single-package projects with proto files in one
  flat directory continue to work by passing that directory as the
  output base.
- Dependencies bumped to current minor/patch lines:
  `google.golang.org/protobuf` v1.28.0 → v1.36.11,
  `golang.org/x/tools` v0.21.1 → v0.45.0,
  `golang.org/x/text` v0.16.0 → v0.37.0,
  `github.com/spf13/afero` v1.8.2 → v1.15.0,
  `github.com/google/go-cmp` v0.6.0 → v0.7.0,
  `github.com/golang/protobuf` v1.5.2 → v1.5.4,
  `github.com/lyft/protoc-gen-star` v0.6.0 → v0.6.2,
  `github.com/microcosm-cc/bluemonday` v1.0.18 → v1.0.27.

### Security

- Vendored dependency tree refreshed. `govulncheck` reports 0 **called**
  vulnerabilities. Uncalled-CVE count in required modules dropped from 7
  to 5 thanks to the dependency bumps above.

<!-- Internal-only changes (CI bootstrap, AGENTS.md, lint fixes, docs)
     intentionally omitted per Keep-a-Changelog wording rules. -->

[Unreleased]: https://github.com/Intrinsec/protoc-gen-sanitize/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/Intrinsec/protoc-gen-sanitize/compare/v0.0.15...v0.1.0
