//Credit: KonradIT https://github.com/KonradIT/gopro-utils

package main

import (
	"flag"
	"fmt"
	"github.com/mpr90/gopro-utils/telemetry"//getting rid of some bugs
	"io"
	"os"
	"time"
	"strconv"
)

func main() {
	inName := flag.String("i", "", "Required: telemetry file to read")
	outName := flag.String("o", "", "Output kml map")
	accuracyThreshold := flag.Int("a", 1000, "Optional: GPS accuracy threshold, defaults to 1000")
	fixThreshold := flag.Int("f", 3, "Optional: GPS fix state. Defaults to 0 (no fix), can be 2 (2D) or 3 (3D)")
	flag.Parse()
	if *inName == "" {
		flag.Usage()
		return
	}
	if *outName == "" {
		flag.Usage()
		return
	}
	/*
		<?xml version="1.0" encoding="UTF-8"?>
		<kml xmlns="http://earth.google.com/kml/2.0">
		<Document>
		<Placemark>
		<Point><coordinates>Longitude,Latitude,Altitude</coordinates></Point>
		<TimeStamp>
     		<when>Timestamp</when>
		</TimeStamp>
		</Placemark>

		[LOOP]
		<Placemark>
		<Point><coordinates>LON,LAT,ALT</coordinates></Point>
		<TimeStamp>
     		<when>Timestamp</when>
		</TimeStamp>
		</Placemark>
		[/LOOP]

		</Document>
		</kml>
	*/
	var gpsData = "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<kml xmlns=\"http://earth.google.com/kml/2.0\">\n<Document>\n"
	gpsFile, err := os.Create(*outName)
	gpsFile.WriteString(gpsData)
	defer gpsFile.Close()

	telemFile, err := os.Open(*inName)
	if err != nil {
		fmt.Printf("Cannot access telemetry file %s.\n", *inName)
		os.Exit(1)
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Printf("Cannot close file %s: %s", file.Name(), err)
			os.Exit(1)
		}
	}(telemFile)

	// currently processing sentence
	t := &telemetry.TELEM{}
	t_prev := &telemetry.TELEM{}

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
		t_prev.FillTimes(t.GpsTime.Time)

		// 		<Placemark>
		// 		<Point><coordinates>LON,LAT,ALT</coordinates></Point>
		// 		<TimeStamp>
		// 			<when>Timestamp</when>
		// 		</TimeStamp>
		// 		</Placemark>

		if t_prev.GpsAccuracy.Accuracy < uint16(*accuracyThreshold) && t_prev.GpsFix.F >= uint32(*fixThreshold) {
			telems := t_prev.ShitJson()
			for i, _ := range telems {
				var TempGpsData string
				TempGpsData = "<Placemark>\n<Point><coordinates>" + floattostr(telems[i].Longitude) + "," + floattostr(telems[i].Latitude) + "," + floattostr(telems[i].Altitude) + "</coordinates></Point><TimeStamp><when>" + time.Unix(telems[i].TS/1000/1000, telems[i].TS%(1000*1000)*1000).UTC().Format("2006-01-02T15:04:05.000Z") + "</when></TimeStamp>" + "\n</Placemark>\n"
				gpsFile.WriteString(TempGpsData)
			}
		}

		*t_prev = *t
		t = &telemetry.TELEM{}
	}
	gpsFile.WriteString("</Document>\n</kml>")

}

func floattostr(input_num float64) string {

	// to convert a float number to a string
	return strconv.FormatFloat(input_num, 'f', -1, 64)
}
