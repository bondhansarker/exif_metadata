package file_template

import (
	"context"
	"io/ioutil"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type awsS3Implementation struct {
	config   *Config
	s3Client *s3.Client
}

// NewBucketImplementation initializes the aws config, client awsS3Implementation struct and returns an interface
func NewBucketImplementation(config *Config) IBucketImplementation {
	// creates a new client from the provided config
	s3Client := s3.NewFromConfig(aws.Config{
		Region: config.S3.Region,
		// Credentials: credentials.NewStaticCredentialsProvider(config.AccessKeyID, config.SecretAccessKey, ""),
	})

	return &awsS3Implementation{
		config:   config,
		s3Client: s3Client,
	}
}

func (awsRepo *awsS3Implementation) BasicFileData(context context.Context, objectKey string) (*s3.HeadObjectOutput, error) {
	headInput := &s3.HeadObjectInput{
		Bucket: aws.String(awsRepo.config.S3.BucketName),
		Key:    aws.String(objectKey),
	}

	headOutput, err := awsRepo.s3Client.HeadObject(context, headInput)
	if err != nil {
		log.Printf("Couldn't head object %v:%v. Here's why: %v\n", awsRepo.config.S3.BucketName, objectKey, err)
		return nil, err
	}
	return headOutput, nil
}

// DownloadObjectAsBuffer downloads a chunk of original file and returns as bytes data
func (awsRepo *awsS3Implementation) DownloadObjectAsBuffer(context context.Context, objectKey string, fileRange string) (*FileObject, error) {
	fileObject := NewFileObject()

	headOutput, err := awsRepo.BasicFileData(context, objectKey)
	if err != nil {
		return nil, err
	}
	fileObject.fileSize = headOutput.ContentLength

	getObjectInput := &s3.GetObjectInput{
		Bucket: aws.String(awsRepo.config.S3.BucketName),
		Key:    aws.String(objectKey),
		Range:  aws.String(fileRange),
	}

	// Download the object from S3 into the buffer
	resp, err := awsRepo.s3Client.GetObject(context, getObjectInput)
	if err != nil {
		log.Printf("Couldn't get object %v:%v. Here's why: %v\n", awsRepo.config.S3.BucketName, objectKey, err)
		return nil, err
	}
	// read the file
	fileObject.byteData, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Couldn't read object body from %v. Here's why: %v\n", objectKey, err)
		return nil, err
	}
	defer resp.Body.Close()
	return fileObject, nil
}

// DownloadObjectAsFile gets an object from a bucket and stores it in a local file.
func (awsRepo *awsS3Implementation) DownloadObjectAsFile(context context.Context, objectKey string, fileRange string, fileName string) (*FileObject, error) {
	fileObject, err := awsRepo.DownloadObjectAsBuffer(context, objectKey, fileRange)
	file, err := os.Create(fileName)
	if err != nil {
		log.Printf("Couldn't create file %v. Here's why: %v\n", fileName, err)
		return nil, err
	}
	fileObject.fileName = fileName
	defer file.Close()

	_, err = file.Write(fileObject.byteData)
	if err != nil {
		log.Printf("Couldn't write file %v. Here's why: %v\n", fileName, err)
		return nil, err
	}
	return fileObject, nil
}

// DownloadLargeObject uses a download manager to download an object from a bucket.
// The download manager gets the data in parts and writes them to a buffer until all the data has been downloaded.
func (awsRepo *awsS3Implementation) DownloadLargeObject(context context.Context, objectKey string, fileRange string) (*FileObject, error) {
	fileObject := NewFileObject()
	var partSize int64 = 10 * 1024 * 1024 // 10 mb
	downloader := manager.NewDownloader(awsRepo.s3Client, func(d *manager.Downloader) {
		d.PartSize = partSize
	})
	buffer := manager.NewWriteAtBuffer([]byte{})
	_, err := downloader.Download(context, buffer, &s3.GetObjectInput{
		Bucket: aws.String(awsRepo.config.S3.BucketName),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		log.Printf("Couldn't download large object from %v:%v. Here's why: %v\n",
			awsRepo.config.S3.BucketName, objectKey, err)
	}
	fileObject.byteData = buffer.Bytes()
	fileObject.fileSize = int64(len(fileObject.byteData))

	return fileObject, err
}
