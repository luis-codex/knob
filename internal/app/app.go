// Package app es el único tea.Model de la aplicación: mantiene el foco,
// enruta el teclado y compone el layout.
package app

import (
	"context"

	tea "charm.land/bubbletea/v2"

	"settings-cli/internal/application/devices"
	"settings-cli/internal/application/sound"
	"settings-cli/internal/pages"
	"settings-cli/internal/shared/components"
	"settings-cli/internal/shared/icons"
	"settings-cli/internal/shared/layouts"
	"settings-cli/internal/shared/styles"
	"settings-cli/internal/shared/ui"
)

// Contratos opcionales de las páginas. Sin exportar: los declara quien los
// consume.
type (
	// initializer necesita cargar algo al arrancar.
	initializer interface {
		Init() tea.Cmd
	}

	// keyConsumer gestiona su propio teclado. true = tecla consumida.
	keyConsumer interface {
		HandleKey(msg tea.KeyPressMsg) (bool, tea.Cmd)
	}

	// msgConsumer aplica los resultados de sus comandos.
	msgConsumer interface {
		HandleMsg(msg tea.Msg) tea.Cmd
	}

	// overlayProvider expone el modal a componer sobre el body, o nil.
	overlayProvider interface {
		Overlay() layouts.Section
	}
)

// focus indica quién recibe el teclado.
type focus int

const (
	focusSidebar focus = iota
	focusBody
)

type Model struct {
	theme styles.Theme
	// palette son los colores del usuario, con una variante por fondo. Se
	// guardan porque el tema se reconstruye al saber si el fondo es claro u
	// oscuro.
	palette styles.Custom
	layout  layouts.App
	nav     components.Nav
	router  map[string]layouts.Section
	// fallback evita el nil deref del layout ante un ID desconocido.
	fallback layouts.Section
	focus    focus

	width  int
	height int
}

// New recibe el cableado ya hecho: el contexto que acota los procesos que
// escuchan al sistema, la paleta del usuario y los casos de uso. Todo eso lo
// decide el composition root, no la interfaz.
func New(ctx context.Context, custom styles.Custom, deviceSvc *devices.Service, soundSvc *sound.Service) Model {
	return Model{
		palette:  custom,
		theme:    styles.NewWithPalette(true, custom.For(true)), // provisional hasta el BackgroundColorMsg
		nav:      components.NewNav(navGroups()...),
		router:   newRouter(ctx, deviceSvc, soundSvc),
		fallback: pages.NewFallback("Settings"),
	}
}

func (m Model) Init() tea.Cmd {
	// RequestBackgroundColor elige la paleta; cada página carga lo suyo.
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

// broadcast reparte los mensajes que no son de la app entre las páginas. Va a
// todas y no solo a la activa porque un comando puede terminar después de
// navegar fuera; cada página ignora los tipos que no son suyos.
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
	// ctrl+c sale siempre, incluso con un modal abierto.
	if msg.String() == "ctrl+c" {
		return m, tea.Quit
	}

	if m.focus == focusBody {
		return m.handleBodyKey(msg)
	}
	return m.handleSidebarKey(msg.String())
}

func (m Model) handleBodyKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	// La página va primero: así escribir "q" en un campo no cierra la app.
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
		// Entrar solo tiene sentido si la página gestiona teclado.
		if _, ok := m.currentPage().(keyConsumer); ok {
			m.focus = focusBody
		}
	}
	return m, nil
}

// currentPage resuelve qué página toca según la selección del sidebar.
func (m Model) currentPage() layouts.Section {
	if page, ok := m.router[m.nav.Selected().ID]; ok {
		return page
	}
	return m.fallback
}

func (m Model) View() tea.View {
	v := tea.NewView("")
	v.AltScreen = true // en v2 el altscreen es del View, no del Program

	if m.width == 0 {
		return v // aún no llegó el primer WindowSizeMsg
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

// keys son las pistas globales. Dentro de una página solo se anuncia la
// salida: el resto de teclas las explica la propia página.
func (m Model) keys() []ui.Key {
	if m.focus == focusBody {
		return []ui.Key{{Name: icons.Escape, Action: "volver al menú"}}
	}

	return []ui.Key{
		{Name: icons.UpDown, Action: "navegar"},
		{Name: icons.Enter, Action: "entrar"},
		{Name: "q", Action: "salir"},
	}
}
