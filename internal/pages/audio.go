package pages

import (
	"context"
	"time"

	tea "charm.land/bubbletea/v2"

	"knob/internal/application/sound"
	"knob/internal/domain/audio"
	"knob/internal/shared/components"
	"knob/internal/shared/icons"
	"knob/internal/shared/styles"
	"knob/internal/shared/ui"
	audioui "knob/internal/ui/audio"
)

const (
	// volumeStep is how much a left/right press moves.
	volumeStep = 5
	// blinkInterval is how often the dot of audible streams turns on and off.
	blinkInterval = 600 * time.Millisecond
)

// section tells the page's two lists apart. Focus is always on one of them.
type section int

const (
	sectionOutputs section = iota
	sectionInputs
	sectionStreams
)

// sections is the on-screen order; tab moves through it.
var sections = []section{sectionOutputs, sectionInputs, sectionStreams}

var sectionLabels = map[section]string{
	sectionOutputs: "Outputs",
	sectionInputs:  "Microphones",
	sectionStreams: "Mixer",
}

type (
	audioLoadedMsg struct {
		outputs []audio.Device
		inputs  []audio.Device
		streams []audio.Stream
	}
	audioChangedMsg struct{}
	audioFailedMsg  struct{ err error }

	// audioBlinkMsg toggles the dot of audible streams. It re-arms itself only
	// while something is playing.
	audioBlinkMsg struct{}

	// audioWatchingMsg delivers the notification channel. Subscribing is I/O,
	// so it happens in a command and the channel is stored on receipt: state
	// only changes in HandleMsg.
	audioWatchingMsg struct{ changes <-chan struct{} }
	// audioExternalMsg is a change made outside the application.
	audioExternalMsg struct{}
)

// Audio manages outputs and microphones on a single screen, with one list per
// section and focus on one of them.
type Audio struct {
	title string
	svc   *sound.Service
	// ctx bounds the subscription to the application's lifetime. Without it
	// the process listening to the sound server is orphaned on exit.
	ctx context.Context

	outputs []audio.Device
	inputs  []audio.Device
	streams []audio.Stream

	// One list per section: each keeps its own cursor, so switching sections
	// does not lose where you were.
	lists   map[section]*components.List
	focused section

	// changes reports external changes: volume keys, a graphical mixer,
	// plugging in headphones.
	changes <-chan struct{}
	failure string

	// blinkOn turns on the dot of audible streams; blinking prevents arming
	// two timers at once.
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

// Init loads the state and subscribes to external changes.
func (p *Audio) Init() tea.Cmd {
	return tea.Batch(p.reload(), p.subscribe())
}

// subscribe opens the listener, bounded to the application's context.
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

// waitForChange blocks until the next notification. It must be re-armed after
// each one: a Bubble Tea command runs exactly once.
func waitForChange(changes <-chan struct{}) tea.Cmd {
	if changes == nil {
		return nil
	}

	return func() tea.Msg {
		if _, ok := <-changes; !ok {
			return nil // the channel closed: stop listening
		}
		return audioExternalMsg{}
	}
}

// blinkTick schedules the next change of the playback dot.
func blinkTick() tea.Cmd {
	return tea.Tick(blinkInterval, func(time.Time) tea.Msg { return audioBlinkMsg{} })
}

// ensureBlink starts the blink if some stream is playing and it is not already
// running. Without the guard, each reload would chain one more timer.
func (p *Audio) ensureBlink() tea.Cmd {
	if p.blinking || !p.anyStreamAudible() {
		return nil
	}
	p.blinking, p.blinkOn = true, true
	return blinkTick()
}

// anyStreamAudible reports whether some stream is actually audible: neither
// paused nor muted.
func (p *Audio) anyStreamAudible() bool {
	for _, s := range p.streams {
		if !s.Paused() && !s.Muted() {
			return true
		}
	}
	return false
}

// --- commands -------------------------------------------------------------

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

// audioMutate wraps a write operation: on success it returns audioChangedMsg
// so the page re-reads itself, and on failure, audioFailedMsg. Each page has
// its own because the success message differs.
func audioMutate(op func(context.Context) error) tea.Cmd {
	return func() tea.Msg {
		if err := op(context.Background()); err != nil {
			return audioFailedMsg{err: err}
		}
		return audioChangedMsg{}
	}
}

// --- state --------------------------------------------------------------

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

// --- messages -----------------------------------------------------------

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
		// It stops once nothing is audible; the next reload re-arms it.
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
		// Re-read and listen again. The notification does not say what
		// changed, so it is re-read whole rather than inferred.
		return tea.Batch(p.reload(), waitForChange(p.changes))

	case audioFailedMsg:
		p.failure = msg.err.Error()
	}
	return nil
}

// --- keyboard ---------------------------------------------------------------

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

// makeDefault does not apply to streams: nobody picks "the default stream".
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

// --- render -------------------------------------------------------------

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

	// The sections split the remaining height. The dividers go between them,
	// so they are one row fewer than sections, plus their blank line.
	dividers := (len(sections) - 1) * 2
	available := height - frameChrome - len(tail) - len(sections)*audioui.Chrome - dividers
	listHeight := max(1, available/len(sections))

	rows := make([]string, 0, height)
	for i, s := range sections {
		if i > 0 {
			// Never after the last one: there the blank before the footer
			// already separates it.
			rows = append(rows, ui.HDivider(t, inner), "")
		}
		rows = append(rows, p.section(t, s).Render(t, inner, listHeight)...)
	}
	rows = append(rows, tail...)

	return frame(t, width, height, p.title, rows...)
}

// section builds the relevant section with its view model.
func (p *Audio) section(t styles.Theme, s section) audioui.Section {
	out := audioui.Section{
		Label:   sectionLabels[s],
		List:    p.list(s),
		Focused: p.focused == s,
	}

	if s == sectionStreams {
		out.Empty = "Nothing playing."
		out.Meters = make([]audioui.Meter, len(p.streams))
		for i, stream := range p.streams {
			out.Meters[i] = streamMeter(t, stream, p.blinkOn)
		}
		return out
	}

	items := p.devices(s)
	out.Empty = "No devices."
	out.Meters = make([]audioui.Meter, len(items))
	for i, d := range items {
		out.Meters[i] = deviceMeter(t, d)
	}
	return out
}

// deviceMeter and streamMeter translate the domain to the view model. It is
// the only place where the screen decides what is shown of each thing.
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

	// The indicator slot reflects what is heard: a pause holds it fixed and
	// without accent because it warns rather than stands out; a playing stream
	// makes it blink. Muted shows nothing: no sound comes out.
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
		{Name: icons.UpDown, Action: "move"},
		{Name: icons.Tab, Action: "section"},
		{Name: icons.LeftRight, Action: "volume"},
		{Name: "m", Action: "mute"},
	}

	// In the mixer there is no default to pick.
	if p.focused != sectionStreams {
		keys = append(keys, ui.Key{Name: icons.Enter, Action: "default"})
	}
	return ui.Hints(t.Body.Hint, keys...)
}
