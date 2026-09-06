# Changelog

All notable changes to this project are documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).
While the version is `0.x` the public surface (config, flags, keys, exit codes)
may change between minor releases.

## [Unreleased]

### Added
- `--version` flag: prints version, commit and build date (injected at build
  time via `-ldflags`).
- Descriptive `--help` output with a summary, examples and the issues link.
- Friendly message and dedicated exit code (`3`) when there is no interactive
  terminal.
- Packaging and release tooling: `.goreleaser.yaml` (tar.gz, checksums, `.deb` /
  `.rpm` / Arch packages), GitHub Actions for CI and releases.
- Project docs: `README.md`, `CONTRIBUTING.md`, `SECURITY.md`, `LICENSE`
  (GPL-3.0-or-later).

### Changed
- `Makefile`: build now injects version metadata and uses `-trimpath` /
  `CGO_ENABLED=0`; added `test`, `vet`, `lint`, `check` and `snapshot` targets.

[Unreleased]: https://github.com/luis-codex/settings-cli/commits/main
