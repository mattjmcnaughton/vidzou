package main

import (
	"github.com/go-logr/logr"
	"time"
)

type ContentGarbageCollector interface {
	DeleteStaleFiles(cutoffTime time.Time) error
}

type RemoteStoreContentGarbageCollector struct {
	remoteStoreClient RemoteStoreClient
	logger            logr.Logger
}

var _ ContentGarbageCollector = (*RemoteStoreContentGarbageCollector)(nil)

func NewRemoteStoreContentGarbageCollector(remoteStoreClient RemoteStoreClient, logger logr.Logger) *RemoteStoreContentGarbageCollector {
	return &RemoteStoreContentGarbageCollector{
		remoteStoreClient: remoteStoreClient,
		logger:            logger,
	}
}

// DeleteStaleFiles garbage collects all stale files we've uploaded. It's
// intended to be run as a go rountine at a regular cadence.
func (r *RemoteStoreContentGarbageCollector) DeleteStaleFiles(cutoffTime time.Time) error {
	r.logger.V(2).Info("Deleting all stale files uploaded before cutoff time", "cutoffTime", cutoffTime)
	remoteFiles, err := r.remoteStoreClient.ListAllUploadedFiles()

	if err != nil {
		return err
	}

	// Perhaps eventually we will want to do a batch delete... however,
	// since the garbage collection process runs in a separate thread from
	// responding to web requests, I'm not super worried about it for now...
	for _, remoteFile := range remoteFiles {
		if remoteFile.LastModified.Before(cutoffTime) {
			r.logger.V(2).Info("Deleting stale file", "filePath", remoteFile.FilePath)
			r.remoteStoreClient.DeleteFile(remoteFile.FilePath)
		}
	}

	return nil
}

// TODO: Potentially decide whether there's benefit/interest in unit testing
// this method? Tbh, I'm not sure how much it would add...
func RunGarbageCollectionForever(gc ContentGarbageCollector, sleepDuration time.Duration, logger logr.Logger) {
	for {
		cutoff := time.Now().Add(-24 * time.Hour)
		err := gc.DeleteStaleFiles(cutoff)
		if err != nil {
			logger.V(1).Info("Error garbage collecting stale files", "error", err)
		}

		logger.V(3).Info("Sleeping before next garbage collection", "sleepDuration", sleepDuration)
		time.Sleep(sleepDuration)
	}

}
