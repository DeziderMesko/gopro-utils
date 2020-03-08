package telemetry

import (
	"math"
	"encoding/binary"
	"errors"
)

// Accelerometer in m/s for XYZ
type ACCL struct {
	X float64
	Y float64
	Z float64
	TS int64
}

func (accl *ACCL) Parse(bytes []byte, scale *SCAL) error {
	// hero 5/6 have 2 bytes per axis
	if 6 == len(bytes) {
		// Hero 6/7 order (from ORIN) is typically YxZ (where lower case is a sign inversion)
		accl.Y =  float64(int16(binary.BigEndian.Uint16(bytes[0:2]))) / float64(scale.Values[0])
		accl.X = -float64(int16(binary.BigEndian.Uint16(bytes[2:4]))) / float64(scale.Values[0])
		accl.Z =  float64(int16(binary.BigEndian.Uint16(bytes[4:6]))) / float64(scale.Values[0])

		return nil
	}

	// fusion cameras have 4 bytes per axis (float32)
	if 12 == len(bytes) {
		accl.X = float64(math.Float32frombits(binary.BigEndian.Uint32(bytes[0:4]))) / float64(scale.Values[0])
		accl.Y = float64(math.Float32frombits(binary.BigEndian.Uint32(bytes[4:8]))) / float64(scale.Values[0])
		accl.Z = float64(math.Float32frombits(binary.BigEndian.Uint32(bytes[8:12]))) / float64(scale.Values[0])

		return nil
	}

	return errors.New("Invalid length ACCL packet")
}
