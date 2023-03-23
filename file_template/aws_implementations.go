package file_template

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/bondhansarker/exif_metadata"
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
		exif_metadata.PrintLog(fmt.Sprintf("Couldn't head object %v:%v. Here's why: %v\n", awsRepo.config.S3.BucketName, objectKey, err))
		return nil, err
	}
	return headOutput, nil
}

// DownloadObjectAsBuffer downloads a chunk of original file and returns as bytes data
func (awsRepo *awsS3Implementation) DownloadObjectAsBuffer(context context.Context, objectKey string, fileRange string) (*exif_metadata.FileObject, error) {
	fileObject := exif_metadata.NewFileObject()

	headOutput, err := awsRepo.BasicFileData(context, objectKey)
	if err != nil {
		return nil, err
	}
	fileObject.SetFileSize(headOutput.ContentLength)

	getObjectInput := &s3.GetObjectInput{
		Bucket: aws.String(awsRepo.config.S3.BucketName),
		Key:    aws.String(objectKey),
		Range:  aws.String(fileRange),
	}

	// Download the object from S3 into the buffer
	resp, err := awsRepo.s3Client.GetObject(context, getObjectInput)
	if err != nil {
		exif_metadata.PrintLog(fmt.Sprintf("couldn't get object %v:%v. Here's why: %v\n", awsRepo.config.S3.BucketName, objectKey, err))
		return nil, err
	}
	// read the file
	byteData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		exif_metadata.PrintLog(fmt.Sprintf("couldn't read object body from %v. Here's why: %v\n", objectKey, err))
		return nil, err
	}
	fileObject.SetByteData(byteData)
	defer resp.Body.Close()
	return fileObject, nil
}

// DownloadObjectAsFile gets an object from a bucket and stores it in a local file.
func (awsRepo *awsS3Implementation) DownloadObjectAsFile(context context.Context, objectKey string, fileRange string, fileName string) (*exif_metadata.FileObject, error) {
	fileObject, err := awsRepo.DownloadObjectAsBuffer(context, objectKey, fileRange)
	file, err := os.Create(fileName)
	if err != nil {
		exif_metadata.PrintLog(fmt.Sprintf("couldn't create file %v. here's why: %v\n", fileName, err))
		return nil, err
	}
	fileObject.SetFilePath(fileName)
	defer file.Close()

	_, err = file.Write(fileObject.FileDataAsByte())
	if err != nil {
		exif_metadata.PrintLog(fmt.Sprintf("couldn't write file %v. here's why: %v\n", fileName, err))
		return nil, err
	}
	return fileObject, nil
}
