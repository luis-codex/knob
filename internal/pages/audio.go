package pages

import (
	"context"
	"time"

	tea "charm.land/bubbletea/v2"

	"settings-cli/internal/application/sound"
	"settings-cli/internal/domain/audio"
	"settings-cli/internal/shared/components"
	"settings-cli/internal/shared/icons"
	"settings-cli/internal/shared/styles"
	"settings-cli/internal/shared/ui"
	audioui "settings-cli/internal/ui/audio"
)

const (
	// volumeStep es cuánto mueve una pulsación de ←/→.
	volumeStep = 5
	// La fila es una rejilla de columnas fijas. Si alguna dependiera del
	// contenido —el nivel ocupa 3 o 4 caracteres—, la barra bailaría de una
	// fila a otra y nada quedaría alineado.
	markerColumn = 2 // marcador de predeterminado
	gapColumn    = 2 // separación entre columnas
	levelColumn  = 5 // "100%", "45%" o "mudo", alineados a la derecha
	maxAudioName = 38
	minAudioName = 12

	// blinkInterval es cada cuánto se enciende y se apaga el punto de los
	// flujos que suenan.
	blinkInterval = 600 * time.Millisecond
)

// section distingue las dos listas de la página. El foco está siempre en una.
type section int

const (
	sectionOutputs section = iota
	sectionInputs
	sectionStreams
)

// sections es el orden de pantalla; tab avanza por él.
var sections = []section{sectionOutputs, sectionInputs, sectionStreams}

var sectionLabels = map[section]string{
	sectionOutputs: "Salidas",
	sectionInputs:  "Micrófonos",
	sectionStreams: "Mezclador",
}

type (
	audioLoadedMsg struct {
		outputs []audio.Device
		inputs  []audio.Device
		streams []audio.Stream
	}
	audioChangedMsg struct{}
	audioFailedMsg  struct{ err error }

	// audioBlinkMsg alterna el punto de los flujos que suenan. Se rearma solo
	// mientras haya algo sonando.
	audioBlinkMsg struct{}

	// audioWatchingMsg entrega el canal de avisos. Suscribirse es E/S, así
	// que se hace en un comando y el canal se guarda al recibirlo: el estado
	// solo cambia en HandleMsg.
	audioWatchingMsg struct{ changes <-chan struct{} }
	// audioExternalMsg es un cambio hecho fuera de la aplicación.
	audioExternalMsg struct{}
)

// Audio administra salidas y micrófonos en una sola pantalla, con una lista
// por sección y el foco en una de ellas.
type Audio struct {
	title string
	svc   *sound.Service
	// ctx acota la suscripción a la vida de la aplicación. Sin él el proceso
	// que escucha al servidor de sonido queda huérfano al salir.
	ctx context.Context

	outputs []audio.Device
	inputs  []audio.Device
	streams []audio.Stream

	// Una lista por sección: cada una guarda su propio cursor, de modo que
	// cambiar de sección no pierde dónde estabas.
	lists   map[section]*components.List
	focused section

	// changes avisa de cambios externos: teclas de volumen, un mezclador
	// gráfico, enchufar auriculares.
	changes <-chan struct{}
	failure string

	// blinkOn enciende el punto de los flujos que suenan; blinking evita
	// armar dos temporizadores a la vez.
	blinkOn  bool
	blinking bool
}

func NewAudio(ctx context.Context, title string, svc *sound.Service) *Audio {
	return &Audio{
		title: title,
		svc:   svc,
		ctx:   ctx,
		lists: map[section]*components.List{
			sectionOutputs: components.NewList(),
			sectionInputs:  components.NewList(),
			sectionStreams: components.NewList(),
		},
	}
}

// Init carga el estado y se suscribe a los cambios externos.
func (p *Audio) Init() tea.Cmd {
	return tea.Batch(p.reload(), p.subscribe())
}

// subscribe abre la escucha, acotada al contexto de la aplicación.
func (p *Audio) subscribe() tea.Cmd {
	svc, ctx := p.svc, p.ctx

	return func() tea.Msg {
		changes, err := svc.Changes(ctx)
		if err != nil {
			return audioFailedMsg{err: err}
		}
		return audioWatchingMsg{changes: changes}
	}
}

// waitForChange se bloquea hasta el siguiente aviso. Hay que rearmarlo tras
// cada uno: un comando de Bubble Tea se ejecuta una sola vez.
func waitForChange(changes <-chan struct{}) tea.Cmd {
	if changes == nil {
		return nil
	}

	return func() tea.Msg {
		if _, ok := <-changes; !ok {
			return nil // el canal se cerró: se deja de escuchar
		}
		return audioExternalMsg{}
	}
}

// blinkTick programa el siguiente cambio del punto de reproducción.
func blinkTick() tea.Cmd {
	return tea.Tick(blinkInterval, func(time.Time) tea.Msg { return audioBlinkMsg{} })
}

