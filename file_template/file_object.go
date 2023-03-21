package file_template

type FileObject struct {
	fileName string
	fileSize int64
	byteData []byte
}

func NewFileObject() *FileObject {
	return &FileObject{}
}

func (fileObj *FileObject) FilePath() string {
	return fileObj.fileName
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
