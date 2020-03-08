//Credit: KonradIT https://github.com/KonradIT/gopro-utils
//and: mlouielu https://github.com/mlouielu/gopro-utils

package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"time"
	"strconv"

	"github.com/JuanIrache/gopro-utils/telemetry"
	"github.com/mlouielu/gpxgo/gpx"
)

func main() {
	gpxData := new(gpx.GPX)

	inName := flag.String("i", "", "Required: telemetry file to read")
	outName := flag.String("o", "", "Required: gpx file to write")
	accuracyThreshold := flag.Int("a", 1000, "Optional: GPS accuracy threshold, defaults to 1000")
	fixThreshold := flag.Int("f", 3, "Optional: GPS fix state. Defaults to 0 (no fix), can be 2 (2D) or 3 (3D)")
	flag.Parse()

	if *inName == "" {
		flag.Usage()
		return
	}

	telemFile, err := os.Open(*inName)
	if err != nil {
		fmt.Printf("Cannot access telemetry file %s.\n", *inName)
		os.Exit(1)
	}
	defer telemFile.Close()

	t := &telemetry.TELEM{}
	t_prev := &telemetry.TELEM{}

	track := new(gpx.GPXTrack)
	track.Name = string(*inName)[:len(string(*inName))-4]
	segment := new(gpx.GPXTrackSegment)

	for {
		t, err = telemetry.Read(telemFile)
		if err != nil && err != io.EOF {
			fmt.Println("Error reading telemetry file", err)
			os.Exit(1)
		} else if err == io.EOF || t == nil {
			break
		}

		// first full, guess it's about a second
		if t_prev.IsZero() {
			*t_prev = *t
			t.Clear()
			continue
		}

		// process until t.GpsTime
		t_prev.FillTimes(t.Time.Time)
		if t_prev.GpsAccuracy.Accuracy < uint16(*accuracyThreshold) && t_prev.GpsFix.F >= uint32(*fixThreshold) {
			telems := t_prev.ShitJson()
			
			for i, _ := range telems {
			segment.AppendPoint(
				&gpx.GPXPoint{
					Point: gpx.Point{
						Latitude:  telems[i].Latitude,
						Longitude: telems[i].Longitude,
						Elevation: *gpx.NewNullableFloat64(telems[i].Altitude),
					},
					Timestamp: time.Unix(telems[i].TS/1000/1000, telems[i].TS%(1000*1000)*1000).UTC(),
					Comment: "GpsAccuracy: " + strconv.Itoa(int(t_prev.GpsAccuracy.Accuracy)) + "; GpsFix: " + strconv.Itoa(int(t_prev.GpsFix.F)),
				},
			)
		}
		}

		*t_prev = *t
		t = &telemetry.TELEM{}
	}

	track.AppendSegment(segment)
	gpxData.AppendTrack(track)

	gpxFile, err := os.Create(*outName)
	if err != nil {
		fmt.Printf("Cannot make output file %s.\n", *outName)
		os.Exit(1)
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Printf("Cannot close gpx file %s: %s", file.Name(), err)
			os.Exit(1)
		}
	}(gpxFile)

	xml, err := gpxData.ToXml(gpx.ToXmlParams{Version: "1.1", Indent: true})
	gpxFile.Write(xml)
}
