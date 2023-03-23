package exif_metadata

import (
	"log"
	"math"
	"strconv"
)

// ReadableFileSize
// accepts the file size (number of bytes in float64)
// returns string as a readable format {"B","KB", "MB", "GB", "TB"}
func ReadableFileSize(size float64) string {
	var suffixes = [5]string{"B", "KB", "MB", "GB", "TB"}
	base := math.Log(size) / math.Log(1024)
	getSize := Round(math.Pow(1024, base-math.Floor(base)), .5, 2)
	getSuffix := suffixes[int(math.Floor(base))]
	return strconv.FormatFloat(getSize, 'f', -1, 64) + " " + getSuffix
}

func Round(val float64, roundOn float64, places int) (newVal float64) {
	var round float64
	pow := math.Pow(10, float64(places))
	digit := pow * val
	_, div := math.Modf(digit)
	if div >= roundOn {
		round = math.Ceil(digit)
	} else {
		round = math.Floor(digit)
	}
	newVal = round / pow
	return
}

const LOGGING = false

func PrintLog(message string) {
	if LOGGING {
		log.Println(message)
	}
}
