// Package styles concentra los tokens visuales: espaciado, tamaños, colores
// semánticos y los estilos ya resueltos. Ningún otro paquete declara colores
// ni paddings.
package styles

import (
	"image/color"

	"charm.land/lipgloss/v2"

	"settings-cli/internal/shared/icons"
)

// Tokens de espaciado y tamaño: el ritmo visual de la app.
const (
	PadX = 1 // padding horizontal de cada sección
	PadY = 0 // las secciones ya se separan con dividers

	SidebarWidth = 26
	HeaderHeight = 1
	FooterHeight = 1
	DividerSize  = 1 // grosor de un divider

	DialogWidth = 50
)

// Palette nombra los colores por función (Accent, Muted) y no por tono, para
// que cambiar la paleta no obligue a renombrar.
//
// Los valores oscuros son los de bun.com: fondo casi negro de matiz cálido,
// tres niveles de texto y un magenta usado con cuentagotas.
type Palette struct {
	Text  color.Color
	Muted color.Color
	// Faint es el tercer nivel de texto: etiquetas y metadatos que solo se
	// leen si se buscan.
	Faint color.Color

	Line color.Color
	// LineStrong enmarca lo que sí debe destacar, como un modal.
	LineStrong color.Color

	Accent color.Color
	// AccentStrong es el acento enfatizado o sobre fondo de acento.
	AccentStrong color.Color
	OnAccent     color.Color // texto legible sobre Accent

	Bg         color.Color
	BgSelected color.Color

	Success color.Color
	Warning color.Color
	Danger  color.Color
}

type headerStyles struct {
	Base  lipgloss.Style
	Title lipgloss.Style
	// Crumb es la sección actual, a la derecha del nombre.
	Crumb lipgloss.Style
}

type navStyles struct {
	Base     lipgloss.Style
	Group    lipgloss.Style
	Item     lipgloss.Style
	Selected lipgloss.Style
	// Blurred es el item seleccionado sin foco.
	Blurred lipgloss.Style
}

type inputStyles struct {
	Label       lipgloss.Style
	Blurred     lipgloss.Style
	Focused     lipgloss.Style
	Placeholder lipgloss.Style
	Cursor      lipgloss.Style
}

type scrollbarStyles struct {
	Track lipgloss.Style
	Thumb lipgloss.Style
}

type listStyles struct {
	Item     lipgloss.Style
	Selected lipgloss.Style
	Empty    lipgloss.Style
}

type bodyStyles struct {
	Base   lipgloss.Style
	Title  lipgloss.Style
	Label  lipgloss.Style
	Value  lipgloss.Style
	Muted  lipgloss.Style
	Badge  lipgloss.Style
	Danger lipgloss.Style
	Note   lipgloss.Style
	// Section es la etiqueta de una sección dentro de la página.
	Section lipgloss.Style
	// Hint son las pistas de teclado dentro de la página. La tecla va en
	// texto pleno y no en acento: el acento ya lo lleva el pie, y dos cosas
	// gritando a la vez no jerarquizan nada.
	Hint HintStyles
}

// HintStyles pinta una lista de pistas de teclado.
type HintStyles struct {
	// Key es la tecla en sí, que es lo que el usuario busca.
	Key lipgloss.Style
	// Action es lo que hace esa tecla.
	Action lipgloss.Style
	// Sep separa una pista de la siguiente.
	Sep lipgloss.Style
}

type footerStyles struct {
	Base lipgloss.Style
	Hint HintStyles
}

type buttonStyles struct {
	Blurred lipgloss.Style
	Focused lipgloss.Style
	// Danger es el botón primario de una acción destructiva con el foco.
	Danger lipgloss.Style
}

type dialogStyles struct {
	Box   lipgloss.Style
	Title lipgloss.Style
	Text  lipgloss.Style
	Label lipgloss.Style
	Error lipgloss.Style
	Hint  lipgloss.Style
}

// Icons son los símbolos de la interfaz.
//
// Viven en el tema y no sueltos por el código para que cambiar el juego de
// símbolos sea tocar un solo sitio. Todos deben ocupar una celda: la rejilla
// de las filas cuenta con ello.
type Icons struct {
	// Cursor marca la fila bajo el cursor.
	Cursor string
	// Active marca el elemento en uso: el dispositivo predeterminado.
	Active string
	// Paused marca un flujo de audio detenido.
	Paused string
	// Playing marca un flujo que suena; la pantalla lo hace parpadear.
	Playing string
	// BarOn y BarOff son los tramos lleno y vacío de una barra.
	BarOn  string
	BarOff string
	// DividerH y DividerV separan las regiones del layout.
	DividerH string
	DividerV string
	// ScrollThumb y ScrollTrack son la barra de desplazamiento.
	ScrollThumb string
	ScrollTrack string
	// Ellipsis cierra un texto recortado.
	Ellipsis string
}

