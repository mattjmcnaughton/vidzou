package main

import (
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"
)

// An e2e test of the non-http components of the app.
func TestAppE2ENoHTTPIntegration(t *testing.T) {
	markIntegrationTest(t)

	tmpS3Bucket, err := createTmpS3Bucket()
	if err != nil {
		t.Fatalf("Error creating tmp s3 bucket: %s", err)
	}
	defer deleteTmpS3Bucket(tmpS3Bucket)

	fsClient, err := NewTmpFsClient()
	if err != nil {
		t.Fatalf("Error generating fsClient: %s", err)
	}
	defer fsClient.CleanUp()

	s3ConfigOptions := &s3ConfigurationOptions{
		awsRegion: awsRegion,
		awsBucket: tmpS3Bucket,
	}
	s3Client, err := NewS3Client(s3ConfigOptions, testLogger)
	if err != nil {
		t.Fatalf("Error creating new S3 Client: %s", err)
	}

	downloader, err := NewDockerYoutubeDlContentDownloader(fsClient, testLogger)
	if err != nil {
		t.Fatalf("Error creating content downloader: %s", err)
	}
	go downloader.BestEffortInit()
	uploader := NewRemoteStoreContentUploader(s3Client, testLogger)
	garbageCollector := NewRemoteStoreContentGarbageCollector(s3Client, testLogger)

	downloadOptions := &DownloadOptions{
		audioOnly: true,
	}
	remotePath := youtubeURL
	downloadedFilePath, err := downloader.DownloadContent(remotePath, downloadOptions)
	if err != nil {
		t.Fatalf("Should not have error downloading content: %s", err)
	}

	publicFileURL, err := uploader.UploadContentPublicly(downloadedFilePath)
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

	// In the real application, we will run the garbage collector
	// in a separate go routine, so we simulate doing so here.
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		garbageCollector.DeleteStaleFiles(time.Now())
		wg.Done()
	}()

	// We've found it can sometimes take one or two tries for our delete to
	// be reflected in the s3 UI (despite us using
	// `WaitUntilObjectNotExists`...).
	numAttempts := 10
	waitBetweenAttempts := 5 * time.Second
	err = retryWithTimeout(numAttempts, waitBetweenAttempts, func() error {
		remainingFiles, err := s3Client.ListAllUploadedFiles()
		if err != nil {
			return err
		}

		if len(remainingFiles) != 0 {
			return fmt.Errorf("Found %d files remaining instead of the expected 0.", len(remainingFiles))
		}

		return nil
	})

	if err != nil {
		t.Fatal(err)
	}
}
