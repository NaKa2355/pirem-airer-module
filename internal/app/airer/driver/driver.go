package driver

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"time"

	"periph.io/x/conn/v3/driver/driverreg"
	"periph.io/x/conn/v3/spi"
	"periph.io/x/conn/v3/spi/spireg"
	"periph.io/x/host/v3"
)

type Driver struct {
	buf           *bytes.Buffer
	spiPortCloser spi.PortCloser
	spiConn       spi.Conn
}

func New(devFile string) (*Driver, error) {
	d := &Driver{}
	d.buf = &bytes.Buffer{}
	if _, err := host.Init(); err != nil {
		return d, err
	}
	if _, err := driverreg.Init(); err != nil {
		return d, err
	}

	p, err := spireg.Open(devFile)
	if err != nil {
		return d, err
	}
	d.spiPortCloser = p

	d.spiConn, err = d.spiPortCloser.Connect(SpiFreq, SpiMode, 8)
	if err != nil {
		p.Close()
		return d, err
	}

	return d, nil
}

func (d *Driver) GetVersion() (string, error) {
	d.buf.Reset()
	d.buf.Write([]byte{ComGetVersion, 0x0, 0x0, 0x0})
	err := d.spiConn.Tx(d.buf.Bytes(), d.buf.Bytes())
	if err != nil {
		return "", err
	}
	d.buf.Next(1)
	v := d.buf.Next(3)
	return fmt.Sprintf("%d.%d.%d", v[0], v[1], v[2]), nil
}

func (d *Driver) getRecvDataSize() (uint16, error) {
	var dataSize uint16

	d.buf.Reset()
	d.buf.Write([]byte{ComGetRecvDataSize, 0x0, 0x0})

	if err := d.spiConn.Tx(d.buf.Bytes(), d.buf.Bytes()); err != nil {
		return dataSize, err
	}

	d.buf.Next(1)
	if err := binary.Read(d.buf, binary.LittleEndian, &dataSize); err != nil {
		return dataSize, err
	}
	return dataSize, nil
}

func (d *Driver) receiveReq() error {
	d.buf.Reset()
	d.buf.Write([]byte{ComReceive, 0x0})
	return d.spiConn.Tx(d.buf.Bytes(), d.buf.Bytes())
}

func (d *Driver) getRecvData(dataSize uint16) ([]int16, error) {
	var irData []int16
	d.buf.Reset()
	d.buf.Write([]byte{ComGetRecvData})
	d.buf.Write(make([]byte, int(dataSize)*2))
	if err := d.spiConn.Tx(d.buf.Bytes(), d.buf.Bytes()); err != nil {
		return irData, err
	}

	d.buf.Next(1)
	irData = make([]int16, dataSize)
	if err := binary.Read(d.buf, binary.LittleEndian, irData); err != nil {
		return irData, err
	}
	return irData, nil
}

func (d *Driver) ReceiveIr(ctx context.Context) ([]int16, error) {
	var irData []int16
	if err := d.receiveReq(); err != nil {
		return irData, err
	}

	//wait until becoming busy
	time.Sleep(100 * time.Millisecond)
	t := time.NewTicker(300 * time.Millisecond)
	defer t.Stop()

Wait:
	for {
		select {
		case <-ctx.Done():
			return irData, ctx.Err()
		case <-t.C:
			busy, err := d.IsBusy()
			if err != nil {
				return irData, err
			}

			if !busy {
				t.Stop()
				break Wait
			}
		}
	}

	dataSize, err := d.getRecvDataSize()
	if err != nil {
		return irData, err
	}

	err = d.GetErr()
	if err != nil {
		return irData, err
	}
	return d.getRecvData(dataSize)
}

func (d *Driver) GetStatus() (uint8, error) {
	var status uint8
	d.buf.Reset()
	d.buf.Write([]byte{ComGetStat, 0x0})

	if err := d.spiConn.Tx(d.buf.Bytes(), d.buf.Bytes()); err != nil {
		return status, err
	}

	d.buf.Next(1)
	if err := binary.Read(d.buf, binary.LittleEndian, &status); err != nil {
		return status, err

	}
	return status, nil
}

func (d *Driver) GetBufSize() (uint16, error) {
	var bufSize uint16
	d.buf.Reset()
	d.buf.Write([]byte{ComGetBufSize, 0x0, 0x0})

	if err := d.spiConn.Tx(d.buf.Bytes(), d.buf.Bytes()); err != nil {
		return bufSize, err
	}

	d.buf.Next(1)
	if err := binary.Read(d.buf, binary.LittleEndian, &bufSize); err != nil {
		return bufSize, err
	}

	return bufSize, nil
}

func (d *Driver) IsBusy() (bool, error) {
	bufSize, err := d.GetStatus()
	if err != nil {
		return false, err
	}
	return bufSize != 0xf0, nil
}

func (d *Driver) SendIr(ctx context.Context, irData []int16) error {
	d.buf.Reset()
	d.buf.Write([]byte{ComSend})
	if err := binary.Write(d.buf, binary.LittleEndian, irData); err != nil {
		return err
	}
	d.buf.Write([]byte{0x0})
	if err := d.spiConn.Tx(d.buf.Bytes(), d.buf.Bytes()); err != nil {
		return err
	}
	//wait until becoming busy
	time.Sleep(100 * time.Millisecond)
	t := time.NewTicker(300 * time.Millisecond)
	defer t.Stop()

Wait:
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-t.C:
			busy, err := d.IsBusy()
			if err != nil {
				return err
			}

			if !busy {
				t.Stop()
				break Wait
			}
		}
	}

	return d.GetErr()
}

func (d *Driver) GetErr() error {
	var errNum int8
	d.buf.Reset()
	d.buf.Write([]byte{ComGetErr, 0x0, 0x0})

	if err := d.spiConn.Tx(d.buf.Bytes(), d.buf.Bytes()); err != nil {
		return err
	}

	d.buf.Next(1)
	if err := binary.Read(d.buf, binary.LittleEndian, &errNum); err != nil {
		return err
	}
	switch errNum {
	case ERR_DATA_TOO_LONG:
		return ErrDataTooLong
	case ERR_REQ_TIMEOUT:
		return ErrReqTimeout
	case ERR_INVAILD_DATA:
		return ErrInvaildData
	case ERR_UNSUPPORTED_DATA:
		return ErrUnsupportedData
	default:
		return nil
	}
}

func (d *Driver) Close() error {
	return d.spiPortCloser.Close()
}
