package file_template

type FileObject struct {
	fileName string
	fileSize int64
	byteData []byte
}

func (fileObj *FileObject) SetByteData(byteData []byte) {
	fileObj.byteData = byteData
}

func (fileObj *FileObject) SetFileName(fileName string) {
	fileObj.fileName = fileName
}

func (fileObj *FileObject) SetFileSize(size int64) {
	fileObj.fileSize = size
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
