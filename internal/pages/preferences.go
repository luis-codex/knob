package pages

import (
	"context"
	"fmt"

	tea "charm.land/bubbletea/v2"

	"knob/internal/application/prefs"
	"knob/internal/domain/preferences"
	"knob/internal/shared/components"
	"knob/internal/shared/icons"
	"knob/internal/shared/styles"
	"knob/internal/shared/ui"
)

// PreferencesSavedMsg reports a successful save, carrying the settings that
// are now current. app.Model applies it to the theme; every page that
// implements msgConsumer sees it too, since app broadcasts every message it
// does not own to all of them -- that is how the audio page picks up a
// changed volume step without a restart.
type PreferencesSavedMsg struct{ Settings preferences.Settings }

// preferencesFailedMsg reports a save the store refused, most commonly
// because config.toml currently has a syntax error.
type preferencesFailedMsg struct{ err error }

// row is one editable line, in on-screen order.
type row int

const (
	rowVolumeStep row = iota
	rowThemeMode
	rowAnimations
	rowSidebarHidden
)

var rows = []row{rowVolumeStep, rowThemeMode, rowAnimations, rowSidebarHidden}

// themeModes is the cycle order left/right steps through.
var themeModes = []string{preferences.ModeAuto, preferences.ModeLight, preferences.ModeDark}

// Preferences edits the settings stored in config.toml. Every change saves
// immediately, the same way an audio control applies as soon as it is
// pressed -- there is no separate confirm step.
type Preferences struct {
	title string
	uc    prefs.UseCases
	ctx   context.Context

	settings preferences.Settings
	list     *components.List
	failure  string
}

func NewPreferences(ctx context.Context, title string, uc prefs.UseCases, initial preferences.Settings) *Preferences {
	return &Preferences{
		title:    title,
		uc:       uc,
		ctx:      ctx,
		settings: initial,
		list:     components.NewList(),
	}
}

func (p *Preferences) HandleKey(msg tea.KeyPressMsg) (bool, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		p.list.Prev()
	case "down", "j":
		p.list.Next()
	case "left", "h":
		return true, p.adjust(-1)
	case "right", "l":
		return true, p.adjust(+1)
	default:
		return false, nil
	}
	return true, nil
}

func (p *Preferences) HandleMsg(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case PreferencesSavedMsg:
		p.settings = msg.Settings
		p.failure = ""
	case preferencesFailedMsg:
		p.failure = msg.err.Error()
	}
	return nil
}

// adjust changes the value under the cursor and saves the result. Booleans
// flip regardless of direction; only the step and the mode cycle use it.
func (p *Preferences) adjust(dir int) tea.Cmd {
	next := p.settings

	switch rows[p.list.Cursor()] {
	case rowVolumeStep:
		step, err := preferences.NewVolumeStep(next.Audio.VolumeStep.Value() + dir)
		if err != nil {
			return nil // already at the bound
		}
		next.Audio.VolumeStep = step

	case rowThemeMode:
		i := indexOf(themeModes, next.Theme.Mode.String())
		i = clamp(i+dir, 0, len(themeModes)-1)
		mode, err := preferences.NewMode(themeModes[i])
		if err != nil {
			return nil // unreachable: themeModes only holds valid modes
		}
		next.Theme.Mode = mode

	case rowAnimations:
		next.Interface.Animations = !next.Interface.Animations

	case rowSidebarHidden:
		next.Interface.SidebarHidden = !next.Interface.SidebarHidden
	}

	// Applied immediately so a second keypress before the save round-trip
	// completes reads the already-adjusted value, not a stale copy.
	p.settings = next

	return p.save(next)
}

func (p *Preferences) save(next preferences.Settings) tea.Cmd {
	uc, ctx := p.uc, p.ctx
	return func() tea.Msg {
		if _, err := uc.Save.Execute(ctx, prefs.SaveCommand{Settings: next}); err != nil {
			return preferencesFailedMsg{err: err}
		}
		return PreferencesSavedMsg{Settings: next}
	}
}

func (p *Preferences) View(t styles.Theme, width, height int) string {
	lines := make([]string, len(rows))
	for i, r := range rows {
		lines[i] = p.rowText(r)
	}

	contentHeight := max(height-frameChrome, 1)
	var tail []string
	if p.failure != "" {
		tail = []string{"", t.Body.Muted.Render(p.failure)}
		contentHeight = max(contentHeight-len(tail), 1)
	}

	body := p.list.Render(t, len(lines), width, contentHeight, true, components.PlainRows(t, lines))
	body = append(body, tail...)
	body = append(body, "", p.hint(t))
	return frame(t, width, height, p.title, body...)
}

func (p *Preferences) rowText(r row) string {
	switch r {
	case rowVolumeStep:
		return fmt.Sprintf("%-28s %d%%", "Volume step", p.settings.Audio.VolumeStep.Value())
	case rowThemeMode:
		return fmt.Sprintf("%-28s %s", "Theme", p.settings.Theme.Mode.String())
	case rowAnimations:
		return fmt.Sprintf("%-28s %s", "Animations", onOff(p.settings.Interface.Animations))
	case rowSidebarHidden:
		return fmt.Sprintf("%-28s %s", "Start with sidebar hidden", onOff(p.settings.Interface.SidebarHidden))
	}
	return ""
}

func (p *Preferences) hint(t styles.Theme) string {
	return ui.Hints(t.Body.Hint,
		ui.Key{Name: icons.UpDown, Action: "move"},
		ui.Key{Name: icons.LeftRight, Action: "change"},
	)
}

func onOff(b bool) string {
	if b {
		return "on"
	}
	return "off"
}

func indexOf(values []string, v string) int {
	for i, candidate := range values {
		if candidate == v {
			return i
		}
	}
	return 0
}

func clamp(v, lo, hi int) int {
	return min(max(v, lo), hi)
}