// ensureBlink arranca el parpadeo si hay algún flujo sonando y no está ya en
// marcha. Sin la guarda, cada recarga encadenaría un temporizador más.
func (p *Audio) ensureBlink() tea.Cmd {
	if p.blinking || !p.anyStreamAudible() {
		return nil
	}
	p.blinking, p.blinkOn = true, true
	return blinkTick()
}

// anyStreamAudible indica si algún flujo suena de verdad: ni en pausa ni
// silenciado.
func (p *Audio) anyStreamAudible() bool {
	for _, s := range p.streams {
		if !s.Paused() && !s.Muted() {
			return true
		}
	}
	return false
}

// --- comandos ---------------------------------------------------------------

func (p *Audio) reload() tea.Cmd {
	svc := p.svc

	return func() tea.Msg {
		ctx := context.Background()

		outputs, err := svc.Outputs(ctx)
		if err != nil {
			return audioFailedMsg{err: err}
		}

		inputs, err := svc.Inputs(ctx)
		if err != nil {
			return audioFailedMsg{err: err}
		}

		streams, err := svc.Streams(ctx)
		if err != nil {
			return audioFailedMsg{err: err}
		}
		return audioLoadedMsg{outputs: outputs, inputs: inputs, streams: streams}
	}
}

// audioMutate envuelve una operación de escritura: si va bien devuelve
// audioChangedMsg para que la página se relea, y si falla, audioFailedMsg.
// Cada página tiene la suya porque el mensaje de éxito es distinto.
func audioMutate(op func(context.Context) error) tea.Cmd {
	return func() tea.Msg {
		if err := op(context.Background()); err != nil {
			return audioFailedMsg{err: err}
		}
		return audioChangedMsg{}
	}
}

// --- estado -----------------------------------------------------------------

func (p *Audio) devices(s section) []audio.Device {
	if s == sectionInputs {
		return p.inputs
	}
	return p.outputs
}

func (p *Audio) list(s section) *components.List { return p.lists[s] }

func (p *Audio) selected() (audio.Device, bool) {
	items := p.devices(p.focused)
	if cursor := p.list(p.focused).Cursor(); cursor < len(items) {
		return items[cursor], true
	}
	return audio.Device{}, false
}

func (p *Audio) selectedStream() (audio.Stream, bool) {
	if cursor := p.list(sectionStreams).Cursor(); cursor < len(p.streams) {
		return p.streams[cursor], true
	}
	return audio.Stream{}, false
}

// --- mensajes ---------------------------------------------------------------

func (p *Audio) HandleMsg(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case audioLoadedMsg:
		p.outputs, p.inputs, p.streams, p.failure = msg.outputs, msg.inputs, msg.streams, ""
		p.list(sectionOutputs).SetCount(len(p.outputs))
		p.list(sectionInputs).SetCount(len(p.inputs))
		p.list(sectionStreams).SetCount(len(p.streams))
		return p.ensureBlink()

	case audioBlinkMsg:
		p.blinkOn = !p.blinkOn
		// Se para al quedarse sin nada que suene; lo rearma la siguiente
		// recarga.
		if !p.anyStreamAudible() {
			p.blinking = false
			return nil
		}
		return blinkTick()

	case audioChangedMsg:
		return p.reload()

	case audioWatchingMsg:
		p.changes = msg.changes
		return waitForChange(p.changes)

	case audioExternalMsg:
		// Releer y volver a escuchar. El aviso no dice qué cambió, así que
		// se relee entero en vez de deducirlo.
		return tea.Batch(p.reload(), waitForChange(p.changes))

	case audioFailedMsg:
		p.failure = msg.err.Error()
	}
	return nil
}

// --- teclado ----------------------------------------------------------------

func (p *Audio) HandleKey(msg tea.KeyPressMsg) (bool, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		p.list(p.focused).Prev()
	case "down", "j":
		p.list(p.focused).Next()
	case "tab":
		p.focused = (p.focused + 1) % section(len(sections))
	case "shift+tab":
		p.focused = (p.focused - 1 + section(len(sections))) % section(len(sections))
	case "left", "h":
		return true, p.adjust(-volumeStep)
	case "right", "l":
		return true, p.adjust(volumeStep)
	case "m":
		return true, p.toggleMute()
	case "enter":
		return true, p.makeDefault()
	default:
		return false, nil
	}
	return true, nil
}

func (p *Audio) adjust(delta int) tea.Cmd {
	if p.focused == sectionStreams {
		st, ok := p.selectedStream()
		if !ok {
			return nil
		}
		svc, index := p.svc, st.ID().Index()
		return audioMutate(func(ctx context.Context) error {
			_, err := svc.AdjustStreamVolume(ctx, index, delta)
			return err
		})
	}

	d, ok := p.selected()
	if !ok {
		return nil
	}

	svc, id := p.svc, d.ID().String()
	return audioMutate(func(ctx context.Context) error {
		_, err := svc.AdjustVolume(ctx, id, delta)
		return err
	})
}

