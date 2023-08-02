package driver

import (
	"fmt"

	"periph.io/x/conn/v3/physic"
	"periph.io/x/conn/v3/spi"
)

const SpiFreq = physic.Frequency(30 * physic.KiloHertz)
const SpiMode = spi.Mode3

const (
	ComNop byte = iota
	ComGetVersion
	ComGetBufSize
	ComGetRecvDataSize
	ComGetRecvData
	ComSend
	ComReceive
	ComGetErr
	ComGetStat
)

const (
	ERR_INVAILD_DATA     = -1
	ERR_REQ_TIMEOUT      = -2
	ERR_DATA_TOO_LONG    = -3
	ERR_UNSUPPORTED_DATA = -4
)

var ErrInvaildData = fmt.Errorf("invaild data")
var ErrReqTimeout = fmt.Errorf("request time out")
var ErrDataTooLong = fmt.Errorf("data is too long")
var ErrUnsupportedData = fmt.Errorf("data is not supported")
