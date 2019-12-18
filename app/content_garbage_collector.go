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
			r.remoteStoreClient.DeleteFile(remoteFile.FilePath)
		}
	}

	return nil
}
