package main

import (
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"testing"
)

func TestS3ClientUploadFilePubliclyIntegration(t *testing.T) {
	markIntegrationTest(t)

	tmpS3Bucket, err := createTmpS3Bucket()
	if err != nil {
		t.Fatalf("Error creating tmp s3 bucket: %s", err)
	}
	defer deleteTmpS3Bucket(tmpS3Bucket)

	// Update to use `TmpFsClient`.
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
