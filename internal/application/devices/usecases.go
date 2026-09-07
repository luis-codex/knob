package devices

// UseCases bundles every device use case so the composition root wires them in
// one place and a page holds a single field.
type UseCases struct {
	ScanDevices      ScanDevices
	GetAdapter       GetAdapter
	EnableAdapter    EnableAdapter
	DisableAdapter   DisableAdapter
	DiscoverDevice   DiscoverDevice
	PairDevice       PairDevice
	UnpairDevice     UnpairDevice
	ConnectDevice    ConnectDevice
	DisconnectDevice DisconnectDevice
	RenameDevice     RenameDevice
	ReportBattery    ReportBattery
	RemoveDevice     RemoveDevice
	ListDevices      ListDevices
	SearchDevices    SearchDevices
}

// NewUseCases builds every use case from the shared ports.
func NewUseCases(deps Deps) UseCases {
	return UseCases{
		ScanDevices:      NewScanDevices(deps),
		GetAdapter:       NewGetAdapter(deps),
		EnableAdapter:    NewEnableAdapter(deps),
		DisableAdapter:   NewDisableAdapter(deps),
		DiscoverDevice:   NewDiscoverDevice(deps),
		PairDevice:       NewPairDevice(deps),
		UnpairDevice:     NewUnpairDevice(deps),
		ConnectDevice:    NewConnectDevice(deps),
		DisconnectDevice: NewDisconnectDevice(deps),
		RenameDevice:     NewRenameDevice(deps),
		ReportBattery:    NewReportBattery(deps),
		RemoveDevice:     NewRemoveDevice(deps),
		ListDevices:      NewListDevices(deps),
		SearchDevices:    NewSearchDevices(deps),
	}
}