func (p *Audio) toggleMute() tea.Cmd {
	if p.focused == sectionStreams {
		st, ok := p.selectedStream()
		if !ok {
			return nil
		}
		svc, index := p.svc, st.ID().Index()
		return audioMutate(func(ctx context.Context) error {
			_, err := svc.ToggleStreamMuted(ctx, index)
			return err
		})
	}

	d, ok := p.selected()
	if !ok {
		return nil
	}

	svc, id := p.svc, d.ID().String()
	return audioMutate(func(ctx context.Context) error {
		_, err := svc.ToggleMuted(ctx, id)
		return err
	})
}

// makeDefault no aplica a los flujos: nadie elige "el flujo por defecto".
func (p *Audio) makeDefault() tea.Cmd {
	if p.focused == sectionStreams {
		return nil
	}

	d, ok := p.selected()
	if !ok || d.IsDefault() {
		return nil
	}

	svc, id := p.svc, d.ID().String()
	return audioMutate(func(ctx context.Context) error {
		_, err := svc.MakeDefault(ctx, id)
		return err
	})
}

// --- render -----------------------------------------------------------------

func (p *Audio) View(t styles.Theme, width, height int) string {
	inner := width - t.Body.Base.GetHorizontalFrameSize()
	if inner <= 0 {
		return ""
	}

	tail := []string{}
	if p.failure != "" {
		tail = append(tail, "", t.Body.Danger.Render(fit(p.failure, inner)))
	}
	tail = append(tail, "", fit(p.hint(t), inner))

	// Las secciones reparten el alto restante. Los divisores van entre ellas,
	// así que son una fila menos que secciones, más su línea en blanco.
	dividers := (len(sections) - 1) * 2
	available := height - frameChrome - len(tail) - len(sections)*audioui.Chrome - dividers
	listHeight := max(1, available/len(sections))

	rows := make([]string, 0, height)
	for i, s := range sections {
		if i > 0 {
			// Nunca tras la última: ahí ya separa el blanco que precede al pie.
			rows = append(rows, ui.HDivider(t, inner), "")
		}
		rows = append(rows, p.section(t, s).Render(t, inner, listHeight)...)
	}
	rows = append(rows, tail...)

	return frame(t, width, height, p.title, rows...)
}

// section arma la sección que toca con su modelo de vista.
func (p *Audio) section(t styles.Theme, s section) audioui.Section {
	out := audioui.Section{
		Label:   sectionLabels[s],
		List:    p.list(s),
		Focused: p.focused == s,
	}

	if s == sectionStreams {
		out.Empty = "Nada reproduciéndose."
		out.Meters = make([]audioui.Meter, len(p.streams))
		for i, stream := range p.streams {
			out.Meters[i] = streamMeter(t, stream, p.blinkOn)
		}
		return out
	}

	items := p.devices(s)
	out.Empty = "Sin dispositivos."
	out.Meters = make([]audioui.Meter, len(items))
	for i, d := range items {
		out.Meters[i] = deviceMeter(t, d)
	}
	return out
}

// deviceMeter y streamMeter traducen el dominio al modelo de vista. Es el
// único punto donde la pantalla decide qué se enseña de cada cosa.
func deviceMeter(t styles.Theme, d audio.Device) audioui.Meter {
	m := audioui.Meter{
		Name:  d.Name().String(),
		Level: d.Volume().Level(),
		Muted: d.Muted(),
	}
	if d.IsDefault() {
		m.Marker, m.Accent = t.Icon.Active, true
	}
	return m
}

func streamMeter(t styles.Theme, s audio.Stream, blinkOn bool) audioui.Meter {
	name := s.App().String()
	if title := s.Title(); title != "" {
		name += " · " + title
	}

	m := audioui.Meter{
		Name:  name,
		Level: s.Volume().Level(),
		Muted: s.Muted(),
	}

	// El hueco del indicador cuenta lo que se oye: la pausa lo ocupa fija y
	// sin acento porque avisa, no destaca; un flujo que suena lo hace
	// parpadear. Silenciado no enseña nada: no sale sonido.
	switch {
	case s.Paused():
		m.Marker = t.Icon.Paused
	case !s.Muted() && blinkOn:
		m.Marker, m.Accent = t.Icon.Playing, true
	}
	return m
}

func (p *Audio) hint(t styles.Theme) string {
	keys := []ui.Key{
		{Name: icons.UpDown, Action: "mover"},
		{Name: icons.Tab, Action: "sección"},
		{Name: icons.LeftRight, Action: "volumen"},
		{Name: "m", Action: "silenciar"},
	}

	// En el mezclador no hay predeterminado que elegir.
	if p.focused != sectionStreams {
		keys = append(keys, ui.Key{Name: icons.Enter, Action: "predeterminado"})
	}
	return ui.Hints(t.Body.Hint, keys...)
}
