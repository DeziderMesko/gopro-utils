package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"github.com/mpr90/gopro-utils/telemetry"
	"io"
	"log"
	"math"
	"os"
	//////used for csv
	"strconv"
	"strings"
)

type ACCLGYRO struct {
	X  []float64
	Y  []float64
	Z  []float64
	FilteredX  []float64
	FilteredY  []float64
	FilteredZ  []float64
	Ms []float64
	TS []int64
}

func main() {

	/*

	   FIR filter designed with
	   http://t-filter.appspot.com

	   sampling frequency: 200 Hz

	   * 0 Hz - 5 Hz
	     gain = 1
	     desired ripple = 5 dB
	     actual ripple = 3.7428345110667873 dB

	   * 10 Hz - 100 Hz
	     gain = 0
	     desired attenuation = -40 dB
	     actual attenuation = -40.91898560530998 dB

	*/

	filter_taps := []float64{
		-0.005946901269245115,
		-0.003221105934752317,
		-0.003844069990276836,
		-0.004335185893314174,
		-0.004618082667223955,
		-0.004610544806595842,
		-0.004229573289430783,
		-0.003391745396869743,
		-0.0020280510416206413,
		-0.00007797378693000673,
		0.002492691628631226,
		0.00569982765877049,
		0.009534812572960348,
		0.013961833701309632,
		0.018917781140524174,
		0.02430952185560112,
		0.030016610101736976,
		0.03589781173597167,
		0.04180040186066689,
		0.04755534497978652,
		0.052984157628638756,
		0.057920150927138224,
		0.062213967331716837,
		0.06570516590449127,
		0.06829411291109368,
		0.069881349050518,
		0.07041715978547827,
		0.069881349050518,
		0.06829411291109368,
		0.06570516590449127,
		0.062213967331716837,
		0.057920150927138224,
		0.052984157628638756,
		0.04755534497978652,
		0.04180040186066689,
		0.03589781173597167,
		0.030016610101736976,
		0.02430952185560112,
		0.018917781140524174,
		0.013961833701309632,
		0.009534812572960348,
		0.00569982765877049,
		0.002492691628631226,
		-0.00007797378693000673,
		-0.0020280510416206413,
		-0.003391745396869743,
		-0.004229573289430783,
		-0.004610544806595842,
		-0.004618082667223955,
		-0.004335185893314174,
		-0.003844069990276836,
		-0.003221105934752317,
		-0.005946901269245115,
	}

	inName := flag.String("i", "", "Required: telemetry file to read")
	outName := flag.String("o", "", "Output csv files")
	userSelect := flag.String("s", "", "Select sensors to output a accelerometer, g gps, y gyroscope, t temperature")
	flag.Parse()

	if *inName == "" {
		flag.Usage()
		return
	}

	///////////////////////////////////////////////////////////////////////////////////////////csv
	nameOut := string(*inName)
	if *outName != "" {
		nameOut = string(*outName)
	}
	selected := string(*userSelect)
	if *userSelect == "" {
		selected = "agyt"
	}

	////////////////////variables for CSV
	var acclCsv, gyroCsv, tempCsv, gpsCsv [][]string
	var acclWriter, gyroWriter, tempWriter, gpsWriter *csv.Writer

	////////////////////accelerometer
	
	if strings.Contains(selected, "a") {
		acclCsv = [][]string{{"Milliseconds","AcclX","AcclY","AcclZ","TS","AcclX_Filt","AcclY_Filt","AcclZ_Filt"}}
		acclFile, err := os.Create(nameOut[:len(nameOut)-4]+"-accl.csv")
		checkError("Cannot create accl.csv file", err)
		defer acclFile.Close()
		acclWriter = csv.NewWriter(acclFile)
	}
	
	/////////////////////gyroscope
	if strings.Contains(selected, "y") {
		gyroCsv = [][]string{{"Milliseconds","GyroX","GyroY","GyroZ","TS","GyroX_Filt","GyroY_Filt","GyroZ_Filt"}}
		gyroFile, err := os.Create(nameOut[:len(nameOut)-4]+"-gyro.csv")
		checkError("Cannot create gyro.csv file", err)
		defer gyroFile.Close()
		gyroWriter = csv.NewWriter(gyroFile)
	}
	//////////////////////temperature
	if strings.Contains(selected, "t") {
		tempCsv = [][]string{{"Milliseconds","Temp"}}
		tempFile, err := os.Create(nameOut[:len(nameOut)-4]+"-temp.csv")
		checkError("Cannot create temp.csv file", err)
		defer tempFile.Close()
		tempWriter = csv.NewWriter(tempFile)
	}
	///////////////////////Uncomment for Gps
	if strings.Contains(selected, "g") {
		gpsCsv = [][]string{{"Milliseconds","Latitude","Longitude","Altitude","Speed","Speed3D","TS","GpsAccuracy","GpsFix"}}
		gpsFile, err := os.Create(nameOut[:len(nameOut)-4]+"-gps.csv")
		checkError("Cannot create gps.csv file", err)
		defer gpsFile.Close()
		gpsWriter = csv.NewWriter(gpsFile)
	}
    //////////////////////////////////////////////////////////////////////////////////////////////

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

	// all acceleration data
	var accl ACCLGYRO
	var gyro ACCLGYRO

	var seconds float64 = -2
	var initialMilliseconds float64 = 0
	for {
		t, err = telemetry.Read(telemFile)
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println(err)
			os.Exit(1)
		}
		if t == nil {
			break
		}

		delta := t_prev.FillTimes(t.GpsTime.Time)

		//fmt.Println(t.GpsTime.GpsTime)

		///////////////////////////////////////////////////////////////////Modified to save CSV
		////////////////////Gps
		
		for i, _ := range t_prev.Gps {
			if (initialMilliseconds <= 0) && (t_prev.Gps[i].TS > 0) { initialMilliseconds = float64(t_prev.Gps[i].TS) / 1000 }
			milliseconds := (float64(t_prev.Gps[i].TS) / 1000) - initialMilliseconds
			if i == 0 {	//if GPS time we can use it for other sensors, otherwise keep seconds guess
				seconds = milliseconds/1000
			}
			if strings.Contains(selected, "g") {
				gpsCsv = append(gpsCsv, []string{floattostr(milliseconds),floattostr(t_prev.Gps[i].Latitude),floattostr(t_prev.Gps[i].Longitude),floattostr(t_prev.Gps[i].Altitude),floattostr(t_prev.Gps[i].Speed),floattostr(t_prev.Gps[i].Speed3D),int64tostr(t_prev.Gps[i].TS),strconv.Itoa(int(t_prev.GpsAccuracy.Accuracy)),strconv.Itoa(int(t_prev.GpsFix.F))})
			}
		}
		/////////////////////Accelerometer
		if strings.Contains(selected, "a") {
			for i, _ := range t_prev.Accl {
				milliseconds := float64(seconds*1000) + (float64(delta.Milliseconds())/float64(len(t_prev.Accl)))*float64(i)
				t_prev.Accl[i].Ms  = milliseconds

				accl.X  = append(accl.X,  t_prev.Accl[i].X)
				accl.Y  = append(accl.Y,  t_prev.Accl[i].Y)
				accl.Z  = append(accl.Z,  t_prev.Accl[i].Z)
				accl.TS = append(accl.TS, t_prev.Accl[i].TS)
				accl.Ms = append(accl.Ms, t_prev.Accl[i].Ms)
			}
		}
		/////////////////////Gyroscope
		if strings.Contains(selected, "y") {
			for i, _ := range t_prev.Gyro {
				milliseconds := float64(seconds*1000) + (float64(delta.Milliseconds())/float64(len(t_prev.Gyro)))*float64(i)
				t_prev.Gyro[i].Ms  = milliseconds

				gyro.X  = append(gyro.X,  t_prev.Gyro[i].X)
				gyro.Y  = append(gyro.Y,  t_prev.Gyro[i].Y)
				gyro.Z  = append(gyro.Z,  t_prev.Gyro[i].Z)
				gyro.TS = append(gyro.TS, t_prev.Gyro[i].TS)
				gyro.Ms = append(gyro.Ms, t_prev.Gyro[i].Ms)
			}
		}
		////////////////////Temperature
		if strings.Contains(selected, "t") {
			milliseconds := seconds*1000
			tempCsv = append(tempCsv, []string{floattostr(milliseconds),floattostr(float64(t_prev.Temp.Temp))})
		}
	    //////////////////////////////////////////////////////////////////////////////////
		
		*t_prev = *t
		t = &telemetry.TELEM{}
		seconds++
	}
	/////////////////////////////////////////////////////////////////////////////////////for csv
	// Filter the accel and gyro data
	accl.FilteredX, err = MyConvolve(accl.X, filter_taps)
	accl.FilteredY, err = MyConvolve(accl.Y, filter_taps)
	accl.FilteredZ, err = MyConvolve(accl.Z, filter_taps)

	gyro.FilteredX, err = MyConvolve(gyro.X, filter_taps)
	gyro.FilteredY, err = MyConvolve(gyro.Y, filter_taps)
	gyro.FilteredZ, err = MyConvolve(gyro.Z, filter_taps)

	///////////////accelerometer
	if strings.Contains(selected, "a") {
		for i, _ := range accl.FilteredX {
			acclCsv = append(acclCsv, []string{floattostr(accl.Ms[i]), floattostr(accl.X[i]), floattostr(accl.Y[i]), floattostr(accl.Z[i]), int64tostr(accl.TS[i]), floattostr(accl.FilteredX[i]), floattostr(accl.FilteredY[i]), floattostr(accl.FilteredZ[i])})
		}

		for _, value := range acclCsv {
			err := acclWriter.Write(value)
			checkError("Cannot write to accl.csv file", err)
		}
		defer acclWriter.Flush()
	}
	///////////////gyroscope
	if strings.Contains(selected, "y") {
		for i, _ := range gyro.FilteredX {
			gyroCsv = append(gyroCsv, []string{floattostr(gyro.Ms[i]), floattostr(gyro.X[i]), floattostr(gyro.Y[i]), floattostr(gyro.Z[i]), int64tostr(gyro.TS[i]), floattostr(gyro.FilteredX[i]), floattostr(gyro.FilteredY[i]), floattostr(gyro.FilteredZ[i])})
		}

		for _, value := range gyroCsv {
			err := gyroWriter.Write(value)
			checkError("Cannot write to gyro.csv file", err)
		}
		defer gyroWriter.Flush()
	}
	/////////////temperature
	if strings.Contains(selected, "t") {
		for _, value := range tempCsv {
			err := tempWriter.Write(value)
			checkError("Cannot write to temp.csv file", err)
		}
		defer tempWriter.Flush()
	}
	/////////////Uncomment for Gps
	if strings.Contains(selected, "g") {
		for _, value := range gpsCsv {
			err := gpsWriter.Write(value)
			checkError("Cannot write to gps.csv file", err)
		}
		defer gpsWriter.Flush()
	}
    /////////////////////////////////////////////////////////////////////////////////////
}

func MyConvolve(input, kernels []float64) ([]float64, error) {
	if !(len(input) > len(kernels)) {
		return nil, fmt.Errorf("provided data set is not greater than the filter weights")
	}

	output := make([]float64, len(input))
	start := int(math.Floor(float64(len(kernels))/2))
	end := len(input)-start

	for i := start; i < end; i++ {
		var sum float64

		for j := 0; j < len(kernels); j++ {
			sum += input[i-start+j] * kernels[j]
		}
		output[i] = sum

	}

	return output, nil
}


///////////for csv

func floattostr(input_num float64) string {

        // to convert a float number to a string
    return strconv.FormatFloat(input_num, 'f', -1, 64)
}



func int64tostr(input_num int64) string {

        // to convert a float number to a string
    return strconv.FormatInt(input_num, 10)
}

 func checkError(message string, err error) {
    if err != nil {
        log.Fatal(message, err)
    }
}

