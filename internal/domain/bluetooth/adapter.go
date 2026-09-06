package bluetooth

// Adapter is the machine's Bluetooth radio. It has no identity: there is only
// one.
type Adapter struct {
	enabled      bool
	discoverable bool
	pairable     bool
}

// NewAdapter rebuilds the adapter from storage. The zero value is off and not
// visible, which is the safe startup state.
func NewAdapter(enabled bool) Adapter {
	return Adapter{enabled: enabled}
}

func (a Adapter) Enabled() bool { return a.enabled }

// Discoverable reports whether other devices can see this machine.
func (a Adapter) Discoverable() bool { return a.discoverable }

// Pairable reports whether this machine accepts pairing requests.
func (a Adapter) Pairable() bool { return a.pairable }

// Visible sums up what a settings panel shows as a single switch: the machine
// advertises itself and accepts pairing requests.
func (a Adapter) Visible() bool { return a.discoverable && a.pairable }

// Enable and Disable are idempotent: a switch that fails on the second press
// adds nothing, and there is no invariant to protect here.
func (a Adapter) Enable() Adapter {
	a.enabled = true
	return a
}

// Disable turns off the radio and, with it, visibility: a radio that is off
// cannot advertise itself or accept requests.
func (a Adapter) Disable() Adapter {
	a.enabled = false
	a.discoverable, a.pairable = false, false
	return a
}

// SetDiscoverable and SetPairable do nothing while the radio is off. That is
// the aggregate's invariant: there is no "off but visible" state.
func (a Adapter) SetDiscoverable(v bool) Adapter {
	if !a.enabled {
		return a
	}
	a.discoverable = v
	return a
}

func (a Adapter) SetPairable(v bool) Adapter {
	if !a.enabled {
		return a
	}
	a.pairable = v
	return a
}

// SetVisible toggles both at once.
func (a Adapter) SetVisible(v bool) Adapter {
	return a.SetDiscoverable(v).SetPairable(v)
}
