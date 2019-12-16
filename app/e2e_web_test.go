package main

import (
	"fmt"
	"testing"

	"github.com/sclevine/agouti"
	"k8s.io/klog/v2/klogr"
)

const testServerPort = 8081

// TODO: Should probably be used throughout the entire program.
type cleanUpFunc func() error

func TestAppE2EHTTPIntegration(t *testing.T) {
	markIntegrationTest(t)

	cleanUpFunc, err := runWebServer()
	if err != nil {
		t.Fatalf("Error starting web server: %s", err)
	}
	defer cleanUpFunc()

	// Could also install `GeckoDriver` and test on firefox.
	driver := agouti.ChromeDriver()
	if err := driver.Start(); err != nil {
		t.Fatalf("Failed to start Selenium: %s", err)
	}

	page, err := driver.NewPage()
	if err != nil {
		t.Fatalf("Failed to open page: %s", err)
	}

	if err := page.Navigate(fmt.Sprintf("http://localhost:%d", testServerPort)); err != nil {
		t.Fatalf("Failed to navigate: %s", err)
	}

	// TODO: Add steps to try and actually download content.

	if err := driver.Stop(); err != nil {
		t.Fatalf("Failed to close pages and stop WebDriver: %s", err)
	}
}

func runWebServer() (cleanUpFunc, error) {
	tmpS3Bucket, err := createTmpS3Bucket()
	if err != nil {
		return nil, err
	}

	fsClient, err := NewTmpFsClient()
	if err != nil {
		return nil, err
	}

	// Already running clean up as part of the defer blocks...
	cleanUpFunc := func() error {
		if err := deleteTmpS3Bucket(tmpS3Bucket); err != nil {
			return err
		}

		return fsClient.CleanUp()
	}

	s3ConfigOptions := &s3ConfigurationOptions{
		awsRegion: awsRegion,
		awsBucket: tmpS3Bucket,
	}
	s3Client, err := NewS3Client(s3ConfigOptions)
	if err != nil {
		return cleanUpFunc, err
	}

	downloader, err := NewDockerYoutubeDlContentDownloader(fsClient)
	if err != nil {
		return cleanUpFunc, err
	}
	uploader := NewRemoteStoreContentUploader(s3Client)

	logger := klogr.New()
	server := NewServer(testServerPort, downloader, uploader, logger)

	go func() {
		server.ListenAndServe(func() error {
			return nil
		})
	}()

	return cleanUpFunc, nil
}
