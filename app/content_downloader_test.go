package main

import (
	"os"
	"path/filepath"
	"testing"
)

const youtubeURL = "https://www.youtube.com/watch?v=hLswuIQ5Tjk"
const invalidURL = "https://mattjmcnaughton.com/this-url/is-invalid"

// Can do unit tests without the `Docker` component... is there value? Maybe do
// later...

func TestDockerContainerYoutubeDlContentDownloaderDownloadContentIntegration(t *testing.T) {
	markIntegrationTest(t)

	audioOnlyValues := []bool{true, false}

	for _, audioOnlyValue := range audioOnlyValues {
		fsClient, err := NewTmpFsClient()
		if err != nil {
			t.Fatalf("Error generating fsClient: %s", err)
		}

		contentDownloader, err := NewDockerYoutubeDlContentDownloader(fsClient)
		if err != nil {
			t.Fatalf("Error creating content downloader: %s", err)
		}

		downloadOptions := &DownloadOptions{
			audioOnly: audioOnlyValue,
		}

		remotePath := youtubeURL

		filePath, err := contentDownloader.DownloadContent(remotePath, downloadOptions)
		if err != nil {
			t.Fatalf("Should not have error downloading content: %s", err)
		}

		// We may also want to wrap these calls in the `fsClient`.
		if _, err := os.Stat(filePath); os.IsNotExist(err) || os.IsPermission(err) {
			t.Fatalf("After download, file should exist and be accessible: %s", err)
		}

		if audioOnlyValue {
			if filepath.Ext(filePath) != "."+defaultAudioFormat {
				t.Fatalf("Expected audio only formats to have %s format, but got %s", defaultAudioFormat, filepath.Ext(filePath))
			}
		}

		fsClient.CleanUp()
	}

}

func TestDockerContainerYoutubeDlContentDownloaderDownloadContentFailsWhenInvalidRemotePathIntegration(t *testing.T) {
	markIntegrationTest(t)

	fsClient, err := NewTmpFsClient()
	if err != nil {
		t.Fatalf("Error generating fsClient: %s", err)
	}

	contentDownloader, err := NewDockerYoutubeDlContentDownloader(fsClient)
	if err != nil {
		t.Fatalf("Error creating content downloader: %s", err)
	}

	downloadOptions := &DownloadOptions{
		audioOnly: true,
	}

	remotePath := invalidURL

	_, err = contentDownloader.DownloadContent(remotePath, downloadOptions)
	if err == nil {
		t.Fatalf("Should not be able to download content from invalid url")
	}

	fsClient.CleanUp()
}
