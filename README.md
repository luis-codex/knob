# knob

System settings — audio, network — in a terminal UI.

`knob` is a keyboard-driven TUI for Linux that shows and changes what your
system is actually doing: output/input volume and devices via PulseAudio /
PipeWire, and (soon) network via NetworkManager or iwd. Built with
[Bubble Tea](https://github.com/charmbracelet/bubbletea) and
[Lip Gloss](https://github.com/charmbracelet/lipgloss).

> **Status: beta.** It works day to day, but interfaces are not frozen yet — see
> [Compatibility](#compatibility) and [Known limitations](#known-limitations).

<!-- TODO: screenshot / gif -->

## Requirements

- Linux with an interactive terminal (true-color recommended; it degrades on
  poorer terminals).
- **Audio:** a running PulseAudio or PipeWire (`pipewire-pulse`) session.
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
`rm` the binary. The config at `~/.config/knob/` is left in place — delete it
with `rm -rf ~/.config/knob`.

## Usage

```
knob              open the settings
knob -version     print version, commit and build date
knob -h           full help
```

### Keys

| Key | Action |
| --- | --- |
| `↑`/`↓` or `k`/`j` | move within a list |
| `→`/`l` or `Enter`/`Tab` | enter a section |
| `←`/`h` or `Esc` | back to the menu |
| `Ctrl-B` | hide or show the sidebar |
| `q` | quit |
| `Ctrl-C` | quit immediately (always) |

Each screen shows its own extra keys in the footer.

### Configuration

`knob` keeps every preference — audio, theme and interface — in one file,
`$XDG_CONFIG_HOME/knob/config.toml` (usually `~/.config/knob/config.toml`).
knob owns it: it is created with every key at its default on first run, and
edited from the Preferences screen in the TUI, which saves immediately.

```toml
[audio]
volume_step = 5        # how much a left/right press moves the volume (1–50)

[theme]
mode = "auto"           # auto | light | dark

[theme.dark]
accent = "#516BEB"      # every color role, per variant; unset means "use the default"
# ...

[theme.light]
# ...

[interface]
sidebar_hidden = false  # start with the sidebar collapsed
animations     = true   # the playing-stream blink, and anything added later
```

You can still hand-edit it: a missing key keeps its default, an out-of-range
or malformed value falls back to its default and is reported on stderr naming
the key, and a whole-file syntax error is reported and knob starts on the
defaults. The one thing hand-editing does not survive is comments, or a
rejected value itself — knob rewrites the file on every change made from the
TUI, so annotations and typoed values you haven't fixed yet are both lost at
the next save.

> Older versions kept the theme in a separate, read-only `theme.toml`. Delete
> it after upgrading; knob no longer reads it.

### Exit codes

| Code | Meaning |
| --- | --- |
| `0` | success |
| `1` | runtime failure |
| `2` | bad arguments |
| `3` | stdout is not an interactive terminal (pipe, CI, cron) |
| `130` | interrupted with Ctrl-C |

## Known limitations

- **Network / Wi-Fi:** the screen is a placeholder — it does not manage networks
  yet (needs a NetworkManager or iwd adapter). It is hidden from the menu for
  now.
- Only tested against PulseAudio / PipeWire on `x86_64` Arch Linux so far.

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
