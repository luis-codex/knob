# One `config.toml` store, behind a domain port

## Context

Today knob reads two hand-edited files through two loose functions in
`internal/infrastructure/config`:

- `theme.toml` → `LoadTheme()` (`theme.go:48`), returns `styles.Custom`
- `config.toml` → `LoadSettings()` (`settings.go:49`), returns `config.Settings`
  with a single field, `VolumeStep`

Neither sits behind a port, nothing can write them, and the package doc states
outright that knob "never writes theme.toml or config.toml".

We want a single `~/.config/knob/config.toml` acting as knob's store — its
"database" for now. Every top-level key of the file is a domain concern
(`[audio]`, `[theme]`, `[interface]`), the contract lives in the domain as a
repository port, and TOML is just one implementation of it. The point of the
port is that swapping TOML for SQLite later is a change to the composition
root plus one new infrastructure package, nothing else. That is the payoff DDD
is here to buy.

This supersedes two earlier decisions, deliberately:

1. `infrastructure/config/config.go` says presentation/theme stays out of the
   domain. The *stored* theme values (validated hex strings) now move into the
   domain; *rendering* does not. See "Layer boundary" below.
2. Two files become one. The project is `0.x` and `CHANGELOG.md` states the
   config surface may change between minor releases, so there is **no
   migration shim** — the change is documented instead.

## Layer boundary (the checkable invariant)

> The new domain package must not import `charm.land/lipgloss/v2` or
> `image/color`.

Domain holds hex strings validated into a `Color` value object. The conversion
`Color` → `lipgloss.Color` happens in `internal/shared/styles`. If an import of
either package appears under `internal/domain/`, the boundary has been
violated. `internal/domain/` is currently clean of both — verified.

## The file

```toml
[audio]
volume_step = 5

[theme]
mode = "auto"          # auto | light | dark

[theme.dark]
accent = "#516BEB"
# … the 13 colour roles, per variant

[theme.light]
# …

[interface]
sidebar_hidden = false
animations     = true
```

## Preferences and their existing consumers

Every knob below already controls live behaviour — nothing here is invented.

| Key | Replaces | Consumer |
| --- | --- | --- |
| `audio.volume_step` | `config.Settings.VolumeStep` | `audio.go:286,288` via `app.New` → `newRouter` |
| `theme.dark` / `theme.light` | all of `theme.toml` | `styles.Custom` → `app.go:77,101` |
| `theme.mode` | *new — closes a real gap* | `app.go:77,101` |
| `interface.sidebar_hidden` | zero value at `app.go:64` | `app.go:64`, initial state |
| `interface.animations` | `blinkInterval` at `pages/audio.go:19` | `ensureBlink`, `audio.go:147,152` |

`theme.mode` fixes something genuinely broken: light/dark is detected *only*
via `tea.RequestBackgroundColor` (`app.go:86`), and `app.go:77` hardcodes dark
as the provisional palette. A terminal that never answers the query stays dark
forever with no way to override it.

Explicitly **out of scope** (real constants, but each needs its own change):
`styles.SidebarWidth`, `audio.MaxVolume` (a domain invariant, not a
preference), `meter.BarSegments`, and any keybinding remapping — bindings are
string literals in eight places with hint strings hardcoded separately
(`app.go:220-238`, `audio.go:557-567`), so that is a distinct project.

## Design

### 1. Domain — `internal/domain/preferences`

Modelled on `internal/domain/audio` (`device.go` for value-object style,
`repository.go` for port style).

```
preferences/
  preferences.go   Settings aggregate + Default()
  audio.go         Audio; VolumeStep value object (1–50)
  theme.go         Theme; Mode value object; Palette of Color (hex only)
  interface.go     Interface: SidebarHidden, Animations
  repository.go    Repository port
  *_test.go
```

Root type is `preferences.Settings` — not `preferences.Preferences`, to avoid
the stutter:

```go
type Settings struct {
    Audio     Audio
    Theme     Theme
    Interface Interface
}
```

Validation currently stranded in infrastructure moves here: the 1–50 volume
range (`settings.go:15-17`) and the `#RGB`/`#RRGGBB` check (`theme.go:19`)
become constructors returning `errs.Invalid(...)`, exactly as
`audio.NewVolume` already does.

### 2. The port and its error semantics

The one decision that could silently regress behaviour, so it is explicit.
Both current loaders return `(value, []error)` on purpose — a bad hex colour
must never stop knob starting, and `README.md:76-78` documents that publicly.
A plain `Load(ctx) (Settings, error)` would throw it away.

