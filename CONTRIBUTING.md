# Contributing

Thanks for your interest in knob. This is a beta project — bug reports,
small fixes and feedback on the UX are all welcome.

## Getting started

```sh
git clone https://github.com/luis-codex/knob
cd knob
make dev                                  # run the TUI from source
```

## Before opening a PR

```sh
make check         # gofmt + go vet + go test -race
make lint          # golangci-lint (install it first)
```

CI runs the same checks. Keep `gofmt` clean and add tests for behavior changes.

## Conventions

- **Language:** code, comments, docs and commit messages are in English.
- **Commits:** [Conventional Commits](https://www.conventionalcommits.org)
  (`feat:`, `fix:`, `docs:`, `refactor:`, `test:`, `chore:`). The changelog is
  grouped from these.
- **Comments:** short and precise, godoc style. Explain *why*, not *what*; the
  long rationale belongs in the PR description, not the code.

## Architecture

Hexagonal / ports-and-adapters, one direction of dependency:

```
internal/
  domain/          entities and invariants; ports (interfaces). No I/O.
  application/     use cases; orchestrate the domain through the ports.
  infrastructure/  everything that does I/O: port adapters (pulse, tomlstore).
  app/, pages/, shared/, ui/   the Bubble Tea model, screens and widgets.
cmd/knob/      composition root: wires concrete adapters to use cases.
```

A use case is one `Execute`-only type per action, with its own `Command` and
`Response`, grouped per aggregate under `application/<aggregate>/` and wired
into that aggregate's `UseCases` bundle. Shared plumbing (`apply`,
`requireAdapter`, …) stays unexported on `Deps`.

The `cmd/knob/main.go` composition root is the only place that names
concrete implementations. Swapping a backend is changing those lines.

## License

By contributing you agree that your contributions are licensed under
[GPL-3.0-or-later](LICENSE).
