package bluetooth

// Adapter es la radio Bluetooth del equipo. No tiene identidad: solo hay una.
type Adapter struct {
	enabled      bool
	discoverable bool
	pairable     bool
}

// NewAdapter reconstruye el adaptador desde el almacenamiento. El valor cero
// es apagado y no visible, que es el arranque seguro.
func NewAdapter(enabled bool) Adapter {
	return Adapter{enabled: enabled}
}

func (a Adapter) Enabled() bool { return a.enabled }

// Discoverable indica si otros dispositivos ven este equipo.
func (a Adapter) Discoverable() bool { return a.discoverable }

// Pairable indica si este equipo acepta solicitudes de emparejamiento.
func (a Adapter) Pairable() bool { return a.pairable }

// Visible resume lo que un panel de ajustes enseña como un solo interruptor:
// el equipo se anuncia y acepta que le pidan emparejarse.
func (a Adapter) Visible() bool { return a.discoverable && a.pairable }

// Enable y Disable son idempotentes: un interruptor que falla al pulsarlo dos
// veces no aporta nada, y aquí no hay invariante que proteger.
func (a Adapter) Enable() Adapter {
	a.enabled = true
	return a
}

// Disable apaga la radio y, con ella, la visibilidad: una radio apagada no
// puede anunciarse ni aceptar solicitudes.
func (a Adapter) Disable() Adapter {
	a.enabled = false
	a.discoverable, a.pairable = false, false
	return a
}

// SetDiscoverable y SetPairable no hacen nada con la radio apagada. Es la
// invariante del agregado: no existe un estado "apagado pero visible".
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

// SetVisible conmuta ambos a la vez.
func (a Adapter) SetVisible(v bool) Adapter {
	return a.SetDiscoverable(v).SetPairable(v)
}
