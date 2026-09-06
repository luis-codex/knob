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
- `install.sh` one-liner installer (`curl -fsSL …/install.sh | sh`) — detects
  OS/arch, verifies the checksum, installs the latest release.
- `INSTALL.md` with distro-package, tarball and from-source instructions; the
  README's install section trimmed to point at it.

### Changed
- Renamed the project and binary to **`knob`** (module `knob`, command `knob`,
  config directory `~/.config/knob/`).
- All code comments, doc strings and user-facing text translated to English.
- `Makefile`: build now injects version metadata and uses `-trimpath` /
  `CGO_ENABLED=0`; added `test`, `vet`, `lint`, `check` and `snapshot` targets.

### Removed
- Dead helpers in `internal/pages` (unused text/column helpers) flagged by
  `golangci-lint`.

[Unreleased]: https://github.com/luis-codex/knob/commits/main
