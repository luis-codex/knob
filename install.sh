#!/bin/sh
# knob installer — https://github.com/luis-codex/knob
#
#   curl -fsSL https://raw.githubusercontent.com/luis-codex/knob/main/install.sh | sh
#
# Installs the latest release binary (pre-releases included) for your OS/arch,
# after verifying its SHA-256 checksum.
#
# Environment:
#   KNOB_VERSION   tag to install (e.g. v0.1.0-beta.1). Default: latest.
#   BINDIR         install directory. Default: /usr/local/bin as root,
#                  otherwise ~/.local/bin.
#
# For distro packages (.deb / .rpm / .pkg.tar.zst) and other options, see
# https://github.com/luis-codex/knob/blob/main/INSTALL.md

set -eu

REPO="luis-codex/knob"
BIN="knob"

say() { printf 'knob-install: %s\n' "$*"; }
err() { printf 'knob-install: %s\n' "$*" >&2; exit 1; }
have() { command -v "$1" >/dev/null 2>&1; }

if have curl; then
	fetch() { curl -fsSL "$1"; }
	fetch_to() { curl -fsSL -o "$2" "$1"; }
elif have wget; then
	fetch() { wget -qO- "$1"; }
	fetch_to() { wget -qO "$2" "$1"; }
else
	err "need curl or wget"
fi

os=$(uname -s | tr '[:upper:]' '[:lower:]')
if [ "$os" != "linux" ]; then
	err "no $os build — knob is Linux-only (it talks to PulseAudio/PipeWire). Build from source: https://github.com/$REPO#from-source"
fi

arch=$(uname -m)
case "$arch" in
	x86_64 | amd64) arch=amd64 ;;
	aarch64 | arm64) arch=arm64 ;;
	*) err "unsupported architecture: $arch" ;;
esac

ver="${KNOB_VERSION:-}"
if [ -z "$ver" ]; then
	ver=$(fetch "https://api.github.com/repos/$REPO/releases" 2>/dev/null \
		| grep -m1 '"tag_name":' | cut -d'"' -f4) || true
	[ -n "$ver" ] || err "could not resolve the latest version. Set KNOB_VERSION=vX.Y.Z and retry."
fi

if [ -n "${BINDIR:-}" ]; then
	bindir="$BINDIR"
elif [ "$(id -u)" = "0" ]; then
	bindir="/usr/local/bin"
else
	bindir="$HOME/.local/bin"
fi

asset="${BIN}_${ver#v}_${os}_${arch}.tar.gz"
base="https://github.com/$REPO/releases/download/$ver"

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT INT TERM

say "$ver ($os/$arch) -> $bindir"

fetch_to "$base/$asset" "$tmp/$asset" || err "download failed: $base/$asset"

if fetch_to "$base/checksums.txt" "$tmp/checksums.txt" 2>/dev/null; then
	if have sha256sum; then
		sum=$(sha256sum "$tmp/$asset" | cut -d' ' -f1)
	elif have shasum; then
		sum=$(shasum -a 256 "$tmp/$asset" | cut -d' ' -f1)
	else
		sum=""
		say "warning: no sha256 tool found, skipping checksum verification"
	fi
	if [ -n "$sum" ]; then
		grep -Fq -- "$sum  $asset" "$tmp/checksums.txt" || err "checksum mismatch for $asset"
	fi
else
	say "warning: could not fetch checksums.txt, skipping verification"
fi

tar -xzf "$tmp/$asset" -C "$tmp" "$BIN" || err "could not extract $BIN from the archive"

mkdir -p "$bindir"
if ! { install -m 0755 "$tmp/$BIN" "$bindir/$BIN" 2>/dev/null \
	|| { cp "$tmp/$BIN" "$bindir/$BIN" && chmod 0755 "$bindir/$BIN"; }; }; then
	err "could not write $bindir/$BIN — retry as root (curl ... | sudo sh) or set BINDIR=~/bin"
fi

say "installed $("$bindir/$BIN" -version 2>/dev/null || echo "$ver")"
case ":$PATH:" in
	*":$bindir:"*) ;;
	*) say "note: $bindir is not on your PATH — add it to use 'knob' directly" ;;
esac
