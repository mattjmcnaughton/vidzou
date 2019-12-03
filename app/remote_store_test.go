package main

import (
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

const awsRegion = "us-east-1"

func TestS3ClientUploadFilePubliclyIntegration(t *testing.T) {
	markIntegrationTest(t)

	tmpS3Bucket, err := createTmpS3Bucket()
	if err != nil {
		t.Fatalf("Error creating tmp s3 bucket: %s", err)
	}
	defer deleteTmpS3Bucket(tmpS3Bucket)

	useDefaultTempDirectory := ""
	tmpFile, err := ioutil.TempFile(useDefaultTempDirectory, "test.*.txt")
	if err != nil {
		t.Fatalf("Error creating tmp file: %s", err)
	}
	tmpFilePath := tmpFile.Name()
	defer os.Remove(tmpFilePath)

	s3ConfigOptions := &s3ConfigurationOptions{
		awsRegion: awsRegion,
		awsBucket: tmpS3Bucket,
	}

	s3Client, err := NewS3Client(s3ConfigOptions)
	if err != nil {
		t.Fatalf("Error creating new S3 Client: %s", err)
	}

	remoteFileName := path.Base(tmpFilePath)
	publicFileURL, err := s3Client.UploadFilePublicly(tmpFilePath, remoteFileName)
	if err != nil {
		t.Fatalf("Error uploading file publicly: %s", err)
	}

	getFileResp, err := http.Get(publicFileURL)
	if err != nil {
		t.Fatalf("Error calling GET on public file url: %s", err)
	}
	if getFileResp.StatusCode != 200 {
		t.Fatalf("Request for public file had non-success status code: %d", getFileResp.StatusCode)
	}

	// We don't actually check the file contents... I don't feel like doing
	// so is necessary.
}

func createTmpS3Bucket() (string, error) {
	svc, err := rawS3Client(awsRegion)
	if err != nil {
		return "", err
	}

	bucketNameLength := 16
	randomBucketName := generateRandomString(bucketNameLength)

	_, err = svc.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(randomBucketName),
	})
	if err != nil {
		return "", err
	}

	err = svc.WaitUntilBucketExists(&s3.HeadBucketInput{
		Bucket: aws.String(randomBucketName),
	})
	if err != nil {
		return "", err
	}

	return randomBucketName, nil
}

func deleteTmpS3Bucket(tmpS3BucketName string) error {
	svc, err := rawS3Client(awsRegion)
	if err != nil {
		return err
	}

	if err := prepareTmpS3BucketForDeletion(svc, tmpS3BucketName); err != nil {
		return err
	}

	_, err = svc.DeleteBucket(&s3.DeleteBucketInput{
		Bucket: aws.String(tmpS3BucketName),
	})
	if err != nil {
		return err
	}

	// We do not wait for the bucket to actually not appear anymore, because
	// that operation can sometimes take a really long time... since we are
	// using a unique name for each bucket, we aren't super concerned about
	// starting a new test suite before the old bucket has been completely
	// cleaned up.

	return nil
}

func prepareTmpS3BucketForDeletion(svc *s3.S3, tmpS3BucketName string) error {
	return deleteAllObjectsInBucket(svc, tmpS3BucketName)
}

// deleteAllObjectsInBucket exists because s3 requires a bucket be empty before
// we delete it.
func deleteAllObjectsInBucket(svc *s3.S3, tmpS3BucketName string) error {
	deleteIter := s3manager.NewDeleteListIterator(svc, &s3.ListObjectsInput{
		Bucket: aws.String(tmpS3BucketName),
	})

	return s3manager.NewBatchDeleteWithClient(svc).Delete(aws.BackgroundContext(), deleteIter)
}

func TestS3ClientUploadFilePubliclyFailWhenNonExistentLocalFileIntegration(t *testing.T) {
	markIntegrationTest(t)

	tmpS3Bucket, err := createTmpS3Bucket()
	if err != nil {
		t.Fatalf("Error creating tmp s3 bucket: %s", err)
	}
	defer deleteTmpS3Bucket(tmpS3Bucket)

	s3ConfigOptions := &s3ConfigurationOptions{
		awsRegion: awsRegion,
		awsBucket: tmpS3Bucket,
	}

	s3Client, err := NewS3Client(s3ConfigOptions)
	if err != nil {
		t.Fatalf("Error creating new S3 Client: %s", err)
	}

	tmpFilePath := "/tmp/path/to/nonexistent/file.txt"
	remoteFileName := "doesnt-matter.txt"
	_, err = s3Client.UploadFilePublicly(tmpFilePath, remoteFileName)
	if err == nil {
		t.Fatalf("Should not be able to successfully upload non-existent file")
	}
}

func TestS3ClientUploadFilePubliclyFailWhenInvalidRemoteConfigurationIntegration(t *testing.T) {
	markIntegrationTest(t)

	// S3Bucket is nonexistent bc we intentionally don't create it.
	nonExistentS3Bucket := generateRandomString(16)

	useDefaultTempDirectory := ""
	tmpFile, err := ioutil.TempFile(useDefaultTempDirectory, "test.*.txt")
	if err != nil {
		t.Fatalf("Error creating tmp file: %s", err)
	}
	tmpFilePath := tmpFile.Name()
	defer os.Remove(tmpFilePath)

	s3ConfigOptions := &s3ConfigurationOptions{
		awsRegion: awsRegion,
		awsBucket: nonExistentS3Bucket,
	}

	s3Client, err := NewS3Client(s3ConfigOptions)
	if err != nil {
		t.Fatalf("Error creating new S3 Client: %s", err)
	}

	remoteFileName := path.Base(tmpFilePath)
	_, err = s3Client.UploadFilePublicly(tmpFilePath, remoteFileName)
	if err == nil {
		t.Fatalf("Should not be able to upload a file if the s3 bucket doesn't exist")
	}
}

