package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"github.com/mpr90/gopro-utils/telemetry"

	//////used for csv
	"strconv"
    "log"
	"encoding/csv"
	"strings"

)

func main() {

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
		acclCsv = [][]string{{"Milliseconds","AcclX","AcclY","AcclZ"}}
		acclFile, err := os.Create(nameOut[:len(nameOut)-4]+"-accl.csv")
		checkError("Cannot create accl.csv file", err)
		defer acclFile.Close()
		acclWriter = csv.NewWriter(acclFile)
	}
	
	/////////////////////gyroscope
	if strings.Contains(selected, "y") {
		gyroCsv = [][]string{{"Milliseconds","GyroX","GyroY","GyroZ"}}
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

		t_prev.FillTimes(t.GpsTime.Time)

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
				milliseconds := float64(seconds*1000)+float64(((float64(1000)/float64(len(t_prev.Accl)))*float64(i)))
				acclCsv = append(acclCsv, []string{floattostr(milliseconds),floattostr(t_prev.Accl[i].X),floattostr(t_prev.Accl[i].Y),floattostr(t_prev.Accl[i].Z)})
			}
		}
		/////////////////////Gyroscope
		if strings.Contains(selected, "y") {
			for i, _ := range t_prev.Gyro {
				milliseconds := float64(seconds*1000)+float64(((float64(1000)/float64(len(t_prev.Gyro)))*float64(i)))
				gyroCsv = append(gyroCsv, []string{floattostr(milliseconds),floattostr(t_prev.Gyro[i].X),floattostr(t_prev.Gyro[i].Y),floattostr(t_prev.Gyro[i].Z)})
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
	///////////////accelerometer
	if strings.Contains(selected, "a") {
		for _, value := range acclCsv {
			err := acclWriter.Write(value)
			checkError("Cannot write to accl.csv file", err)
		}
		defer acclWriter.Flush()
	}
	///////////////gyroscope
	if strings.Contains(selected, "y") {
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

