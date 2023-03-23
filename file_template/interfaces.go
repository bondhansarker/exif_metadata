package file_template

import (
	"context"

	"github.com/bondhansarker/exif_metadata"
)

type IBucketImplementation interface {
	DownloadObjectAsBuffer(context context.Context, objectKey string, fileRange string) (*exif_metadata.FileObject, error)
	DownloadObjectAsFile(context context.Context, objectKey string, fileRange string, fileName string) (*exif_metadata.FileObject, error)
}
