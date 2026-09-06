package pages

import (
	"context"
	"strconv"

	tea "charm.land/bubbletea/v2"

	"settings-cli/internal/application/devices"
	"settings-cli/internal/domain/bluetooth"
	"settings-cli/internal/domain/errs"
	"settings-cli/internal/shared/components"
	"settings-cli/internal/shared/icons"
	"settings-cli/internal/shared/layouts"
	"settings-cli/internal/shared/styles"
	"settings-cli/internal/shared/ui"
	btui "settings-cli/internal/ui/bluetooth"
)

// developmentNote avisa de lo que aún no hace esta pantalla. Se quita cuando
// el emparejamiento entrante esté implementado.
const developmentNote = "En desarrollo · falta el emparejamiento entrante"

// devicePending recuerda qué aplicar cuando se acepte el modal.
type devicePending int

const (
	devicePendingNone devicePending = iota
	devicePendingRename
	devicePendingRemove
)

type (
	devicesLoadedMsg struct {
		items   []bluetooth.Device
		adapter bluetooth.Adapter
	}
	deviceSavedMsg   struct{}
	scanFinishedMsg  struct{ added int }
	devicesFailedMsg struct{ err error }
)

// Bluetooth administra los dispositivos emparejados.
type Bluetooth struct {
	title string
	svc   *devices.Service

	items   []bluetooth.Device
	adapter bluetooth.Adapter
	list    *components.List

	search    components.TextInput
	filtering bool

	dialog      components.Dialog
	pending     devicePending
	pendingAddr string

	// scanning bloquea la acción mientras la búsqueda está en curso; el
	// comando corre en su goroutine y la interfaz sigue respondiendo.
	scanning bool
	// lastScan es cuántos dispositivos añadió la última búsqueda. -1 es "aún
	// no se ha buscado", que no es lo mismo que haber encontrado cero.
	lastScan int
	failure  string
}

func NewBluetooth(title string, svc *devices.Service) *Bluetooth {
	return &Bluetooth{title: title, svc: svc, list: components.NewList(), lastScan: -1}
}

func (p *Bluetooth) Init() tea.Cmd { return p.reload() }

// --- comandos ---------------------------------------------------------------

func (p *Bluetooth) reload() tea.Cmd {
	svc, query := p.svc, p.search.Trimmed()

	return func() tea.Msg {
		ctx := context.Background()

		adapter, err := svc.Adapter(ctx)
		if err != nil {
			return devicesFailedMsg{err: err}
		}

		items, err := svc.Search(ctx, query)
		if err != nil {
			return devicesFailedMsg{err: err}
		}
		return devicesLoadedMsg{items: items, adapter: adapter}
	}
}

// scan busca dispositivos cercanos.
func (p *Bluetooth) scan() tea.Cmd {
	svc := p.svc

	return func() tea.Msg {
		added, err := svc.Scan(context.Background())
		if err != nil {
			return devicesFailedMsg{err: err}
		}
		return scanFinishedMsg{added: len(added)}
	}
}

// toggleAdapter enciende o apaga la radio según su estado actual.
func (p *Bluetooth) toggleAdapter() tea.Cmd {
	svc, enabled := p.svc, p.adapter.Enabled()

	return deviceMutate(func(ctx context.Context) error {
		var err error
		if enabled {
			_, err = svc.DisableAdapter(ctx)
		} else {
			_, err = svc.EnableAdapter(ctx)
		}
		return err
	})
}

// deviceMutate envuelve una operación de escritura: si va bien devuelve
// deviceSavedMsg para que la página se relea, y si falla, devicesFailedMsg.
// Cada página tiene la suya porque el mensaje de éxito es distinto.
func deviceMutate(op func(context.Context) error) tea.Cmd {
	return func() tea.Msg {
		if err := op(context.Background()); err != nil {
			return devicesFailedMsg{err: err}
		}
		return deviceSavedMsg{}
	}
}

// advance ejecuta la acción natural para el estado actual: emparejar lo
// descubierto, conectar lo emparejado y desconectar lo conectado.
func (p *Bluetooth) advance(d bluetooth.Device) tea.Cmd {
	svc, addr := p.svc, d.Address().String()

	return deviceMutate(func(ctx context.Context) error {
		var err error
		switch d.State() {
		case bluetooth.StateDiscovered:
			_, err = svc.Pair(ctx, addr)
		case bluetooth.StatePaired:
			_, err = svc.Connect(ctx, addr)
		case bluetooth.StateConnected:
			_, err = svc.Disconnect(ctx, addr)
		}
		return err
	})
}

// --- estado -----------------------------------------------------------------

func (p *Bluetooth) selected() (bluetooth.Device, bool) {
	if cursor := p.list.Cursor(); cursor < len(p.items) {
		return p.items[cursor], true
	}
	return bluetooth.Device{}, false
}

// --- modales ----------------------------------------------------------------

func (p *Bluetooth) openRename(d bluetooth.Device) {
	p.dialog = components.NewFormDialog("Renombrar dispositivo", "Nombre", "¿Cómo se llama?", "Guardar").
		WithValue(d.Name().String()).
		Open()
	p.pending, p.pendingAddr = devicePendingRename, d.Address().String()
}

