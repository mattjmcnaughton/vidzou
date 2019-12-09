package main

import (
	"testing"
)

func TestRemoteStoreContentUploaderUploadContentPublicly(t *testing.T) {
	fakeRemoteStoreClient := NewFakeRemoteStoreClient()

	tmpFsClient, err := NewTmpFsClient()
	if err != nil {
		t.Fatalf("Error creating TmpFsClient: %s", err)
	}

	fakeContentDownloader := NewFakeContentDownloader(tmpFsClient)
	uploader := NewRemoteStoreContentUploader(fakeRemoteStoreClient)

	remotePath := "https://mattjmcnaughton.com/fake-remote-path-does-not-matter.mp3"
	downloadOptions := &DownloadOptions{}

	// Should we give back the full file path or just the file name?
	downloadedFilePath, err := fakeContentDownloader.DownloadContent(remotePath, downloadOptions)
	if err != nil {
		t.Fatalf("Error downloading content using fake content downloader: %s", err)
	}

	allUploadedFiles, _ := fakeRemoteStoreClient.ListAllUploadedFiles()
	if len(allUploadedFiles) != 0 {
		t.Fatalf("Expected 0 uploaded files, but found %d", len(allUploadedFiles))
	}

	_, err = uploader.UploadContentPublicly(downloadedFilePath)
	if err != nil {
		t.Fatalf("Error uploading file publicly: %s", err)
	}

	allUploadedFiles, _ = fakeRemoteStoreClient.ListAllUploadedFiles()
	if len(allUploadedFiles) != 1 {
		t.Fatalf("Expected 1 uploaded files, but found %d", len(allUploadedFiles))
	}
}
