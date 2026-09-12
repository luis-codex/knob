package app

import (
	"context"
	"testing"

	tea "charm.land/bubbletea/v2"

	"knob/internal/application/sound"
	"knob/internal/shared/styles"
	"knob/internal/shared/ui"
)

var (
	keyCtrlB = tea.KeyPressMsg{Mod: tea.ModCtrl, Code: 'b'}
	keyEsc   = tea.KeyPressMsg{Code: tea.KeyEscape}
)

func testModel() Model {
	return New(context.Background(), styles.Custom{}, sound.UseCases{}, 5)
}

func update(m Model, msg tea.Msg) Model {
	next, _ := m.Update(msg)
	return next.(Model)
}

func TestSidebarToggle(t *testing.T) {
	m := testModel()
	if m.sidebarHidden {
		t.Fatal("the sidebar starts visible")
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

func hasKey(keys []ui.Key, name, action string) bool {
	for _, k := range keys {
		if k.Name == name && k.Action == action {
			return true
		}
	}
	return false
}
