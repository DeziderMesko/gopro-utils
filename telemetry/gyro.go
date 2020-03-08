package telemetry

import (
	"encoding/binary"
	"errors"
)

// 3-axis Gyroscope data in rad/s
type GYRO struct {
	X float64
	Y float64
	Z float64
	TS int64
}

func (gyro *GYRO) Parse(bytes []byte, scale *SCAL) error {
	if 6 != len(bytes) {
		return errors.New("Invalid length GYRO packet")
	}

	// Hero 6/7 order (from ORIN) is typically YxZ (where lower case is a sign inversion)
	gyro.Y =  float64(int16(binary.BigEndian.Uint16(bytes[0:2]))) / float64(scale.Values[0])
	gyro.X = -float64(int16(binary.BigEndian.Uint16(bytes[2:4]))) / float64(scale.Values[0])
	gyro.Z =  float64(int16(binary.BigEndian.Uint16(bytes[4:6]))) / float64(scale.Values[0])

	return nil
}