```go
type Repository interface {
    // Load returns the stored preferences. A value the store holds but the
    // domain rejects is replaced by its default and reported in
    // Loaded.Rejected; the error is reserved for not being able to read the
    // store at all.
    Load(ctx context.Context) (Loaded, error)
    Save(ctx context.Context, s Settings) error
}

// Rejected is a stored value the domain refused, and why.
type Rejected struct {
    Key    string // "audio.volume_step"
    Reason string
}

type Loaded struct {
    Settings Settings
    Rejected []Rejected
}
```

Missing file → defaults, no error. Out-of-range or unparseable *value* → that
field defaults, one `Rejected` naming its key. Unreadable directory, or a file
that is not valid TOML at all → `error`.

A whole-file parse failure is an `error` and **not** a `Rejected`, for two
reasons. There is no key to name, and `README.md:76-78` promises stderr output
that names the key. More importantly it guards the file: see "Never overwrite
a file we could not parse" below.

### 2b. Never overwrite a file we could not parse

The store is writable now, which creates a data-loss path that does not exist
today. A syntax error is currently harmless — knob ignores the file and the
user fixes their typo. Under a writable store: bad TOML → defaults loaded →
the user saves anything → `toml.Marshal` writes the defaults over their file
and their config is gone.

So `tomlstore` records that its last `Load` failed to parse and **refuses to
`Save`** until a successful `Load`, returning an error the UI surfaces. A
`parseFailed bool` on the store struct is enough. `main` reports the parse
error on stderr and carries on with `preferences.Default()` — today's
behaviour exactly, minus the risk.

### 3. Infrastructure — `internal/infrastructure/tomlstore`

Implements `preferences.Repository` against `~/.config/knob/config.toml`.
Named for the mechanism, not the concern, so the swap story is literal: a
later `sqlitestore` sits beside it and the composition root picks one.

- Owns the XDG-aware path helper moved from `config.go:26-34`.
- `Load`: read → `toml.Unmarshal` into raw structs → domain constructors →
  collect `Rejected`. Absent keys keep their zero value, read as "use the
  default" — the rule `settings.go:35-36` already uses.
- `Load` **materialises the file with defaults when it is absent** — the
  store's "create schema if missing" step, which is what makes the file
  self-documenting now that there is no commented template. Failing to write
  it is non-fatal: knob must still start on a read-only home directory.
- `Save`: `toml.Marshal` → atomic write (temp file + rename, `0o600`,
  `MkdirAll` on the parent).

`internal/infrastructure/config` is deleted — `settings.go`, `theme.go`,
`config.go` and their tests are replaced by the above plus §5.

**Comments are lost on write — accepted.** Verified empirically:
`toml.Marshal` of a parsed document drops comments. An app-written file cannot
carry a commented template; the fully-populated default file replaces it.

### 4. Application — `internal/application/prefs`

Named to avoid clashing with the domain package — the same trick
`application/sound` uses against `domain/audio` (`sound.go:1-6`). Two use
cases, one `Execute` each, sharing `Deps`, matching the shape
`application/sound` establishes:

```
prefs/
  prefs.go      Deps{ Repo preferences.Repository }
  load.go       Load
  save.go       Save
  usecases.go   UseCases + NewUseCases
```

Pages call these. **No page touches `preferences.Repository` directly** — that
would break the dependency direction documented in `CONTRIBUTING.md`.

### 5. Presentation mapping

`internal/shared/styles` grows a constructor turning domain colours into a
`styles.Custom`, replacing `toPalette` (`theme.go:75-118`). `shared/styles`
importing `domain/preferences` is an outer layer importing an inner one — the
allowed direction.

Resolving `preferences.Mode` against the detected background stays in
`app.Update`, which already holds the `tea.BackgroundColorMsg`
(`app.go:100-102`): `auto` keeps `msg.IsDark()`, `light`/`dark` ignore it.
`app` then calls the existing `styles.NewWithPalette`. Keeping it there is
deliberate — `shared/styles` must not pick up a bubbletea dependency it has
no other need for.

### 6. Settings page

New sidebar entry, new page, following `pages/audio.go` + `internal/ui/audio/`
as the pattern:

- `internal/pages/preferences.go` — `pages.NewPreferences(ctx, title, prefs.UseCases)`
- new nav group/item in `navGroups()` (`router.go:23-32`), mounted in
  `newRouter` (`router.go:35-41`)
- four editable rows: `volume_step` (1–50), `theme.mode` (auto/light/dark),
  `animations` (on/off), `sidebar_hidden` ("start with the sidebar hidden").
  The `[theme.light]`/`[theme.dark]` palettes stay file-only — 26 colour
  fields are not a TUI-editable surface.
- edits go through `prefs.Save` as a `tea.Cmd`, like every audio operation

