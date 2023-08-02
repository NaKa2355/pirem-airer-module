package airer

import (
	"encoding/json"

	"github.com/NaKa2355/pirem-airer-module/internal/app/airer/device"
	"github.com/NaKa2355/pirem/pkg/module/v1"
)

type Module struct{}

func (m *Module) NewDriver(jsonConf json.RawMessage) (module.Driver, error) {
	return device.NewDevice(jsonConf)
}