func defaultIcons() Icons {
	return Icons{
		Cursor:      icons.Cursor,
		Active:      icons.Active,
		Paused:      icons.Paused,
		Playing:     icons.Playing,
		BarOn:       icons.BarOn,
		BarOff:      icons.BarOff,
		DividerH:    icons.DividerH,
		DividerV:    icons.DividerV,
		ScrollThumb: icons.ScrollThumb,
		ScrollTrack: icons.ScrollTrack,
		Ellipsis:    icons.Ellipsis,
	}
}

// Theme son los estilos resueltos para el fondo de terminal actual.
type Theme struct {
	Color Palette
	Icon  Icons

	Header    headerStyles
	Nav       navStyles
	Body      bodyStyles
	Footer    footerStyles
	Dialog    dialogStyles
	Button    buttonStyles
	Input     inputStyles
	List      listStyles
	Scrollbar scrollbarStyles
	Divider   lipgloss.Style
}

// Custom son las paletas que el usuario define para cada tipo de fondo.
type Custom struct {
	Light Palette
	Dark  Palette
}

// For elige la paleta que toca según el fondo del terminal.
func (c Custom) For(isDark bool) Palette {
	if isDark {
		return c.Dark
	}
	return c.Light
}

// Merge devuelve la paleta con los colores no nulos de custom sustituidos.
//
// Los campos son interfaces, así que el valor cero de Palette significa "no
// personalizar nada": el usuario declara solo los colores que quiere cambiar.
func (p Palette) Merge(custom Palette) Palette {
	set := func(dst *color.Color, src color.Color) {
		if src != nil {
			*dst = src
		}
	}

	set(&p.Text, custom.Text)
	set(&p.Muted, custom.Muted)
	set(&p.Faint, custom.Faint)
	set(&p.Line, custom.Line)
	set(&p.LineStrong, custom.LineStrong)
	set(&p.Accent, custom.Accent)
	set(&p.AccentStrong, custom.AccentStrong)
	set(&p.OnAccent, custom.OnAccent)
	set(&p.Bg, custom.Bg)
	set(&p.BgSelected, custom.BgSelected)
	set(&p.Success, custom.Success)
	set(&p.Warning, custom.Warning)
	set(&p.Danger, custom.Danger)

	return p
}

// New construye el tema con la paleta por defecto.
func New(isDark bool) Theme {
	return NewWithPalette(isDark, Palette{})
}

