# GoPro Metadata Format Parser + GPMD2CSV

I forked stilldavid's project ( https://github.com/stilldavid/gopro-utils ) to achieve 3 things:

- Export the data in csv format from /bin/gpmd2csv/gpmd2csv.go
- Allow the project to work with GoPro's h5 v2.00 firmware
- Create a tool for easy data extraction. That's the GPMD2CSV folder. You can just drag and drop the GoPro video files on the BATCH file. If you're not used to github, you can download the tool here: https://tailorandwayne.com/gpmd2csv/

Over time, we have added other exporting tools. They all follow the same pattern when used with the extracted metadata .bin:

`gpmd2csv -i GOPR0001.bin -o GOPR0001.csv`

Aditionally to `-i` and `-o`, the gopro2gpx and gopro2kml tools allow for an `-a` accuracy option for filtering out bad GPS locations (default 1000, the lower the more accurate) and `-f` for type of fix (default 3. 0- no fix, 2 - 2D fix, 3 - 3D fix)

`gopro2gpx -i GOPR0001.bin -a 500 -f 2 -o GOPR0001.gpx`

The gpmd2csv instead allows for a `-s` option to select which data to export. It accepts the following:

- a: Accelerometer
- g: GPS
- y: Gyroscope
- t: Camera temperature

For example, in order to export gyroscope and GPS data only we would do

`gpmd2csv -i GOPR0001.bin -s yg`

If `-s` is not specified, it will export all available data.

Here continues Stilldavid's work:
##############################################################################################################

TLDR:

1.

```
ffmpeg -y -i GOPR0001.MP4 -codec copy -map 0:m:handler_name:"	GoPro MET" -f rawvideo GOPR0001.bin
```

Note the gap before GoPro MET should be a TAB, not a space. Also, the handler_name and position changes between camera models and frame rates. There should be a way to target always the right stream.

2. `gopro2json -i GOPR0001.bin -o GOPR0001.json`

3. There is no step 3

---

I spent some time trying to reverse-engineer the GoPro Metadata Format (GPMD or GPMDF) that is stored in GoPro Hero 5 cameras if GPS is enabled. This is what I found.

## Extracting the Metadata File

The metadata stream is stored in the `.mp4` video file itself alongside the video and audio streams. We can use `ffprobe` to find it:

```
[computar][100GOPRO] ➔ ffprobe GOPR0008.MP4
ffprobe version 3.2.4 Copyright (c) 2007-2017 the FFmpeg developers
[SNIP]
    Stream #0:3(eng): Data: none (gpmd / 0x646D7067), 33 kb/s (default)
    Metadata:
      creation_time   : 2016-11-22T23:42:41.000000Z
      handler_name    : 	GoPro MET
[SNIP]
```

We can identify it by the `gpmd` in the tag string - in this case it's id 3. We can then use `ffmpeg` to extract the metadata stream into a binary file for processing:

`ffmpeg -y -i GOPR0001.MP4 -codec copy -map 0:3 -f rawvideo out-0001.bin`

This leaves us with a binary file with the data.

## Data We Get

- ~400 Hz 3-axis gyro readings
- ~200 Hz 3-axis accelerometer readings
- ~18 Hz GPS position (lat/lon/alt/spd)
- 1 Hz GPS timestamps
- 1 Hz GPS accuracy (cm) and fix (2d/3d)
- 1 Hz temperature of camera

---

## The Protocol

Data starts with a label that describes the data following it. Values are all big endian, and floats are IEEE 754. Everything is packed to 4 bytes where applicable, padded with zeroes so it's 32-bit aligned.

- **Labels** - human readable types of proceeding data
- **Type** - single ascii character describing data
- **Size** - how big is the data type
- **Count** - how many values are we going to get
- **Length** = size \* count

Labels include:

- `ACCL` - accelerometer reading x/y/z
- `DEVC` - device
- `DVID` - device ID, possibly hard-coded to 0x1
- `DVNM` - devicde name, string "Camera"
- `EMPT` - empty packet
- `GPS5` - GPS data (lat, lon, alt, speed, 3d speed)
- `GPSF` - GPS fix (none, 2d, 3d)
- `GPSP` - GPS positional accuracy in cm
- `GPSU` - GPS acquired timestamp; potentially different than "camera time"
- `GYRO` - gryroscope reading x/y/z
- `SCAL` - scale factor, a multiplier for subsequent data
- `SIUN` - SI units; strings (m/s², rad/s)
- `STRM` - ¯\\\_(ツ)\_/¯
- `TMPC` - temperature
- `TSMP` - total number of samples
- `UNIT` - alternative units; strings (deg, m, m/s)

Types include:

- `c` - single char
- `L` - unsigned long
- `s` - signed short
- `S` - unsigned short
- `f` - 32 float

For implementation details, see `reader.go` and other corresponding files in `telemetry/`.

## References

https://github.com/stilldavid/gopro-utils

https://github.com/gopro/gpmf-parser

