package devices

import (
	"context"

	"knob/internal/domain/bluetooth"
)

// ReportBattery records the charge level reported by the device.
type ReportBattery struct{ deps Deps }

func NewReportBattery(deps Deps) ReportBattery { return ReportBattery{deps: deps} }

type ReportBatteryCommand struct {
	Address string
	Level   int
}

type ReportBatteryResponse struct{ Device bluetooth.Device }

func (u ReportBattery) Execute(ctx context.Context, cmd ReportBatteryCommand) (ReportBatteryResponse, error) {
	battery, err := bluetooth.NewBattery(cmd.Level)
	if err != nil {
		return ReportBatteryResponse{}, err
	}

	device, err := u.deps.apply(ctx, cmd.Address, func(d bluetooth.Device) (bluetooth.Device, error) {
		return d.ReportBattery(battery)
	})
	if err != nil {
		return ReportBatteryResponse{}, err
	}
	return ReportBatteryResponse{Device: device}, nil
}
