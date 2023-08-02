package device

import (
	"context"
	"encoding/json"

	plugin "github.com/NaKa2355/pirem/pkg/module/v1"
	"github.com/NaKa2355/pirem_airer_module/internal/app/airer/driver"
)

type Device struct {
	d    *driver.Driver
	info *plugin.DeviceInfo
}

type DeviceConfig struct {
	SpiDevFile string `json:"spi_dev_file"`
}

const DriverVersion = "0.1.0"

var _ plugin.Driver = &Device{}

func convertErr(err error) error {
	if err == nil {
		return nil
	}
	switch err {
	case driver.ErrDataTooLong:
		return plugin.WrapErr(plugin.CodeInvaildInput, err)
	case driver.ErrInvaildData:
		return plugin.WrapErr(plugin.CodeInvaildInput, err)
	case driver.ErrReqTimeout:
		return plugin.WrapErr(plugin.CodeTimeout, err)
	case driver.ErrUnsupportedData:
		return plugin.WrapErr(plugin.CodeInvaildInput, err)
	default:
		return plugin.WrapErr(plugin.CodeDevice, err)
	}
}

func (dev *Device) setInfo() error {
	var err error = nil
	dev.info = &plugin.DeviceInfo{}
	firmVersion, err := dev.d.GetVersion()
	if err != nil {
		return err
	}
	dev.info.CanReceive = true
	dev.info.CanSend = true
	dev.info.FirmwareVersion = firmVersion
	dev.info.DriverVersion = DriverVersion
	return nil
}

func NewDevice(jsonConf json.RawMessage) (dev *Device, err error) {
	dev = &Device{}
	conf := DeviceConfig{}
	err = json.Unmarshal(jsonConf, &conf)
	if err != nil {
		return dev, plugin.WrapErr(plugin.CodeInvaildInput, err)
	}

	d, err := driver.New(conf.SpiDevFile)
	if err != nil {
		return dev, err
	}
	dev.d = d
	if err := dev.setInfo(); err != nil {
		return dev, err
	}

	return dev, nil
}

func (dev *Device) GetInfo(ctx context.Context) (*plugin.DeviceInfo, error) {
	return dev.info, nil
}

func (dev *Device) SendIR(ctx context.Context, irData *plugin.IRData) error {
	err := dev.d.SendIr(ctx, convertToDriverIrRawData(irData.PluseNanoSec))
	return convertErr(err)
}

func (dev *Device) ReceiveIR(ctx context.Context) (*plugin.IRData, error) {
	irData := &plugin.IRData{}
	data, err := dev.d.ReceiveIr(ctx)
	if err != nil {
		return irData, convertErr(err)
	}
	irData.CarrierFreqKiloHz = 40
	irData.PluseNanoSec = convertToApiIrRawData(data)
	return irData, nil
}

func (dev *Device) Drop() error {
	return dev.d.Close()
}
