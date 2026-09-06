# Security Policy

## Supported versions

settings-cli is in beta. Only the latest release (and `main`) receives fixes.

## Reporting a vulnerability

Please **do not** open a public issue for security problems.

Use GitHub's private reporting instead:
<https://github.com/luis-codex/settings-cli/security/advisories/new>, or email
the maintainer at luisprograming0@gmail.com with a description and, if possible,
a reproduction.

Expect an acknowledgement within a few days. Once a fix is ready it will be
released and the report credited unless you prefer otherwise.

## Scope notes

`settings` runs with the invoking user's privileges and talks to the session
D-Bus (BlueZ) and the PulseAudio/PipeWire socket. It needs no elevated
privileges; a report that depends on running it as root is out of scope.
