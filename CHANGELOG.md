# Changelog

All notable changes to this project are documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).
While the version is `0.x` the public surface (config, flags, keys, exit codes)
may change between minor releases.

## [Unreleased]

### Added
- Preferences screen in the sidebar to edit and persist every stored setting
  from the TUI: volume step, theme mode, animations, and whether the sidebar
  starts hidden. Every change saves immediately.
- `theme.mode` (`auto` | `light` | `dark`): overrides the terminal's own
  background detection, which previously had no way to be overridden and
  stayed dark forever if the terminal never answered the query.
- `Ctrl-B` hides or shows the sidebar, giving the body the full width. The
  footer advertises it and any move back to the menu restores it.
- Behavioural settings in `$XDG_CONFIG_HOME/knob/config.toml`: `[audio]
  volume_step`, optional and range-checked. A missing file or a bad value
  never blocks startup; the file is read at the same point, and with the same
  tolerate-and-report rule, as the theme.
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
- Theme and behavioural settings are now one file, `~/.config/knob/config.toml`,
  replacing the separate `theme.toml` + `config.toml`. knob owns the file now:
  it is created with every key at its default on first run, and rewritten
  whenever a preference is saved from the TUI. Hand-editing still works —
  the same tolerate-and-report rule applies — but comments do not survive a
  save made from the app.
- The store sits behind a domain repository port
  (`internal/domain/preferences`), implemented by a TOML adapter
  (`internal/infrastructure/tomlstore`) that replaces
  `internal/infrastructure/config`. A future storage backend is a new adapter
  plus a composition-root change, not a rewrite.
- Renamed the project and binary to **`knob`** (module `knob`, command `knob`,
  config directory `~/.config/knob/`).
- All code comments, doc strings and user-facing text translated to English.
- `Makefile`: build now injects version metadata and uses `-trimpath` /
  `CGO_ENABLED=0`; added `test`, `vet`, `lint`, `check` and `snapshot` targets.

### Removed
- `-write-theme` flag and the `make theme` target: there is no commented
  template anymore now that the file is app-written; every key already
  appears in it at its default after the first run.
- Bluetooth support: domain, application use cases, BlueZ/memory/simulated
  infrastructure, UI and screen. Dropped entirely, no replacement.
- Dead helpers in `internal/pages` (unused text/column helpers) flagged by
  `golangci-lint`.

[Unreleased]: https://github.com/luis-codex/knob/commits/main
