package main

import (
	"testing"
	"time"
)

func TestRemoteStoreContentGarbageCollectDeleteStaleFiles(t *testing.T) {
	fakeRemoteStoreClient := NewFakeRemoteStoreClient()

	numFilesToKeep := 5
	numFilesToDelete := 3

	nonStaleTime := time.Now()
	cutoffTime := nonStaleTime.Add(-1 * time.Hour)
	staleTime := nonStaleTime.Add(-2 * time.Hour)

	lastModifiedToNumFilesToCreate := map[time.Time]int{
		nonStaleTime: numFilesToKeep,
		staleTime:    numFilesToDelete,
	}

	fakeRemoteStoreClient.UploadRandomFilesWithMockedAge(lastModifiedToNumFilesToCreate)

	garbageCollector := NewRemoteStoreContentGarbageCollector(fakeRemoteStoreClient, testLogger)
	garbageCollector.DeleteStaleFiles(cutoffTime)

	stillExistingFiles, _ := fakeRemoteStoreClient.ListAllUploadedFiles()
	if len(stillExistingFiles) != numFilesToKeep {
		t.Fatalf("Expected %d files to remain after garbage collection, but found: %d", numFilesToKeep, len(stillExistingFiles))
	}
}
