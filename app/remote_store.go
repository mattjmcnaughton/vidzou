package main

import (
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/go-logr/logr"
)

// Eventually, we want this to be a parameter?
const presignTime = 5 * time.Minute

type RemoteStoreClient interface {
	// We could have two separate steps... one for uploading a file
	// privately and then another for sharing the public link... but I'm not
	// sure that actually buys us anything.
	UploadFilePublicly(hostFilePath, remoteFileName string) (string, error)
	ListAllUploadedFiles() ([]*RemoteFile, error)
	DeleteFile(remoteFileName string) error
}

type RemoteFile struct {
	FilePath     string
	LastModified time.Time
}

type S3Client struct {
	sess          *session.Session
	svc           *s3.S3
	uploader      *s3manager.Uploader
	configOptions *s3ConfigurationOptions
	logger        logr.Logger
}

type s3ConfigurationOptions struct {
	awsRegion string

	// For now, we do NOT make our program responsible for creating the
	// bucket to which it will write. We don't give it this responsibility
	// because we don't want to give our application permission to create
	// buckets.
	awsBucket string
}

type FakeRemoteStoreClient struct {
	remoteFiles []*RemoteFile
}

var _ RemoteStoreClient = (*S3Client)(nil)
var _ RemoteStoreClient = (*FakeRemoteStoreClient)(nil)

// NewS3Client creates a new S3Client which conforms to our RemoteStoreClient
// interface.
func NewS3Client(configOptions *s3ConfigurationOptions, logger logr.Logger) (*S3Client, error) {
	logger.V(3).Info("Establishing new aws session")
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(configOptions.awsRegion),
	})
	if err != nil {
		return nil, err
	}

	logger.V(3).Info("Establishing new s3 svc and uploader")
	svc := s3.New(sess)
	uploader := s3manager.NewUploader(sess)

	s3Client := &S3Client{
		sess:          sess,
		svc:           svc,
		uploader:      uploader,
		configOptions: configOptions,
		logger:        logger,
	}

	return s3Client, nil
}

// UploadFilePublicly uploads our local content into a remote store, and
// returns a publicly accessible link to download. On the S3Client, this
// entails uploading the file to an S3 bucket, and then generating and returning a presigned
// url.
func (s *S3Client) UploadFilePublicly(hostFilePath, remoteFileName string) (string, error) {
	s.logger.V(2).Info("Uploading file publicly", "hostFilePath", hostFilePath, "remoteFileName", remoteFileName)
	if err := s.uploadFile(hostFilePath, remoteFileName); err != nil {
		return "", err
	}

	return s.generatePublicURLForUploadedFile(remoteFileName)
}

func (s *S3Client) uploadFile(hostFilePath, remoteFileName string) error {
	file, err := os.Open(hostFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = s.uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s.configOptions.awsBucket),
		Key:    aws.String(remoteFileName),
		Body:   file,
	})
	return err
}

func (s *S3Client) generatePublicURLForUploadedFile(remoteFileName string) (string, error) {
	s.logger.V(3).Info("Generating public URL for remote file", "remoteFileName", remoteFileName)

	objectRequest, _ := s.svc.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(s.configOptions.awsBucket),
		Key:    aws.String(remoteFileName),
	})

	s.logger.V(3).Info("Presigning URL")
	urlStr, err := objectRequest.Presign(presignTime)
	if err != nil {
		return "", err
	}

	return urlStr, nil
}

func (s *S3Client) ListAllUploadedFiles() ([]*RemoteFile, error) {
	s.logger.V(3).Info("Listing all uploaded files")
	resp, err := s.svc.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: aws.String(s.configOptions.awsBucket),
	})
	if err != nil {
		return nil, err
	}

	remoteFiles := make([]*RemoteFile, len(resp.Contents))
	for i, item := range resp.Contents {
		remoteFile := &RemoteFile{
			FilePath:     *item.Key,
			LastModified: *item.LastModified,
		}

		remoteFiles[i] = remoteFile
	}

	return remoteFiles, nil
}

func (s *S3Client) DeleteFile(remoteFileName string) error {
	s.logger.V(3).Info("Deleting file", "remoteFileName", remoteFileName)

	// Interestingly, `DeleteObject` spec indicates that deleting an object which
	// doesn't exist is not considered an error
	// (https://docs.aws.amazon.com/sdk-for-go/api/service/s3/#S3.DeleteObject).
	_, err := s.svc.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(s.configOptions.awsBucket),
		Key:    aws.String(remoteFileName),
	})

	if err != nil {
		return err
	}

	s.logger.V(3).Info("Waiting for deleted file to not exist", "remoteFileName", remoteFileName)
	return s.svc.WaitUntilObjectNotExists(&s3.HeadObjectInput{
		Bucket: aws.String(s.configOptions.awsBucket),
		Key:    aws.String(remoteFileName),
	})
}

func NewFakeRemoteStoreClient() *FakeRemoteStoreClient {
	return &FakeRemoteStoreClient{
		remoteFiles: []*RemoteFile{},
	}
}

func (f *FakeRemoteStoreClient) UploadFilePublicly(hostFilePath, remoteFileName string) (string, error) {
	fakeFile := &RemoteFile{
		FilePath:     remoteFileName,
		LastModified: time.Now(),
	}

	f.remoteFiles = append(f.remoteFiles, fakeFile)

	return "fake-presigned-url", nil
}

func (f *FakeRemoteStoreClient) ListAllUploadedFiles() ([]*RemoteFile, error) {
	// We need to copy `remoteFiles` into a stable slice so that anyone
	// interacting with the returned slice sees a consistent list of
	// remoteFiles (regardless of whether those files are deleted later,
	// etc...).
	stableRemoteFiles := make([]*RemoteFile, len(f.remoteFiles))
	copy(stableRemoteFiles, f.remoteFiles)

	return stableRemoteFiles, nil
}

func (f *FakeRemoteStoreClient) DeleteFile(remoteFileName string) error {
	fileNotFound := -1
	indiceOfFileToDelete := fileNotFound

	for i, remoteFile := range f.remoteFiles {
		if remoteFile.FilePath == remoteFileName {
			indiceOfFileToDelete = i
		}
	}

	if indiceOfFileToDelete != fileNotFound {
		f.remoteFiles = append(f.remoteFiles[:indiceOfFileToDelete], f.remoteFiles[indiceOfFileToDelete+1:]...)
	}

	// Not considered an error to delete a file which already doesn't exist.
	return nil
}

func (f *FakeRemoteStoreClient) UploadRandomFilesWithMockedAge(lastModifiedToNumFilesToCreate map[time.Time]int) {
	for lastModifiedTime, numFilesToCreate := range lastModifiedToNumFilesToCreate {
		for i := 0; i < numFilesToCreate; i++ {
			remoteFile := &RemoteFile{
				FilePath:     generateRandomString(16),
				LastModified: lastModifiedTime,
			}

			f.remoteFiles = append(f.remoteFiles, remoteFile)
		}
	}
}
