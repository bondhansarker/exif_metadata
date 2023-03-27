package exif_metadata

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/barasher/go-exiftool"
	timezonemapper "github.com/bradfitz/latlong"
	"github.com/gabriel-vasile/mimetype"
)

const (
	Photo = "photo"
	Video = "video"
)

type MetaDataKeys struct {
	width            string
	height           string
	rotation         string
	createTime       string
	createTimeLayout string
	gpsTime          string
	gpsTimeLayout    string
	gpsPosition      string
	gpsLatitude      string
	gpsLongitude     string
}

func LoadPhotoKeys() *MetaDataKeys {
	return &MetaDataKeys{
		width:            "ImageWidth",
		height:           "ImageHeight",
		rotation:         "Orientation",
		createTime:       "DateTimeOriginal",
		createTimeLayout: "2006:01:02 15:04:05",
		gpsTime:          "GPSDateTime",
		gpsTimeLayout:    "2006:01:02 15:04:05Z",
		gpsPosition:      "GPSPosition",
		gpsLatitude:      "GPSLatitude",
		gpsLongitude:     "GPSLongitude",
	}
}

func LoadVideoKeys() *MetaDataKeys {
	return &MetaDataKeys{
		width:            "ImageWidth",
		height:           "ImageHeight",
		rotation:         "Rotation",
		createTime:       "CreateDate",
		createTimeLayout: "2006:01:02 15:04:05",
		gpsTime:          "CreationDate",
		gpsTimeLayout:    "2006:01:02 15:04:05-07:00",
		gpsPosition:      "GPSPosition",
		gpsLatitude:      "GPSLatitude",
		gpsLongitude:     "GPSLongitude",
	}
}

type Resolution struct {
	Width  uint `json:"width"`
	Height uint `json:"height"`
}

