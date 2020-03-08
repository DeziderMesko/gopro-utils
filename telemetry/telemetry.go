package telemetry

import (
	"time"

	"github.com/paulmach/go.geo"
)

// Represents telemetry data between "DVID" blocks (typically about 1 second)
type TELEM struct {
	Accl        []ACCL
	Gps         []GPS5
	Gyro        []GYRO
	GpsFix      GPSF
	GpsAccuracy GPSP
	GpsTime     GPSU
	Temp        TMPC
}

// the thing we want, json-wise
// GPS data might have a generated timestamp and derived track
type TELEM_OUT struct {
	*GPS5

	GpsAccuracy uint16  `json:"gps_accuracy,omitempty"`
	GpsFix      uint32  `json:"gps_fix,omitempty"`
	Temp        float32 `json:"temp,omitempty"`
	Track       float64 `json:"track,omitempty"`
}

var pp *geo.Point = geo.NewPoint(10, 10)
var last_good_track float64 = 0

// zeroes out the telem struct
func (t *TELEM) Clear() {
	t.Accl = t.Accl[:0]
	t.Gps = t.Gps[:0]
	t.Gyro = t.Gyro[:0]
	t.GpsTime.Time = time.Time{}
}

// determines if the telem has data
func (t *TELEM) IsZero() bool {
	// hack.
	return t.GpsTime.Time.IsZero()
}

// try to populate a timestamp for every GPS row. return delta between prior and current GPSU GPS times
func (t *TELEM) FillTimes(until time.Time) time.Duration {

	// Compute time delta between GPSU timestamps
	diff := until.Sub(t.GpsTime.Time)

	// Populate timestamps for GPS data, assuming linear distribution over the delta period
	length := len(t.Gps)
	offset := diff.Seconds() / float64(length)
	for i, _ := range t.Gps {
		dur := time.Duration(float64(i)*offset*1000) * time.Millisecond
		ts := t.GpsTime.Time.Add(dur)
		t.Gps[i].TS = ts.UnixNano() / 1000
	}

	// Populate timestamps for ACCL data, assuming linear distribution over the delta period
	length = len(t.Accl)
	offset = diff.Seconds() / float64(length)
	for i, _ := range t.Accl {
		dur := time.Duration(float64(i)*offset*1000) * time.Millisecond
		ts := t.GpsTime.Time.Add(dur)
		t.Accl[i].TS = ts.UnixNano() / 1000
	}

	// Populate timestamps for GYRO data, assuming linear distribution over the delta period
	length = len(t.Gyro)
	offset = diff.Seconds() / float64(length)
	for i, _ := range t.Gyro {
		dur := time.Duration(float64(i)*offset*1000) * time.Millisecond
		ts := t.GpsTime.Time.Add(dur)
		t.Gyro[i].TS = ts.UnixNano() / 1000
	}
	return diff
}

func (t *TELEM) ShitJson() []TELEM_OUT {
	var out []TELEM_OUT

	for i, _ := range t.Gps {
		jobj := TELEM_OUT{&t.Gps[i], 0, 0, 0, 0}
		if 0 == i {
			jobj.GpsAccuracy = t.GpsAccuracy.Accuracy
			jobj.GpsFix = t.GpsFix.F
			jobj.Temp = t.Temp.Temp
		}

		p := geo.NewPoint(jobj.GPS5.Longitude, jobj.GPS5.Latitude)
		jobj.Track = pp.BearingTo(p)
		pp = p

		if jobj.Track < 0 {
			jobj.Track = 360 + jobj.Track
		}

		// only set the track if speed is over 1 m/s
		// if it's slower (eg, stopped) it will drift all over with the location
		if jobj.GPS5.Speed > 1 {
			last_good_track = jobj.Track
		} else {
			jobj.Track = last_good_track
		}

		out = append(out, jobj)
	}

	return out
}
