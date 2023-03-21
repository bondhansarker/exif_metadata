package file_template

import "context"

type IBucketImplementation interface {
	DownloadObjectAsBuffer(context context.Context, objectKey string, fileRange string) (*FileObject, error)
	DownloadObjectAsFile(context context.Context, objectKey string, fileRange string, fileName string) (*FileObject, error)
	DownloadLargeObject(context context.Context, objectKey string, fileRange string) (*FileObject, error)
}