func (p *Bluetooth) openRemove(d bluetooth.Device) {
	p.dialog = components.NewConfirmDialog(
		"Olvidar dispositivo",
		"Se va a olvidar «"+d.Name().String()+"».\nHabrá que volver a emparejarlo.",
		"Olvidar",
	).Dangerous().Open()
	p.pending, p.pendingAddr = devicePendingRemove, d.Address().String()
}

// resolve traduce la acción del modal en un comando. El diálogo no se cierra
// aquí: sigue abierto hasta que la operación confirma.
func (p *Bluetooth) resolve(action components.DialogAction) tea.Cmd {
	switch action {
	case components.DialogNone:
		return nil
	case components.DialogDismiss:
		p.closeDialog()
		return nil
	}

	svc, addr, value := p.svc, p.pendingAddr, p.dialog.Value()

	switch p.pending {
	case devicePendingRename:
		return deviceMutate(func(ctx context.Context) error {
			_, err := svc.Rename(ctx, addr, value)
			return err
		})
	case devicePendingRemove:
		return deviceMutate(func(ctx context.Context) error {
			return svc.Remove(ctx, addr)
		})
	}
	return nil
}

func (p *Bluetooth) closeDialog() {
	p.dialog = p.dialog.Close()
	p.pending, p.pendingAddr = devicePendingNone, ""
}

func (p *Bluetooth) Overlay() layouts.Section {
	if !p.dialog.IsOpen() {
		return nil
	}
	return p.dialog
}

// --- mensajes ---------------------------------------------------------------

func (p *Bluetooth) HandleMsg(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case devicesLoadedMsg:
		p.items, p.adapter, p.failure = msg.items, msg.adapter, ""
		p.list.SetCount(len(p.items))

	case deviceSavedMsg:
		p.closeDialog()
		return p.reload()

	case scanFinishedMsg:
		p.scanning = false
		p.lastScan = msg.added
		return p.reload()

	case devicesFailedMsg:
		p.scanning = false
		p.fail(msg.err)
	}
	return nil
}

// fail decide dónde va el error. Un fallo de validación pertenece al
// formulario que lo provocó; el resto va al cuerpo de la página.
func (p *Bluetooth) fail(err error) {
	if p.dialog.IsOpen() && errs.IsInvalid(err) {
		p.dialog = p.dialog.WithError(err.Error())
		return
	}
	p.closeDialog()
	p.failure = err.Error()
}

// --- teclado ----------------------------------------------------------------

func (p *Bluetooth) HandleKey(msg tea.KeyPressMsg) (bool, tea.Cmd) {
	switch {
	case p.dialog.IsOpen():
		dialog, action := p.dialog.HandleKey(msg)
		p.dialog = dialog
		return true, p.resolve(action)
	case p.filtering:
		return true, p.handleSearchKey(msg)
	default:
		return p.handleBrowseKey(msg.String())
	}
}

func (p *Bluetooth) handleBrowseKey(key string) (bool, tea.Cmd) {
	switch key {
	case "up", "k":
		p.list.Prev()
	case "down", "j":
		p.list.Next()
	case "/":
		p.filtering = true
	case "t":
		return true, p.toggleAdapter()
	case "s":
		if p.scanning || !p.adapter.Enabled() {
			return true, nil
		}
		p.scanning, p.lastScan = true, -1
		return true, p.scan()
	case "enter":
		if d, ok := p.selected(); ok {
			return true, p.advance(d)
		}
	case "u":
		if d, ok := p.selected(); ok && d.State().IsPaired() {
			svc, addr := p.svc, d.Address().String()
			return true, deviceMutate(func(ctx context.Context) error {
				_, err := svc.Unpair(ctx, addr)
				return err
			})
		}
	case "r":
		if d, ok := p.selected(); ok {
			p.openRename(d)
		}
	case "d", "x", "delete":
		if d, ok := p.selected(); ok {
			p.openRemove(d)
		}
	default:
		return false, nil
	}
	return true, nil
}

func (p *Bluetooth) handleSearchKey(msg tea.KeyPressMsg) tea.Cmd {
	switch msg.String() {
	case "esc":
		p.search = p.search.Clear()
		p.filtering = false
	case "enter", "down", "up":
		p.filtering = false
		return nil
	default:
		input, typed := p.search.TypeKey(msg)
		if !typed {
			return nil
		}
		p.search = input
	}

	return p.reload()
}

// --- render -----------------------------------------------------------------

