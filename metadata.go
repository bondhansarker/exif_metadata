package exif_metadata

import (
	"errors"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/barasher/go-exiftool"
	"github.com/bondhansarker/exif_metadata/file_template"
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
		createTime:       "CreateDate",
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

type StructuredFileMetadata struct {
	Type       string     `json:"type"`
	Extension  string     `json:"extension"`
	Resolution Resolution `json:"resolution"`
	Size       string     `json:"size"`
	Timestamp  int64      `json:"time"`
	Location   Location   `json:"location"`
}

type objectMetaData struct {
	UnstructuredFileMetadata *exiftool.FileMetadata
	StructuredFileMetadata   *StructuredFileMetadata
	metaDataKeys             *MetaDataKeys
}

func FetchMetaData(fileObject *file_template.FileObject) (*exiftool.FileMetadata, *StructuredFileMetadata, error) {
	objMetaData, err := newObjectMetaData(fileObject.FilePath())
	if err != nil {
		log.Fatal(err)
		return nil, nil, err
	}
	objMetaData.setInfoFields(fileObject)

	if err = objMetaData.setResolution(); err != nil {
		log.Fatal(err)
		return nil, nil, err
	}
	if err = objMetaData.setLocation(); err != nil {
		log.Fatal(err)
		return nil, nil, err
	}
	if err = objMetaData.setDateTime(); err != nil {
		log.Fatal(err)
		return nil, nil, err
	}
	unstructuredMetadata, structuredMetadata := objMetaData.UnstructuredFileMetadata, objMetaData.StructuredFileMetadata
	return unstructuredMetadata, structuredMetadata, nil
}

func newObjectMetaData(filePath string) (*objectMetaData, error) {
	objMetaData := new(objectMetaData)
	// Open the video file
	et, err := exiftool.NewExiftool()

	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	defer et.Close()
	metaData := et.ExtractMetadata(filePath)[0]
	if metaData.Err != nil {
		log.Fatal(err)
		return nil, err
	}
	objMetaData.UnstructuredFileMetadata = &metaData
	objMetaData.StructuredFileMetadata = new(StructuredFileMetadata)
	return objMetaData, nil
}

func (objMetaData *objectMetaData) setInfoFields(fileObject *file_template.FileObject) {
	// determine the mime type based on the file
	fileMimeType := mimetype.Detect(fileObject.FileDataAsByte())
	contentType := getMimeType(fileMimeType.String())

	// stores the file size in readable format
	objMetaData.StructuredFileMetadata.Size = fileObject.ReadableFileSize()
	objMetaData.StructuredFileMetadata.Type = contentType
	objMetaData.StructuredFileMetadata.Extension = fileMimeType.Extension()[1:]
	if contentType == Photo {
		objMetaData.metaDataKeys = LoadPhotoKeys()
	} else {
		objMetaData.metaDataKeys = LoadVideoKeys()
	}
}

func (objMetaData *objectMetaData) setResolution() error {
	unstructuredMetadata, structuredMetadata := objMetaData.UnstructuredFileMetadata, objMetaData.StructuredFileMetadata
	width, err := unstructuredMetadata.GetInt(objMetaData.metaDataKeys.width)
	if err != nil {
		log.Printf("Couldn't Fetch Width. Here's why: %v\n", err)
		return err
	}
	height, err := unstructuredMetadata.GetInt(objMetaData.metaDataKeys.height)
	if err != nil {
		log.Printf("Couldn't Fetch Height. Here's why: %v\n", err)
		return err
	}

	// Check the content rotation
	rotation, err := unstructuredMetadata.GetString(objMetaData.metaDataKeys.rotation)
	if err != nil {
		log.Fatal(err)
	}
	if strings.Contains(rotation, "90") {
		width, height = height, width
	}
	structuredMetadata.Resolution.Width = uint(width)
	structuredMetadata.Resolution.Height = uint(height)
	return nil
}

func (objMetaData *objectMetaData) setLocation() error {
	unstructuredMetadata, structuredMetadata := objMetaData.UnstructuredFileMetadata, objMetaData.StructuredFileMetadata
	gpsPositionString, err := unstructuredMetadata.GetString(objMetaData.metaDataKeys.gpsPosition)
	if err != nil {
		log.Printf("Couldn't Fetch Location. Here's why: %v\n", err)
		return err
	}
	location, err := parseGPSLocationString(gpsPositionString)
	if err != nil {
		log.Fatal(err)
		return err
	}
	structuredMetadata.Location.Latitude = location.Latitude
	structuredMetadata.Location.Longitude = location.Longitude
	return nil
}

func (objMetaData *objectMetaData) setDateTime() error {
	unstructuredMetadata, structuredMetadata := objMetaData.UnstructuredFileMetadata, objMetaData.StructuredFileMetadata
	time.Local = time.UTC // fixing the local time as UTC
	creationTimeString, err := unstructuredMetadata.GetString(objMetaData.metaDataKeys.gpsTime)
	layout := objMetaData.metaDataKeys.gpsTimeLayout
	if err != nil {
		layout = objMetaData.metaDataKeys.createTimeLayout
		creationTimeString, err = unstructuredMetadata.GetString(objMetaData.metaDataKeys.createTime)
		if err != nil {
			log.Printf("Couldn't Fetch Time. Here's why: %v\n", err)
			return err
		}
	}
	timeFromPost, err := time.Parse(layout, creationTimeString)
	if err != nil {
		log.Fatal(err)
		return err
	}
	structuredMetadata.Timestamp = timeFromPost.Unix()
	return nil
}

func parseGPSLocationString(UnparsedLocation string) (*Location, error) {
	unparsedLocations := strings.Split(UnparsedLocation, ",")
	latitude, err := convertLocationStringToFloat(unparsedLocations[0])
	if err != nil {
		return nil, err
	}
	longitude, err := convertLocationStringToFloat(unparsedLocations[1])
	if err != nil {
		return nil, err
	}
	if latitude < -90.0 || latitude > 90.0 || longitude < -180.0 || longitude > 180.0 {
		return nil, errors.New("invalid latitude or longitude")
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

// getMimeType returns the file type by splitting the content type
func getMimeType(contentType string) string {
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
