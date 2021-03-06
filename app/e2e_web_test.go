package main

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/sclevine/agouti"
)

const testServerPort = 8081

const waitBetweenAttemptsSeconds = 2
const maxRefreshAttempts = 30

// TODO: Should probably be used throughout the entire program.
type cleanUpFunc func() error

func TestAppE2EWebIntegration(t *testing.T) {
	markIntegrationTest(t)

	cleanUpFunc, err := runWebServer()
	if err != nil {
		t.Fatalf("Error starting web server: %s", err)
	}
	defer cleanUpFunc()

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

	formTextInput := page.Find("form#downloadForm input[name='url']")
	formSubmitInput := page.Find("form#downloadForm input[value='submit']")

	if err := formTextInput.Fill(youtubeURL); err != nil {
		t.Fatalf("Failed to fill url input of form: %s", err)
	}

	if err := formSubmitInput.Click(); err != nil {
		t.Fatalf("Failed to click submit url of form: %s", err)
	}

	var publicFileURL string
	var attempts int
	for {
		attempts++
		if attempts > maxRefreshAttempts {
			t.Fatalf("Falsely showing wait after max attempts: %v", maxRefreshAttempts)
		}

		page.Refresh()

		publicFileURL, err = page.FindByID("publicDownloadURL").Attribute("href")
		if err == nil {
			break
		}

		if _, err = page.FindByID("waitMessage").Text(); err != nil {
			t.Fatalf("Should show #waitMessage if waiting to finish download")
		}

		time.Sleep(waitBetweenAttemptsSeconds * time.Second)
	}

	if err := driver.Stop(); err != nil {
		t.Fatalf("Failed to close pages and stop WebDriver: %s", err)
	}

	getFileResp, err := http.Get(publicFileURL)
	if err != nil {
		t.Fatalf("Error calling GET on public file url: %s", err)
	}

	if getFileResp.StatusCode != 200 {
		t.Fatalf("Request for public file had non-success status code: %d", getFileResp.StatusCode)
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
	s3Client, err := NewS3Client(s3ConfigOptions, testLogger)
	if err != nil {
		return cleanUpFunc, err
	}

	downloader, err := NewDockerYoutubeDlContentDownloader(fsClient, testLogger)
	if err != nil {
		return cleanUpFunc, err
	}
	go downloader.BestEffortInit()

	uploader := NewRemoteStoreContentUploader(s3Client, testLogger)
	server := NewServer(testServerPort, downloader, uploader, testLogger)

	go func() {
		server.ListenAndServe(func() error {
			return nil
		})
	}()

	return cleanUpFunc, nil
}
