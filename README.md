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

### Prebuilt binary (recommended)

Download the archive for your architecture from the
[latest release](https://github.com/luis-codex/knob/releases/latest),
then:

```sh
tar -xzf knob_*_linux_amd64.tar.gz
install -Dm755 knob ~/.local/bin/knob   # or /usr/local/bin with sudo
```

Native packages (`.deb`, `.rpm`, Arch `.pkg.tar.zst`) are attached to each
release as well.

### From source

```sh
git clone https://github.com/luis-codex/knob
cd knob
make install        # builds with version info and installs to $GOBIN or $GOPATH/bin
```

Make sure that directory is on your `PATH`.

## Uninstall

```sh
make uninstall            # removes the binary from $GOBIN / $GOPATH/bin
rm -rf ~/.config/knob     # optional: also remove your theme/config
```

If you installed a prebuilt binary manually, just delete it
(`rm ~/.local/bin/knob`). Native packages uninstall with your package manager
(`apt remove knob`, `pacman -R knob`, …).

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
