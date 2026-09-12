package app

import (
	"context"
	"image/color"
	"testing"

	tea "charm.land/bubbletea/v2"

	"knob/internal/application/prefs"
	"knob/internal/application/sound"
	"knob/internal/domain/preferences"
	"knob/internal/pages"
	"knob/internal/shared/ui"
)

var (
	keyCtrlB = tea.KeyPressMsg{Mod: tea.ModCtrl, Code: 'b'}
	keyEsc   = tea.KeyPressMsg{Code: tea.KeyEscape}
)

func testModel() Model {
	return New(context.Background(), prefs.UseCases{}, preferences.Default(), sound.UseCases{})
}

func update(m Model, msg tea.Msg) Model {
	next, _ := m.Update(msg)
	return next.(Model)
}

func TestSidebarToggle(t *testing.T) {
	m := testModel()
	if m.sidebarHidden {
		t.Fatal("the default preference starts with the sidebar visible")
	}

	m = update(m, keyCtrlB)
	if !m.sidebarHidden {
		t.Fatal("ctrl+b did not hide the sidebar")
	}
	if m.focus != focusBody {
		t.Errorf("hiding the sidebar must move focus to the body, got %v", m.focus)
	}

	m = update(m, keyCtrlB)
	if m.sidebarHidden {
		t.Fatal("ctrl+b did not bring the sidebar back")
	}
}

// Moving back to the menu restores it even after Ctrl-B hid it.
func TestReturnToMenuRestoresHiddenSidebar(t *testing.T) {
	m := testModel()
	m = update(m, keyCtrlB)
	if !m.sidebarHidden || m.focus != focusBody {
		t.Fatalf("precondition: hidden sidebar, body focus; got hidden=%v focus=%v", m.sidebarHidden, m.focus)
	}

	m = update(m, keyEsc)
	if m.sidebarHidden {
		t.Error("esc from the body must restore the hidden sidebar")
	}
	if m.focus != focusSidebar {
		t.Errorf("esc must focus the sidebar, got %v", m.focus)
	}
}

// The footer advertises the toggle, with a label that follows the state.
func TestFooterAnnouncesTheToggle(t *testing.T) {
	m := testModel()

	if got := m.keys(); !hasKey(got, "^b", "hide menu") {
		t.Errorf("visible sidebar: footer = %+v, want a \"^b hide menu\" hint", got)
	}

	m = update(m, keyCtrlB)
	if got := m.keys(); !hasKey(got, "^b", "show menu") {
		t.Errorf("hidden sidebar: footer = %+v, want a \"^b show menu\" hint", got)
	}
}

// The sidebar's starting state comes from the stored preference, not always
// visible.
func TestSidebarHiddenStartsFromPreference(t *testing.T) {
	initial := preferences.Default()
	initial.Interface.SidebarHidden = true

	m := New(context.Background(), prefs.UseCases{}, initial, sound.UseCases{})
	if !m.sidebarHidden {
		t.Error("sidebarHidden must start true when the preference says so")
	}
}

// An explicit mode overrides whatever the terminal reports, and does so
// immediately -- no restart needed.
func TestThemeModeOverridesDetectedBackground(t *testing.T) {
	initial := preferences.Default()
	initial.Theme.Mode, _ = preferences.NewMode(preferences.ModeDark)

	m := testModelWith(initial)
	m = update(m, tea.BackgroundColorMsg{Color: color.White}) // a light terminal

	if !m.isDark() {
		t.Error("an explicit dark mode must win over a light terminal")
	}
}

// Saving a preference recolors the running app immediately, not just on the
// next start.
func TestPreferencesSavedMsgRecolorsImmediately(t *testing.T) {
	m := testModel()

	dark, _ := preferences.NewMode(preferences.ModeDark)
	next := preferences.Default()
	next.Theme.Mode = dark

	m = update(m, pages.PreferencesSavedMsg{Settings: next})
	if !m.mode.IsDark() {
		t.Errorf("mode = %q, want dark", m.mode.String())
	}
}

func testModelWith(s preferences.Settings) Model {
	return New(context.Background(), prefs.UseCases{}, s, sound.UseCases{})
}

func hasKey(keys []ui.Key, name, action string) bool {
	for _, k := range keys {
		if k.Name == name && k.Action == action {
			return true
		}
	}
	return false
}