// NewWithPalette construye el tema aplicando los colores del usuario sobre los
// de por defecto. isDark viene de tea.BackgroundColorMsg.IsDark().
func NewWithPalette(isDark bool, custom Palette) Theme {
	pick := lipgloss.LightDark(isDark)

	p := Palette{
		Text:  pick(lipgloss.Color("#1A1A1A"), lipgloss.Color("#EAEAE8")),
		Muted: pick(lipgloss.Color("#6B6B68"), lipgloss.Color("#A8A8A5")),
		Faint: pick(lipgloss.Color("#94948F"), lipgloss.Color("#80807E")),

		Line:       pick(lipgloss.Color("#E2E2E0"), lipgloss.Color("#28282B")),
		LineStrong: pick(lipgloss.Color("#C9C9C6"), lipgloss.Color("#3E3E42")),

		// Por defecto el acento no es un color, es contraste: blanco sobre
		// oscuro y negro sobre claro. Así la app no impone una identidad y
		// quien quiera color lo pone en su theme.toml.
		//
		// Paletas anteriores, por si se quieren recuperar:
		//	Magenta (bun.com):
		//	Accent:       pick(lipgloss.Color("#D6006E"), lipgloss.Color("#FF2E97")),
		//	AccentStrong: pick(lipgloss.Color("#FF2E97"), lipgloss.Color("#FF5CB0")),
		//	Azul:
		//	Accent:       pick(lipgloss.Color("#3B4FC4"), lipgloss.Color("#516BEB")),
		//	AccentStrong: pick(lipgloss.Color("#516BEB"), lipgloss.Color("#7D91F2")),
		Accent:       pick(lipgloss.Color("#1A1A1A"), lipgloss.Color("#EAEAE8")),
		AccentStrong: pick(lipgloss.Color("#000000"), lipgloss.Color("#FFFFFF")),
		OnAccent:     pick(lipgloss.Color("#FFFFFF"), lipgloss.Color("#0D0A0C")),

		Bg: pick(lipgloss.Color("#FAFAF8"), lipgloss.Color("#0D0A0C")),
		// Con acento neutro el fondo de selección también lo es: un gris que
		// separa la fila sin teñirla. Con magenta era #FFE7F2 / #280016 y con
		// azul #E7EBFD / #111634.
		BgSelected: pick(lipgloss.Color("#EDEDEA"), lipgloss.Color("#1F1F22")),

		Success: pick(lipgloss.Color("#12864F"), lipgloss.Color("#28DC82")),
		Warning: pick(lipgloss.Color("#B45309"), lipgloss.Color("#FBBF24")),
		Danger:  pick(lipgloss.Color("#C0392B"), lipgloss.Color("#FF5C5C")),
	}.Merge(custom)

	// Sin borde propio: la separación entre áreas la ponen los dividers.
	section := lipgloss.NewStyle().Padding(PadY, PadX)

	return Theme{
		Color: p,
		Icon:  defaultIcons(),

		Header: headerStyles{
			Base:  section,
			Title: lipgloss.NewStyle().Foreground(p.Accent).Bold(true),
			Crumb: lipgloss.NewStyle().Foreground(p.Faint),
		},
		Nav: navStyles{
			Base: section,
			// La cabecera manda sobre lo que agrupa, así que va en texto
			// pleno: más tenue que sus items la hacía desaparecer entre
			// ellos.
			Group:    lipgloss.NewStyle().Foreground(p.Text).Bold(true),
			Item:     lipgloss.NewStyle().Foreground(p.Muted),
			Selected: lipgloss.NewStyle().Foreground(p.AccentStrong).Background(p.BgSelected).Bold(true),
			Blurred:  lipgloss.NewStyle().Foreground(p.Faint).Background(p.BgSelected),
		},
		Body: bodyStyles{
			Base: section,
			// El título de página va en texto pleno, no en acento: el magenta
			// se reserva para lo que hay que mirar.
			Title:  lipgloss.NewStyle().Foreground(p.Text).Bold(true),
			Label:  lipgloss.NewStyle().Foreground(p.Text),
			Value:  lipgloss.NewStyle().Foreground(p.Muted),
			Muted:  lipgloss.NewStyle().Foreground(p.Muted),
			Badge:  lipgloss.NewStyle().Foreground(p.Success),
			Danger: lipgloss.NewStyle().Foreground(p.Danger),
			Note:   lipgloss.NewStyle().Foreground(p.Faint).Italic(true),
			// Mismo criterio que Nav.Group: la etiqueta pesa más que su lista.
			Section: lipgloss.NewStyle().Foreground(p.Text).Bold(true),
			Hint: HintStyles{
				Key:    lipgloss.NewStyle().Foreground(p.Text).Bold(true),
				Action: lipgloss.NewStyle().Foreground(p.Muted),
				Sep:    lipgloss.NewStyle().Foreground(p.Line),
			},
		},
		Footer: footerStyles{
			Base: section,
			// En el pie la tecla va en acento: es la salida de emergencia.
			Hint: HintStyles{
				Key:    lipgloss.NewStyle().Foreground(p.Accent).Bold(true),
				Action: lipgloss.NewStyle().Foreground(p.Muted),
				Sep:    lipgloss.NewStyle().Foreground(p.Line),
			},
		},
		// El modal lleva borde y fondo: es lo que lo despega del body.
		Dialog: dialogStyles{
			Box: lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(p.LineStrong).
				Background(p.Bg).
				Foreground(p.Text).
				Padding(1, 2),
			Title: lipgloss.NewStyle().Foreground(p.Text).Background(p.Bg).Bold(true),
			Text:  lipgloss.NewStyle().Foreground(p.Muted).Background(p.Bg),
			Label: lipgloss.NewStyle().Foreground(p.Faint).Background(p.Bg),
			Error: lipgloss.NewStyle().Foreground(p.Danger).Background(p.Bg),
			Hint:  lipgloss.NewStyle().Foreground(p.Faint).Background(p.Bg),
		},
		Button: buttonStyles{
			Blurred: lipgloss.NewStyle().Foreground(p.Muted).Background(p.BgSelected),
			// Fondo pleno de acento con texto oscuro: la insignia de bun.com.
			Focused: lipgloss.NewStyle().Foreground(p.OnAccent).Background(p.Accent).Bold(true),
			Danger:  lipgloss.NewStyle().Foreground(p.OnAccent).Background(p.Danger).Bold(true),
		},
		Input: inputStyles{
			Label:       lipgloss.NewStyle().Foreground(p.Faint),
			Blurred:     lipgloss.NewStyle().Foreground(p.Muted).Background(p.BgSelected),
			Focused:     lipgloss.NewStyle().Foreground(p.Text).Background(p.BgSelected),
			Placeholder: lipgloss.NewStyle().Foreground(p.Faint).Background(p.BgSelected).Italic(true),
			Cursor:      lipgloss.NewStyle().Foreground(p.OnAccent).Background(p.Accent),
		},
		List: listStyles{
			Item:     lipgloss.NewStyle().Foreground(p.Muted),
			Selected: lipgloss.NewStyle().Foreground(p.Text).Background(p.BgSelected).Bold(true),
			Empty:    lipgloss.NewStyle().Foreground(p.Faint).Italic(true),
		},
		Scrollbar: scrollbarStyles{
			Track: lipgloss.NewStyle().Foreground(p.Line),
			Thumb: lipgloss.NewStyle().Foreground(p.Accent),
		},
		Divider: lipgloss.NewStyle().Foreground(p.Line),
	}
}