type Location struct {
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"lon"`
}

type DateTime struct {
	Timestamp        time.Time `json:"timestamp"`
	ContainsTimeZone bool      `json:"contains_time_zone"`
}

type ContentInfo struct {
	Type      string `json:"type"`
	Extension string `json:"extension"`
	Size      string `json:"size"`
}

type objectMetaData struct {
	UnstructuredFileMetadata *exiftool.FileMetadata
	ContentInfo              ContentInfo
	metaDataKeys             *MetaDataKeys
}

func FetchMetaData(fileObject *FileObject) (*objectMetaData, error) {
	objMetaData, err := newObjectMetaData(fileObject.FilePath())
	if err != nil {
		fmt.Sprintf("Error for initializing object. The error: %v", err)
		return nil, err
	}
	objMetaData.setInfoFields(fileObject)
	if objMetaData.ContentInfo.Type == Photo {
		objMetaData.metaDataKeys = LoadPhotoKeys()
	} else {
		objMetaData.metaDataKeys = LoadVideoKeys()
	}
	return objMetaData, nil
}

func newObjectMetaData(filePath string) (*objectMetaData, error) {
	objMetaData := new(objectMetaData)
	// Open the video file
	et, err := exiftool.NewExiftool()

	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer et.Close()
	metaData := et.ExtractMetadata(filePath)[0]
	if metaData.Err != nil {
		log.Println(err)
		return nil, err
	}
	objMetaData.UnstructuredFileMetadata = &metaData
	return objMetaData, nil
}

func (objMetaData *objectMetaData) setInfoFields(fileObject *FileObject) {
	// determine the mime type based on the file
	fileMimeType := mimetype.Detect(fileObject.FileDataAsByte())
	contentType := GetMimeType(fileMimeType.String())

	// stores the file size in readable format
	objMetaData.ContentInfo.Size = fileObject.ReadableFileSize()
	objMetaData.ContentInfo.Type = contentType
	objMetaData.ContentInfo.Extension = fileMimeType.Extension()[1:]

}

func (objMetaData *objectMetaData) Resolution() (*Resolution, error) {
	width, err := objMetaData.UnstructuredFileMetadata.GetInt(objMetaData.metaDataKeys.width)
	if err != nil {
		PrintLog(fmt.Sprintf("couldn't fetch width. here's why: %v\n", err))
		return nil, errors.New("width not found")
	}
	height, err := objMetaData.UnstructuredFileMetadata.GetInt(objMetaData.metaDataKeys.height)
	if err != nil {
		PrintLog(fmt.Sprintf("couldn't fetch height. here's why: %v\n", err))
		return nil, errors.New("width not found")
	}

	// Check the content rotation
	rotation, err := objMetaData.UnstructuredFileMetadata.GetString(objMetaData.metaDataKeys.rotation)
	if err != nil {
		PrintLog(fmt.Sprintf("couldn't fetch rotation. here's why: %v\n", err))
	}
	if strings.Contains(rotation, "90") {
		width, height = height, width
	}
	return &Resolution{
		Width:  uint(width),
		Height: uint(height),
	}, nil
}

func (objMetaData *objectMetaData) Location() (*Location, error) {
	gpsPositionString, err := objMetaData.UnstructuredFileMetadata.GetString(objMetaData.metaDataKeys.gpsPosition)
	if err != nil {
		PrintLog(fmt.Sprintf("couldn't fetch location. here's why: %v\n", err))
		return nil, errors.New("location not found")
	}
	location, err := parseGPSLocationString(gpsPositionString)
	if err != nil {
		PrintLog(fmt.Sprintf("failed to parse location. here's why: %v\n", err))
		return nil, errors.New("invalid location")
	}
	return location, nil
}

func (objMetaData *objectMetaData) DateTime() (*DateTime, error) {
	// Check GPS time first
	containsTimeZone := false
	timeString, layout, err := objMetaData.fetchGpsTime()
	if err != nil {
		// Check Create time
		timeString, layout, err = objMetaData.fetchCreateTime()
		if err != nil {
			PrintLog(fmt.Sprintf("couldn't fetch time. here's why: %v\n", err))
			return nil, errors.New("time not found")
		}
		if objMetaData.ContentInfo.Type == Video {
			containsTimeZone = true
		}
	} else {
		containsTimeZone = true
	}

	parsedTime, err := time.Parse(layout, timeString)
	if err != nil {
		PrintLog(fmt.Sprintf("couldn't parse time. here's why: %v\n", err))
		return nil, errors.New("failed to parse time")
	}

	return &DateTime{
		Timestamp:        parsedTime,
		ContainsTimeZone: containsTimeZone,
	}, nil
}

func (objMetaData *objectMetaData) fetchGpsTime() (string, string, error) {
	timeString, err := objMetaData.UnstructuredFileMetadata.GetString(objMetaData.metaDataKeys.gpsTime)
	if err != nil {
		PrintLog(fmt.Sprintf("couldn't fetch gps time. here's why: %v\n", err))
		return "", "", err
	}
	layout := objMetaData.metaDataKeys.gpsTimeLayout
	return timeString, layout, err
}

func (objMetaData *objectMetaData) fetchCreateTime() (string, string, error) {
	timeString, err := objMetaData.UnstructuredFileMetadata.GetString(objMetaData.metaDataKeys.createTime)
	if err != nil {
		PrintLog(fmt.Sprintf("couldn't fetch create time. here's why: %v\n", err))
		return "", "", err
	}
	layout := objMetaData.metaDataKeys.createTimeLayout
	return timeString, layout, err
}

func parseGPSLocationString(UnparsedLocation string) (*Location, error) {
	unparsedLocations := strings.Split(UnparsedLocation, ",")
	latitude, err := convertLocationStringToFloat(unparsedLocations[0])
	if err != nil {
		PrintLog(fmt.Sprintf("failed to parse location string. here's why: %v\n", err))
		return nil, err
	}
	longitude, err := convertLocationStringToFloat(unparsedLocations[1])
	if err != nil {
		PrintLog(fmt.Sprintf("failed to parse location string. here's why: %v\n", err))
		return nil, err
	}

	if ValidateLocation(latitude, longitude) == false {
		return nil, errors.New("invalid location")
	}
	return &Location{
		Latitude:  latitude,
		Longitude: longitude,
	}, nil
}

func convertLocationStringToFloat(location string) (float64, error) {
	location = strings.Replace(location, "deg", "", 1)
	parts := strings.Fields(location)
	degrees, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0, err
	}
	minutes, err := strconv.ParseFloat(strings.TrimSuffix(parts[1], "'"), 64)
	if err != nil {
		return 0, err
	}
	seconds, err := strconv.ParseFloat(strings.TrimSuffix(parts[2], "\""), 64)
	if err != nil {
		return 0, err
	}
	direction := parts[3]
	decimalDegrees := degrees + (minutes / 60) + (seconds / 3600)
	if direction == "S" || direction == "W" {
		decimalDegrees = -decimalDegrees
	}
	return decimalDegrees, nil
}

// GetMimeType returns the file type by splitting the content type
func GetMimeType(contentType string) string {
	pieces := strings.Split(contentType, "/")
	if len(pieces) != 2 {
		return "unknown"
	}
	switch pieces[0] {
	case "image":
		return Photo
	case "video":
		return Video
	}
	return "unknown"
}

func SetTimeBasedOnTimezone(location Location, timeWithoutZone time.Time) (*time.Time, error) {
	// Get the timezone from location
	timezone := timezonemapper.LookupZoneName(location.Latitude, location.Longitude)
	zoneFromLocation, err := time.LoadLocation(timezone)
	if err != nil {
		PrintLog(fmt.Sprintf("failed to load timezone from location. here's why: %v\n", err))
		return &timeWithoutZone, errors.New("failed to load timezone from location")
	}
	timeWithZone := timeWithoutZone
	if zoneFromLocation != nil {
		// get offset/difference with utc in seconds based on the location timezone
		_, differenceWithPostLocation := timeWithoutZone.In(zoneFromLocation).Zone()
		// adding/reducing the difference with the unix time to make it pure unix
		timeWithZone = timeWithoutZone.Add(time.Second * time.Duration(-differenceWithPostLocation))
	}
	return &timeWithZone, nil
}

func ValidateLocation(latitude, longitude float64) bool {
	return latitude >= -90.0 && latitude <= 90.0 && longitude >= -180.0 && longitude <= 180.0
}
