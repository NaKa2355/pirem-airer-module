package device

import (
	"math"
	"time"
)

func convertToApiIrRawData(irData []int16) []uint32 {
	apiIrData := make([]uint32, len(irData))
	for i, pluse := range irData {
		if pluse < 0 {
			apiIrData[i] = uint32(pluse*-1) * uint32(time.Millisecond)
		} else {
			apiIrData[i] = uint32(pluse) * uint32(time.Microsecond)
		}
	}
	return apiIrData
}

func convertToDriverIrRawData(irData []uint32) []int16 {
	driverIrData := make([]int16, len(irData))
	for i, pluse := range irData {
		pluse = pluse / uint32(time.Microsecond)
		if pluse > math.MaxInt16 {
			driverIrData[i] = int16(pluse/1000) * -1
		} else {
			driverIrData[i] = int16(pluse)
		}
	}
	return driverIrData
}
