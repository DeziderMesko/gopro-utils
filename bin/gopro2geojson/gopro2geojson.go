// made by https://github.com/gki/gopro-utils/tree/feature/add_geojson_export

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	geojson "github.com/paulmach/go.geojson"
	"github.com/JuanIrache/gopro-utils/telemetry"
)

func main() {
	inName := flag.String("i", "", "Required: telemetry file to read")
	outName := flag.String("o", "", "Required: geo json file to write")
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

	var coordinates [][]float64
	var absoluteUtcMicroSec []int64
	var relativeMicroSec []int64

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
				if telems[i].GpsAccuracy == 0 || telems[i].GpsAccuracy == 9999 {
					// Invalid GPS accuracy
					continue
				}

				// longitude, latitude
				longLat := []float64{telems[i].Longitude, telems[i].Latitude}
				coordinates = append(coordinates, longLat)
				// timestamp
				var relGap int64 = 0
				if len(absoluteUtcMicroSec) > 0 {
					relGap = telems[i].TS - absoluteUtcMicroSec[0]
				}
				relativeMicroSec = append(relativeMicroSec, relGap)
				absoluteUtcMicroSec = append(absoluteUtcMicroSec, telems[i].TS)
			}
		}

		*t_prev = *t
		t = &telemetry.TELEM{}
	}

	jsonFile, err := os.Create(*outName)
	if err != nil {
		fmt.Printf("Cannot make output file %s.\n", *outName)
		os.Exit(1)
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Printf("Cannot close json file %s: %s", file.Name(), err)
			os.Exit(1)
		}
	}(jsonFile)

	g := geojson.NewLineStringFeature(coordinates)
	g.SetProperty("AbsoluteUtcMicroSec", absoluteUtcMicroSec)
	g.SetProperty("RelativeMicroSec", relativeMicroSec)
	if err := json.NewEncoder(jsonFile).Encode(g); err != nil {
		fmt.Println("Error encoding output json", err)
		os.Exit(1)
	}

	fmt.Println("Data count")
	fmt.Printf("Coordinates: %d, AbsoluteUtcMicroSec: %d, RelativeMicroSec: %d\n",
		len(coordinates), len(absoluteUtcMicroSec), len(relativeMicroSec))
}
