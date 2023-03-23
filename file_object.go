package exif_metadata

type FileObject struct {
	filePath string
	fileSize int64
	byteData []byte
}

func NewFileObject() *FileObject {
	return &FileObject{}
}

func (fileObj *FileObject) SetByteData(byteData []byte) {
	fileObj.byteData = byteData
}

func (fileObj *FileObject) SetFilePath(filePath string) {
	fileObj.filePath = filePath
}

func (fileObj *FileObject) SetFileSize(size int64) {
	fileObj.fileSize = size
}

func (fileObj *FileObject) FilePath() string {
	return fileObj.filePath
}

func (fileObj *FileObject) FileSize() int64 {
	return fileObj.fileSize
}

func (fileObj *FileObject) ReadableFileSize() string {
	return ReadableFileSize(float64(fileObj.fileSize))
}

func (fileObj *FileObject) FileDataAsByte() []byte {
	return fileObj.byteData
}
