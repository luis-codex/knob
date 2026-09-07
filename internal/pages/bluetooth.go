package pages

import (
	"context"
	"strconv"

	tea "charm.land/bubbletea/v2"

	"knob/internal/application/devices"
	"knob/internal/domain/bluetooth"
	"knob/internal/domain/errs"
	"knob/internal/shared/components"
	"knob/internal/shared/icons"
	"knob/internal/shared/layouts"
	"knob/internal/shared/styles"
	"knob/internal/shared/ui"
	btui "knob/internal/ui/bluetooth"
)

// developmentNote warns about what this screen does not do yet. It goes away
// when incoming pairing is implemented.
const developmentNote = "In development · incoming pairing is missing"

// devicePending remembers what to apply when the modal is accepted.
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

// Bluetooth manages the paired devices.
type Bluetooth struct {
	title string
	uc    devices.UseCases

	items   []bluetooth.Device
	adapter bluetooth.Adapter
	list    *components.List

	search    components.TextInput
	filtering bool

	dialog      components.Dialog
	pending     devicePending
	pendingAddr string

	// scanning blocks the action while the search is in progress; the command
	// runs in its goroutine and the interface keeps responding.
	scanning bool
	// lastScan is how many devices the last search added. -1 is "not searched
	// yet", which is not the same as having found zero.
	lastScan int
	failure  string
}

func NewBluetooth(title string, uc devices.UseCases) *Bluetooth {
	return &Bluetooth{title: title, uc: uc, list: components.NewList(), lastScan: -1}
}

func (p *Bluetooth) Init() tea.Cmd { return p.reload() }

// --- commands -------------------------------------------------------------

func (p *Bluetooth) reload() tea.Cmd {
	uc, query := p.uc, p.search.Trimmed()

	return func() tea.Msg {
		ctx := context.Background()

		adapterRes, err := uc.GetAdapter.Execute(ctx, devices.GetAdapterCommand{})
		if err != nil {
			return devicesFailedMsg{err: err}
		}

		devicesRes, err := uc.SearchDevices.Execute(ctx, devices.SearchDevicesCommand{Query: query})
		if err != nil {
			return devicesFailedMsg{err: err}
		}
		return devicesLoadedMsg{items: devicesRes.Devices, adapter: adapterRes.Adapter}
	}
}

// scan looks for nearby devices.
func (p *Bluetooth) scan() tea.Cmd {
	uc := p.uc

	return func() tea.Msg {
		found, err := uc.ScanDevices.Execute(context.Background(), devices.ScanDevicesCommand{})
		if err != nil {
			return devicesFailedMsg{err: err}
		}
		return scanFinishedMsg{added: len(found.Devices)}
	}
}

// toggleAdapter turns the radio on or off depending on its current state.
func (p *Bluetooth) toggleAdapter() tea.Cmd {
	uc, enabled := p.uc, p.adapter.Enabled()

	return deviceMutate(func(ctx context.Context) error {
		var err error
		if enabled {
			_, err = uc.DisableAdapter.Execute(ctx, devices.DisableAdapterCommand{})
		} else {
			_, err = uc.EnableAdapter.Execute(ctx, devices.EnableAdapterCommand{})
		}
		return err
	})
}

// deviceMutate wraps a write operation: on success it returns deviceSavedMsg
// so the page re-reads itself, and on failure, devicesFailedMsg. Each page has
// its own because the success message differs.
func deviceMutate(op func(context.Context) error) tea.Cmd {
	return func() tea.Msg {
		if err := op(context.Background()); err != nil {
			return devicesFailedMsg{err: err}
		}
		return deviceSavedMsg{}
	}
}

// advance runs the natural action for the current state: pair what is
// discovered, connect what is paired and disconnect what is connected.
func (p *Bluetooth) advance(d bluetooth.Device) tea.Cmd {
	uc, addr := p.uc, d.Address().String()

	return deviceMutate(func(ctx context.Context) error {
		var err error
		switch d.State() {
		case bluetooth.StateDiscovered:
			_, err = uc.PairDevice.Execute(ctx, devices.PairDeviceCommand{Address: addr})
		case bluetooth.StatePaired:
			_, err = uc.ConnectDevice.Execute(ctx, devices.ConnectDeviceCommand{Address: addr})
		case bluetooth.StateConnected:
			_, err = uc.DisconnectDevice.Execute(ctx, devices.DisconnectDeviceCommand{Address: addr})
		}
		return err
	})
}

// --- state --------------------------------------------------------------

func (p *Bluetooth) selected() (bluetooth.Device, bool) {
	if cursor := p.list.Cursor(); cursor < len(p.items) {
		return p.items[cursor], true
	}
	return bluetooth.Device{}, false
}

// --- modals -------------------------------------------------------------

func (p *Bluetooth) openRename(d bluetooth.Device) {
	p.dialog = components.NewFormDialog("Rename device", "Name", "What is it called?", "Save").
		WithValue(d.Name().String()).
		Open()
	p.pending, p.pendingAddr = devicePendingRename, d.Address().String()
}

func (p *Bluetooth) openRemove(d bluetooth.Device) {
	p.dialog = components.NewConfirmDialog(
		"Forget device",
		"\""+d.Name().String()+"\" will be forgotten.\nYou will have to pair it again.",
		"Forget",
	).Dangerous().Open()
	p.pending, p.pendingAddr = devicePendingRemove, d.Address().String()
}