func (p *Bluetooth) View(t styles.Theme, width, height int) string {
	inner := width - t.Body.Base.GetHorizontalFrameSize()
	if inner <= 0 {
		return ""
	}

	head := []string{
		t.Body.Note.Render(fit(developmentNote, inner)),
		"",
		btui.Adapter(t, p.adapter.Enabled(), inner),
		"",
		ui.TextField(t, ui.TextFieldOpts{
			Label:       "Buscar",
			Value:       p.search.Value(),
			Placeholder: "escribe para filtrar…",
			Focused:     p.filtering,
			Width:       inner,
		}),
		"",
	}

	tail := []string{"", t.Body.Muted.Render(deviceCounter(p.items) + p.scanStatus())}
	if p.failure != "" {
		tail = append(tail, "", t.Body.Danger.Render(fit(p.failure, inner)))
	}
	tail = append(tail, "", fit(p.hint(t), inner))

	listHeight := max(1, height-frameChrome-len(head)-len(tail))

	rows := make([]string, 0, len(head)+listHeight+len(tail))
	rows = append(rows, head...)
	rows = append(rows, p.deviceRows(t, inner, listHeight)...)
	rows = append(rows, tail...)

	return frame(t, width, height, p.title, rows...)
}

// scanStatus informa de la búsqueda en curso o de su resultado.
func (p *Bluetooth) scanStatus() string {
	switch {
	case p.scanning:
		return " · buscando…"
	case p.lastScan > 0:
		return " · " + plural(p.lastScan, "nuevo", "nuevos")
	case p.lastScan == 0:
		return " · ninguno nuevo"
	default:
		return ""
	}
}

func deviceStatus(d bluetooth.Device) string {
	switch d.State() {
	case bluetooth.StateConnected:
		if b := d.Battery(); b.Known() {
			return "conectado · " + strconv.Itoa(b.Level()) + "%"
		}
		return "conectado"
	case bluetooth.StatePaired:
		return "emparejado"
	default:
		return "no emparejado"
	}
}

var kindLabels = map[bluetooth.Kind]string{
	bluetooth.KindHeadphones: "auriculares",
	bluetooth.KindSpeaker:    "altavoz",
	bluetooth.KindMouse:      "ratón",
	bluetooth.KindKeyboard:   "teclado",
	bluetooth.KindPhone:      "teléfono",
}

// kindLabel traduce el identificador del dominio a etiqueta de pantalla.
func kindLabel(k bluetooth.Kind) string {
	if label, ok := kindLabels[k]; ok {
		return label
	}
	return "desconocido"
}

// hint cambia según el estado del seleccionado: ofrecer "conectar" sobre un
// dispositivo sin emparejar solo confunde.
func (p *Bluetooth) hint(t styles.Theme) string {
	switch {
	case p.filtering:
		return ui.Hints(t.Body.Hint,
			ui.Key{Name: icons.Enter, Action: "aplicar"},
			ui.Key{Name: icons.Escape, Action: "limpiar"},
		)
	case p.scanning:
		return t.Body.Muted.Render("buscando dispositivos…")
	case !p.adapter.Enabled():
		return ui.Hints(t.Body.Hint,
			ui.Key{Name: "t", Action: "activar Bluetooth"},
			ui.Key{Name: icons.UpDown, Action: "mover"},
			ui.Key{Name: "r", Action: "renombrar"},
			ui.Key{Name: "d", Action: "eliminar"},
		)
	}

	d, ok := p.selected()
	if !ok {
		return ui.Hints(t.Body.Hint,
			ui.Key{Name: "s", Action: "buscar dispositivos"},
			ui.Key{Name: "t", Action: "desactivar"},
		)
	}

	keys := []ui.Key{{Name: icons.UpDown, Action: "mover"}}
	switch d.State() {
	case bluetooth.StateConnected:
		keys = append(keys,
			ui.Key{Name: icons.Enter, Action: "desconectar"},
			ui.Key{Name: "u", Action: "olvidar"})
	case bluetooth.StatePaired:
		keys = append(keys,
			ui.Key{Name: icons.Enter, Action: "conectar"},
			ui.Key{Name: "u", Action: "olvidar"})
	default:
		keys = append(keys, ui.Key{Name: icons.Enter, Action: "emparejar"})
	}

	return ui.Hints(t.Body.Hint, append(keys,
		ui.Key{Name: "s", Action: "buscar"},
		ui.Key{Name: "r", Action: "renombrar"},
		ui.Key{Name: "d", Action: "eliminar"},
	)...)
}

func deviceCounter(items []bluetooth.Device) string {
	connected := 0
	for _, d := range items {
		if d.State() == bluetooth.StateConnected {
			connected++
		}
	}

	return plural(len(items), "dispositivo", "dispositivos") + " · " +
		plural(connected, "conectado", "conectados")
}

// deviceRows traduce el dominio al modelo de vista y deja que la sección lo
// pinte. Es el único punto donde la pantalla decide qué se enseña de un
// dispositivo.
func (p *Bluetooth) deviceRows(t styles.Theme, width, height int) []string {
	if len(p.items) == 0 {
		msg := "No hay dispositivos conocidos."
		if q := p.search.Trimmed(); q != "" {
			msg = "Sin resultados para «" + q + "»."
		}
		return []string{t.List.Empty.Render(msg)}
	}

	devices := make([]btui.Device, len(p.items))
	for i, d := range p.items {
		devices[i] = btui.Device{
			Name:   d.Name().String(),
			Kind:   kindLabel(d.Kind()),
			Status: deviceStatus(d),
		}
	}

	return btui.Rows(t, devices, p.list, width, height, !p.filtering)
}