func TestS3ClientListAllUploadedFilesIntegration(t *testing.T) {
	markIntegrationTest(t)

	// TODO: Should we abstract creating an s3 bucket, and uploading a file
	// to it, into a helper function?
	tmpS3Bucket, err := createTmpS3Bucket()
	if err != nil {
		t.Fatalf("Error creating tmp s3 bucket: %s", err)
	}
	defer deleteTmpS3Bucket(tmpS3Bucket)

	useDefaultTempDirectory := ""
	tmpFile, err := ioutil.TempFile(useDefaultTempDirectory, "test.*.txt")
	if err != nil {
		t.Fatalf("Error creating tmp file: %s", err)
	}
	tmpFilePath := tmpFile.Name()
	defer os.Remove(tmpFilePath)

	s3ConfigOptions := &s3ConfigurationOptions{
		awsRegion: awsRegion,
		awsBucket: tmpS3Bucket,
	}

	s3Client, err := NewS3Client(s3ConfigOptions)
	if err != nil {
		t.Fatalf("Error creating new S3 Client: %s", err)
	}

	remoteFileName := path.Base(tmpFilePath)
	_, err = s3Client.UploadFilePublicly(tmpFilePath, remoteFileName)
	if err != nil {
		t.Fatalf("Error uploading file publicly: %s", err)
	}

	remoteFiles, err := s3Client.ListAllUploadedFiles()
	if err != nil {
		t.Fatalf("Error listing all upload files: %s", err)
	}

	numUploadedFiles := 1
	if len(remoteFiles) != numUploadedFiles {
		t.Fatalf("We uploaded %d files but listing all files only returned %d files.", numUploadedFiles, len(remoteFiles))
	}
}

func TestS3ClientListAllUploadFilesFailWhenInvalidRemoteConfigurationIntegration(t *testing.T) {
	markIntegrationTest(t)

	nonExistentS3Bucket := generateRandomString(16)
	s3ConfigOptions := &s3ConfigurationOptions{
		awsRegion: awsRegion,
		awsBucket: nonExistentS3Bucket,
	}

	s3Client, err := NewS3Client(s3ConfigOptions)
	if err != nil {
		t.Fatalf("Error creating new S3 Client: %s", err)
	}

	_, err = s3Client.ListAllUploadedFiles()
	if err == nil {
		t.Fatalf("Should not be able to list uploaded files for non-existent bucket")
	}
}

func TestS3ClientDeleteFileIntegration(t *testing.T) {
	markIntegrationTest(t)

	tmpS3Bucket, err := createTmpS3Bucket()
	if err != nil {
		t.Fatalf("Error creating tmp s3 bucket: %s", err)
	}
	defer deleteTmpS3Bucket(tmpS3Bucket)

	useDefaultTempDirectory := ""
	tmpFile, err := ioutil.TempFile(useDefaultTempDirectory, "test.*.txt")
	if err != nil {
		t.Fatalf("Error creating tmp file: %s", err)
	}
	tmpFilePath := tmpFile.Name()
	defer os.Remove(tmpFilePath)

	s3ConfigOptions := &s3ConfigurationOptions{
		awsRegion: awsRegion,
		awsBucket: tmpS3Bucket,
	}

	s3Client, err := NewS3Client(s3ConfigOptions)
	if err != nil {
		t.Fatalf("Error creating new S3 Client: %s", err)
	}

	remoteFileName := path.Base(tmpFilePath)
	_, err = s3Client.UploadFilePublicly(tmpFilePath, remoteFileName)
	if err != nil {
		t.Fatalf("Error uploading file publicly: %s", err)
	}

	remoteFiles, err := s3Client.ListAllUploadedFiles()
	if err != nil {
		t.Fatalf("Error listing all upload files: %s", err)
	}

	numUploadedFiles := 1
	if len(remoteFiles) != numUploadedFiles {
		t.Fatalf("We uploaded %d files but listing all files only returned %d files.", numUploadedFiles, len(remoteFiles))
	}

	if err = s3Client.DeleteFile(remoteFileName); err != nil {
		t.Fatalf("Error deleting remote file: %s", err)
	}
	remoteFiles, err = s3Client.ListAllUploadedFiles()
	if err != nil {
		t.Fatalf("Error listing all upload files: %s", err)
	}

	if len(remoteFiles) != numUploadedFiles-1 {
		t.Fatalf("We deleted a file, leaving %d files, but seeing %d files returned", numUploadedFiles-1, len(remoteFiles))
	}
}

func TestS3ClientDeleteFileFailWhenInvalidRemoteConfigurationIntegration(t *testing.T) {
	markIntegrationTest(t)

	nonExistentS3Bucket := generateRandomString(16)
	s3ConfigOptions := &s3ConfigurationOptions{
		awsRegion: awsRegion,
		awsBucket: nonExistentS3Bucket,
	}

	s3Client, err := NewS3Client(s3ConfigOptions)
	if err != nil {
		t.Fatalf("Error creating new S3 Client: %s", err)
	}

	// Our usage of a `non-existent-file` should not be cause a `DeleteFile`
	// failure. S3 actually doesn't consider trying to delete a non-existent
	// file a failure... Rather, its the invalid s3 configuration
	// (non-existent bucket) which will cause the error.
	err = s3Client.DeleteFile("some-non-existent-file.txt")
	if err == nil {
		t.Fatalf("Should not be able to delete a file for non-existent bucket")
	}
}