// resolve turns the modal's action into a command. The dialog is not closed
// here: it stays open until the operation confirms.
func (p *Bluetooth) resolve(action components.DialogAction) tea.Cmd {
	switch action {
	case components.DialogNone:
		return nil
	case components.DialogDismiss:
		p.closeDialog()
		return nil
	}

	uc, addr, value := p.uc, p.pendingAddr, p.dialog.Value()

	switch p.pending {
	case devicePendingRename:
		return deviceMutate(func(ctx context.Context) error {
			_, err := uc.RenameDevice.Execute(ctx, devices.RenameDeviceCommand{Address: addr, Name: value})
			return err
		})
	case devicePendingRemove:
		return deviceMutate(func(ctx context.Context) error {
			_, err := uc.RemoveDevice.Execute(ctx, devices.RemoveDeviceCommand{Address: addr})
			return err
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

// --- messages -----------------------------------------------------------

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

// fail decides where the error goes. A validation failure belongs to the form
// that caused it; the rest goes to the page body.
func (p *Bluetooth) fail(err error) {
	if p.dialog.IsOpen() && errs.IsInvalid(err) {
		p.dialog = p.dialog.WithError(err.Error())
		return
	}
	p.closeDialog()
	p.failure = err.Error()
}

// --- keyboard ---------------------------------------------------------------

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
			uc, addr := p.uc, d.Address().String()
			return true, deviceMutate(func(ctx context.Context) error {
				_, err := uc.UnpairDevice.Execute(ctx, devices.UnpairDeviceCommand{Address: addr})
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

// --- render -------------------------------------------------------------

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
			Label:       "Search",
			Value:       p.search.Value(),
			Placeholder: "type to filter…",
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

// scanStatus reports the search in progress or its result.
func (p *Bluetooth) scanStatus() string {
	switch {
	case p.scanning:
		return " · searching…"
	case p.lastScan > 0:
		return " · " + plural(p.lastScan, "new", "new")
	case p.lastScan == 0:
		return " · none new"
	default:
		return ""
	}
}

func deviceStatus(d bluetooth.Device) string {
	switch d.State() {
	case bluetooth.StateConnected:
		if b := d.Battery(); b.Known() {
			return "connected · " + strconv.Itoa(b.Level()) + "%"
		}
		return "connected"
	case bluetooth.StatePaired:
		return "paired"
	default:
		return "not paired"
	}
}

var kindLabels = map[bluetooth.Kind]string{
	bluetooth.KindHeadphones: "headphones",
	bluetooth.KindSpeaker:    "speaker",
	bluetooth.KindMouse:      "mouse",
	bluetooth.KindKeyboard:   "keyboard",
	bluetooth.KindPhone:      "phone",
}

// kindLabel turns the domain identifier into a screen label.
func kindLabel(k bluetooth.Kind) string {
	if label, ok := kindLabels[k]; ok {
		return label
	}
	return "unknown"
}

// hint changes with the selected device's state: offering "connect" on an
// unpaired device only confuses.
func (p *Bluetooth) hint(t styles.Theme) string {
	switch {
	case p.filtering:
		return ui.Hints(t.Body.Hint,
			ui.Key{Name: icons.Enter, Action: "apply"},
			ui.Key{Name: icons.Escape, Action: "clear"},
		)
	case p.scanning:
		return t.Body.Muted.Render("searching for devices…")
	case !p.adapter.Enabled():
		return ui.Hints(t.Body.Hint,
			ui.Key{Name: "t", Action: "turn on Bluetooth"},
			ui.Key{Name: icons.UpDown, Action: "move"},
			ui.Key{Name: "r", Action: "rename"},
			ui.Key{Name: "d", Action: "remove"},
		)
	}

	d, ok := p.selected()
	if !ok {
		return ui.Hints(t.Body.Hint,
			ui.Key{Name: "s", Action: "search for devices"},
			ui.Key{Name: "t", Action: "turn off"},
		)
	}

	keys := []ui.Key{{Name: icons.UpDown, Action: "move"}}
	switch d.State() {
	case bluetooth.StateConnected:
		keys = append(keys,
			ui.Key{Name: icons.Enter, Action: "disconnect"},
			ui.Key{Name: "u", Action: "forget"})
	case bluetooth.StatePaired:
		keys = append(keys,
			ui.Key{Name: icons.Enter, Action: "connect"},
			ui.Key{Name: "u", Action: "forget"})
	default:
		keys = append(keys, ui.Key{Name: icons.Enter, Action: "pair"})
	}

	return ui.Hints(t.Body.Hint, append(keys,
		ui.Key{Name: "s", Action: "search"},
		ui.Key{Name: "r", Action: "rename"},
		ui.Key{Name: "d", Action: "remove"},
	)...)
}

func deviceCounter(items []bluetooth.Device) string {
	connected := 0
	for _, d := range items {
		if d.State() == bluetooth.StateConnected {
			connected++
		}
	}

	return plural(len(items), "device", "devices") + " · " +
		plural(connected, "connected", "connected")
}

// deviceRows translates the domain to the view model and lets the section
// render it. It is the only place where the screen decides what is shown of a
// device.
func (p *Bluetooth) deviceRows(t styles.Theme, width, height int) []string {
	if len(p.items) == 0 {
		msg := "No known devices."
		if q := p.search.Trimmed(); q != "" {
			msg = "No results for \"" + q + "\"."
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
