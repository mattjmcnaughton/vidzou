package main

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type RemoteStoreClient interface {
	// We could have two separate steps... one for uploading a file
	// privately and then another for sharing the public link... but I'm not
	// sure that actually buys us anything.
	UploadFilePublicly(hostFilePath, remoteFileName string) (string, error)
	ListAllUploadedFiles() ([]*RemoteFile, error)
	DeleteFile(remoteFilePath string) error
}

type RemoteFile struct {
	FilePath     string
	LastModified *time.Time
}

type S3Client struct {
	sess          *session.Session
	svc           *s3.S3
	uploader      *s3manager.Uploader
	configOptions *s3ConfigurationOptions
}

type s3ConfigurationOptions struct {
	awsRegion string
	awsBucket string
}

var _ RemoteStoreClient = (*S3Client)(nil)

func NewS3Client(configOptions *s3ConfigurationOptions) (*S3Client, error) {
	sess, err := session.NewSession(&aws.Config{
		// Stop harcoding
		Region: aws.String("us-east-1"),
	})
	if err != nil {
		return nil, err
	}

	svc := s3.New(sess)
	uploader := s3manager.NewUploader(sess)

	s3Client := &S3Client{
		sess:          sess,
		svc:           svc,
		uploader:      uploader,
		configOptions: configOptions,
	}

	return s3Client, nil
}

func (s *S3Client) UploadFilePublicly(hostFilePath, remoteFileName string) (string, error) {
	return "", nil
}

func (s *S3Client) ListAllUploadedFiles() ([]*RemoteFile, error) {
	return nil, nil
}

func (s *S3Client) DeleteFile(remoteFilePath string) error {
	return nil
}
