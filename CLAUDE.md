# CLAUDE.md

Guidance for Claude Code (and any agent) working in this repository. See
`CONTRIBUTING.md` for setup, commands and the architecture overview.

## Testing

Tests live next to the code they test, as sibling `_test.go` files — there is
no central `tests/` directory. What a test is allowed to touch depends on the
layer it sits in:

- **`domain/`**: pure unit tests. No mocks, no I/O. Build entities/value
  objects through their constructors and assert invariants. Use the external
  test package (`package audio_test`) so the test only sees the public API —
  see `domain/audio/device_test.go`.
- **`application/<aggregate>/`**: unit tests for use cases, driven through
  fake implementations of the domain ports. No real adapters, no real I/O.
  Define the fake next to the test, not in the production package.
- **`infrastructure/`**: tests may touch the real detail being adapted (temp
  files, env vars, ...) — that's the layer's job. Use the internal test
  package (`package config`, no `_test` suffix) when the test needs
  unexported helpers — see `infrastructure/config/settings_test.go`.
- **`app/`, `pages/`, `ui/`, `shared/`**: tests for the Bubble Tea model,
  screen navigation and widget logic.

Never write a test for one layer that reaches into another — e.g. an
application-layer test that opens a real pulse/D-Bus connection, or a domain
test that reads a file.