A successful save emits one message carrying the new `preferences.Settings`.
`app.Update` handles it and `app.broadcast` (`app.go:110-119`) already fans
messages out to every page, so the audio page picks up a changed `volume_step`
by implementing `msgConsumer` — no new plumbing. This matters: pages are built
once in `newRouter` (`router.go:35`) and `volumeStep` is injected at
construction (`router.go:36` → `audio.go:71`), so without this the audio page
would hold a stale step for the rest of the session and the change would only
appear after a restart. The same message re-resolves the palette for
`theme.mode`.

**Ctrl-B does not write.** `sidebar_hidden` is the *startup* state, edited on
this page like any other row. Ctrl-B (`app.go:132-137`) stays a transient
session toggle — persisting on every keypress would rewrite the whole file,
comments and all, on a routine view toggle.

### 7. Composition root

`cmd/knob/main.go` replaces the two loader calls (`main.go:69,76`) with one
`tomlstore` wired into `prefs.NewUseCases`, loads once, prints each `Rejected`
to stderr naming its key (preserving documented behaviour), and passes the
whole `preferences.Settings` to `app.New` — replacing the current
`volumeStep int` parameter threaded through `app.New` → `newRouter`
(`app.go:74,79`, `router.go:36`).

### 8. CLI surface

`-write-theme` (`main.go:46`), `writeExampleTheme` (`main.go:131-143`),
`exampleTheme` (`theme.go:122-159`), the usage line (`main.go:119`) and the
`make theme` target (`Makefile:48-51`) are all removed.

## Files touched

| File | Change |
| --- | --- |
| `internal/domain/preferences/*` | new — aggregate, value objects, port |
| `internal/infrastructure/tomlstore/*` | new — TOML adapter |
| `internal/application/prefs/*` | new — Load / Save use cases |
| `internal/pages/preferences.go` | new — settings page |
| `internal/infrastructure/config/*` | deleted |
| `internal/shared/styles/theme.go` | domain → `Custom` mapping; mode → `isDark` |
| `cmd/knob/main.go` | wire the store; drop `-write-theme` |
| `internal/app/app.go`, `internal/app/router.go` | consume prefs; resolve theme mode; nav entry |
| `internal/pages/audio.go` | gate the blink on `interface.animations` |
| `Makefile` | drop the `theme` target |
| `README.md`, `INSTALL.md`, `CHANGELOG.md` | single file; `-write-theme` gone |

## Tests

Per `CLAUDE.md`, by layer:

- `domain/preferences` — pure, external test package
  (`package preferences_test`), constructors and invariants only.
- `application/prefs` — a fake `preferences.Repository` defined next to the
  test; no disk.
- `infrastructure/tomlstore` — real temp files via `t.TempDir()` +
  `t.Setenv("XDG_CONFIG_HOME", …)`, reusing the helper shape at
  `settings_test.go:11-24`. Must cover `Save` → `Load` round-trip and
  first-run file creation.

The existing cases in `settings_test.go` and `theme_test.go` — missing file,
partial override, bad value reported, malformed file, XDG path — carry over to
the new packages rather than being dropped. `app_test.go:30-31` asserts the
sidebar starts visible and must be updated to reflect the stored preference.

## Verification

```sh
make check     # gofmt + go vet + go test -race
make lint
grep -rn "lipgloss\|image/color" internal/domain/   # must print nothing
```

Then manually, in a real terminal (`make dev`) — **this part I cannot drive
myself; a TUI needs an interactive TTY, so these steps are for you to run and
I will not claim the feature works until you confirm**:

1. Back up and remove `~/.config/knob/config.toml`, run `knob` — it must start
   and create the file with every key at its default.
2. Change `volume_step` in the Settings page, then go straight to the audio
   page **without restarting** — the arrow keys must move the volume by the
   new step. This is the check that catches the stale-copy bug; persistence
   across a restart would pass either way.
3. Switch `theme.mode` — the UI must recolour immediately, and `light`/`dark`
   must hold even in a terminal whose background says otherwise.
4. Toggle `sidebar_hidden` on, quit, re-open — the sidebar must start hidden.
   Then press Ctrl-B twice and confirm `config.toml`'s mtime is **unchanged**:
   a view toggle must not write.
5. Hand-edit `config.toml` with an out-of-range `volume_step` and a bad hex
   colour — knob must still start, print both keys on stderr, and default
   exactly those two fields.
6. Corrupt `config.toml` into invalid TOML, start knob, change a setting —
   knob must start on defaults, report the parse error, and **refuse the
   save**. Confirm the corrupt file is still on disk untouched; this is the
   data-loss guard from §2b.
7. `chmod -w ~/.config/knob` and start with no config file — knob must still
   start on defaults rather than failing.
