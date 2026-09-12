# Installing knob

knob is Linux-only (`x86_64` / `aarch64`) — it talks to PulseAudio / PipeWire.

## Quick install (script)

```sh
curl -fsSL https://raw.githubusercontent.com/luis-codex/knob/main/install.sh | sh
```

Downloads the latest release binary for your OS/arch, verifies its SHA-256
checksum, and installs it to `/usr/local/bin` (as root) or `~/.local/bin`.

Overrides:

```sh
# pin a version
curl -fsSL https://raw.githubusercontent.com/luis-codex/knob/main/install.sh | KNOB_VERSION=v0.1.0-beta.1 sh
# choose the directory
curl -fsSL https://raw.githubusercontent.com/luis-codex/knob/main/install.sh | BINDIR=$HOME/bin sh
# system-wide without being root
curl -fsSL https://raw.githubusercontent.com/luis-codex/knob/main/install.sh | sudo sh
```

## Distro packages

Every [release](https://github.com/luis-codex/knob/releases) ships native
packages. They register with your package manager and uninstall cleanly.
Replace `amd64` with `arm64` on ARM, and the version to taste.

### Arch, Manjaro, EndeavourOS, CachyOS…

```sh
curl -LO https://github.com/luis-codex/knob/releases/download/v0.1.0-beta.1/knob_0.1.0-beta.1_linux_amd64.pkg.tar.zst
sudo pacman -U knob_0.1.0-beta.1_linux_amd64.pkg.tar.zst
# uninstall: sudo pacman -R knob
```

### Debian, Ubuntu, Mint, Pop!_OS…

```sh
curl -LO https://github.com/luis-codex/knob/releases/download/v0.1.0-beta.1/knob_0.1.0-beta.1_linux_amd64.deb
sudo apt install ./knob_0.1.0-beta.1_linux_amd64.deb
# uninstall: sudo apt remove knob
```

### Fedora, RHEL, openSUSE…

```sh
curl -LO https://github.com/luis-codex/knob/releases/download/v0.1.0-beta.1/knob_0.1.0-beta.1_linux_amd64.rpm
sudo dnf install ./knob_0.1.0-beta.1_linux_amd64.rpm
# uninstall: sudo dnf remove knob
```

## Tarball (any Linux, no root)

```sh
curl -L https://github.com/luis-codex/knob/releases/download/v0.1.0-beta.1/knob_0.1.0-beta.1_linux_amd64.tar.gz | tar xz
install -Dm755 knob ~/.local/bin/knob      # make sure ~/.local/bin is on PATH
# uninstall: rm ~/.local/bin/knob
```

## Verify checksums

```sh
curl -LO https://github.com/luis-codex/knob/releases/download/v0.1.0-beta.1/checksums.txt
sha256sum -c checksums.txt --ignore-missing
```

## From source

Needs Go 1.27 or newer.

```sh
git clone https://github.com/luis-codex/knob
cd knob
make install        # builds with version info, installs to $GOBIN / $GOPATH/bin
# uninstall: make uninstall
```

## Config

knob keeps every preference in `~/.config/knob/config.toml`, created on first
run and edited from the Preferences screen. See the README's
[Configuration](README.md#configuration) section. Removing knob leaves it in
place; delete it with `rm -rf ~/.config/knob`.
