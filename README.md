# knob

System settings — audio, Bluetooth, network — in a terminal UI.

`knob` is a keyboard-driven TUI for Linux that shows and changes what your
system is actually doing: output/input volume and devices via PulseAudio /
PipeWire, Bluetooth devices via BlueZ, and (soon) network via NetworkManager or
iwd. Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) and
[Lip Gloss](https://github.com/charmbracelet/lipgloss).

> **Status: beta.** It works day to day, but interfaces are not frozen yet — see
> [Compatibility](#compatibility) and [Known limitations](#known-limitations).

<!-- TODO: screenshot / gif -->

## Requirements

- Linux with an interactive terminal (true-color recommended; it degrades on
  poorer terminals).
- **Audio:** a running PulseAudio or PipeWire (`pipewire-pulse`) session.
- **Bluetooth:** BlueZ (`bluetoothd`) reachable on the session/system D-Bus.
- No root required — it uses your session buses only.
- To build from source: Go 1.27 or newer.

## Install

```sh
curl -fsSL https://raw.githubusercontent.com/luis-codex/knob/main/install.sh | sh
```

Installs the latest release for your OS/arch after verifying its checksum.

Native packages (`.deb`, `.rpm`, Arch `.pkg.tar.zst`), tarballs, `KNOB_VERSION` /
`BINDIR` overrides, and building from source: **[INSTALL.md](INSTALL.md)**.

To remove: your package manager (`pacman -R knob`, `apt remove knob`, …), or
`rm` the binary. The theme at `~/.config/knob/` is left in place — delete it
with `rm -rf ~/.config/knob`.

## Usage

```
knob              open the settings
knob -write-theme  drop an example theme in ~/.config/knob/
knob -fake-bluetooth  use fake devices (development, no hardware needed)
knob -version     print version, commit and build date
knob -h           full help
```

### Keys

| Key | Action |
| --- | --- |
| `↑`/`↓` or `k`/`j` | move within a list |
| `→`/`l` or `Enter`/`Tab` | enter a section |
| `←`/`h` or `Esc` | back to the menu |
| `q` | quit |
| `Ctrl-C` | quit immediately (always) |

Each screen shows its own extra keys in the footer.

### Configuration

`knob` reads an optional theme from `$XDG_CONFIG_HOME/knob/theme.toml` (usually
`~/.config/knob/theme.toml`). Run `knob -write-theme` to drop a commented
template with every color and its default; uncomment only what you want to
change. An invalid theme never blocks startup — bad values are reported on
stderr and the defaults are used.

Behavioural settings live in `$XDG_CONFIG_HOME/knob/config.toml`, all optional:

```toml
[audio]
volume_step = 5      # how much a left/right press moves the volume (1–50)

[bluetooth]
scan_seconds = 8     # how long a device scan lasts (1–60)
```

Same rule as the theme: a missing file is a normal first run, a malformed file
falls back to the defaults, and an out-of-range value keeps that field's
default. Anything wrong is reported on stderr, naming the key.

### Exit codes

| Code | Meaning |
| --- | --- |
| `0` | success |
| `1` | runtime failure |
| `2` | bad arguments |
| `3` | stdout is not an interactive terminal (pipe, CI, cron) |
| `130` | interrupted with Ctrl-C |

## Known limitations

- **Bluetooth:** listing, pairing state and connect/disconnect work; accepting
  an *incoming* pairing request is not implemented yet.
- **Network / Wi-Fi:** the screen is a placeholder — it does not manage networks
  yet (needs a NetworkManager or iwd adapter). It is hidden from the menu for
  now.
- Only tested against PulseAudio / PipeWire and BlueZ on `x86_64` Arch Linux so
  far.

## Compatibility

While the version is `0.x` (and during the beta), **anything can change** —
config file format, flags, keybindings, exit codes — between minor releases.
Breaking changes are called out in [`CHANGELOG.md`](CHANGELOG.md). Semantic
versioning stability guarantees start at `1.0.0`.

## Reporting bugs

Open an issue: <https://github.com/luis-codex/knob/issues>. Please include your
distro, terminal, `knob -version`, and how to reproduce it.

## License

[GPL-3.0-or-later](LICENSE) © Luis Tenorio.
