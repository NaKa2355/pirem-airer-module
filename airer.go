package airer

import (
	"encoding/json"

	"github.com/NaKa2355/pirem-airer-module/internal/app/airer/device"
	"github.com/NaKa2355/pirem/pkg/driver_module/v1"
)

type Module struct{}

var _ driver_module.DriverModule = &Module{}

func (m *Module) LoadDevice(jsonConf json.RawMessage) (driver_module.Device, error) {
	return device.NewDevice(jsonConf)
}
