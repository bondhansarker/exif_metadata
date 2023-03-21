package file_template

type S3Config struct {
	BucketName string
	Region     string
}

type Config struct {
	S3              *S3Config
	AccessKeyID     string
	SecretAccessKey string
}
