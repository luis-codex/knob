// Package app is the application's single tea.Model: it keeps focus, routes
// the keyboard and composes the layout.
package app

import (
	"context"

	tea "charm.land/bubbletea/v2"

	"knob/internal/application/devices"
	"knob/internal/application/sound"
	"knob/internal/pages"
	"knob/internal/shared/components"
	"knob/internal/shared/icons"
	"knob/internal/shared/layouts"
	"knob/internal/shared/styles"
	"knob/internal/shared/ui"
)

// Optional page contracts. Unexported: declared by whoever consumes them.
type (
	// initializer needs to load something at startup.
	initializer interface {
		Init() tea.Cmd
	}

	// keyConsumer handles its own keyboard. true = key consumed.
	keyConsumer interface {
		HandleKey(msg tea.KeyPressMsg) (bool, tea.Cmd)
	}

	// msgConsumer applies the results of its commands.
	msgConsumer interface {
		HandleMsg(msg tea.Msg) tea.Cmd
	}

	// overlayProvider exposes the modal to compose over the body, or nil.
	overlayProvider interface {
		Overlay() layouts.Section
	}
)

// focus tells who receives the keyboard.
type focus int

const (
	focusSidebar focus = iota
	focusBody
)

type Model struct {
	theme styles.Theme
	// palette is the user's colors, with one variant per background. It is
	// kept because the theme is rebuilt once the background is known to be
	// light or dark.
	palette styles.Custom
	layout  layouts.App
	nav     components.Nav
	router  map[string]layouts.Section
	// fallback avoids the layout's nil deref on an unknown ID.
	fallback layouts.Section
	focus    focus

	width  int
	height int
}

// New takes the wiring already done: the context that bounds the processes
// listening to the system, the user's palette, the use cases and the
// preferences read from config.toml. All of that is decided by the composition
// root, not the interface.
func New(ctx context.Context, custom styles.Custom, deviceUC devices.UseCases, soundUC sound.UseCases, volumeStep int) Model {
	return Model{
		palette:  custom,
		theme:    styles.NewWithPalette(true, custom.For(true)), // provisional until the BackgroundColorMsg
		nav:      components.NewNav(navGroups()...),
		router:   newRouter(ctx, deviceUC, soundUC, volumeStep),
		fallback: pages.NewFallback("Settings"),
	}
}

func (m Model) Init() tea.Cmd {
	// RequestBackgroundColor picks the palette; each page loads its own thing.
	cmds := []tea.Cmd{tea.RequestBackgroundColor}
	for _, page := range m.router {
		if p, ok := page.(initializer); ok {
			cmds = append(cmds, p.Init())
		}
	}
	return tea.Batch(cmds...)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil

	case tea.BackgroundColorMsg:
		m.theme = styles.NewWithPalette(msg.IsDark(), m.palette.For(msg.IsDark()))
		return m, nil

	case tea.KeyPressMsg:
		return m.handleKey(msg)
	}

	return m, m.broadcast(msg)
}

// broadcast hands the messages that are not the app's to the pages. It goes to
// all of them and not just the active one because a command can finish after
// navigating away; each page ignores the types that are not its own.
func (m Model) broadcast(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd
	for _, page := range m.router {
		if p, ok := page.(msgConsumer); ok {
			cmds = append(cmds, p.HandleMsg(msg))
		}
	}
	return tea.Batch(cmds...)
}

func (m Model) handleKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	// ctrl+c always quits, even with a modal open.
	if msg.String() == "ctrl+c" {
		return m, tea.Quit
	}

	if m.focus == focusBody {
		return m.handleBodyKey(msg)
	}
	return m.handleSidebarKey(msg.String())
}

func (m Model) handleBodyKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	// The page goes first: that way typing "q" in a field does not close the
	// app.
	if page, ok := m.currentPage().(keyConsumer); ok {
		if consumed, cmd := page.HandleKey(msg); consumed {
			return m, cmd
		}
	}

	switch msg.String() {
	case "esc", "tab", "left", "h":
		m.focus = focusSidebar
	case "q":
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) handleSidebarKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "q", "esc":
		return m, tea.Quit
	case "up", "k":
		m.nav = m.nav.Prev()
	case "down", "j":
		m.nav = m.nav.Next()
	case "tab", "enter", "right", "l":
		// Entering only makes sense if the page handles the keyboard.
		if _, ok := m.currentPage().(keyConsumer); ok {
			m.focus = focusBody
		}
	}
	return m, nil
}

// currentPage resolves which page is due based on the sidebar selection.
func (m Model) currentPage() layouts.Section {
	if page, ok := m.router[m.nav.Selected().ID]; ok {
		return page
	}
	return m.fallback
}

func (m Model) View() tea.View {
	v := tea.NewView("")
	v.AltScreen = true // in v2 the alt-screen belongs to the View, not the Program

	if m.width == 0 {
		return v // the first WindowSizeMsg has not arrived yet
	}

	page := m.currentPage()

	l := m.layout
	l.Header = components.NewHeader().WithSection(m.nav.Selected().Label)
	l.Sidebar = m.nav.WithFocus(m.focus == focusSidebar)
	l.Body = page
	if provider, ok := page.(overlayProvider); ok {
		l.Overlay = provider.Overlay()
	}
	l.Footer = components.NewFooter(m.keys()...)

	v.SetContent(l.View(m.theme, m.width, m.height))
	return v
}

// keys are the global hints. Inside a page only the exit is announced: the
// rest of the keys are explained by the page itself.
func (m Model) keys() []ui.Key {
	if m.focus == focusBody {
		return []ui.Key{{Name: icons.Escape, Action: "back to the menu"}}
	}

	return []ui.Key{
		{Name: icons.UpDown, Action: "navigate"},
		{Name: icons.Enter, Action: "enter"},
		{Name: "q", Action: "quit"},
	}
}
