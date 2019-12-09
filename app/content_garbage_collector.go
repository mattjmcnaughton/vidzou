package main

import (
	"time"
)

type ContentGarbageCollector interface {
	DeleteStaleFiles(cutoffTime time.Time) error
}

type RemoteStoreContentGarbageCollector struct {
	remoteStoreClient RemoteStoreClient
}

var _ ContentGarbageCollector = (*RemoteStoreContentGarbageCollector)(nil)

func NewS3RemoteStoreContentGarbageCollector(s3ConfigOptions *s3ConfigurationOptions) (*RemoteStoreContentGarbageCollector, error) {
	s3Client, err := NewS3Client(s3ConfigOptions)
	if err != nil {
		return nil, err
	}

	return NewRemoteStoreContentGarbageCollector(s3Client), nil
}

func NewRemoteStoreContentGarbageCollector(remoteStoreClient RemoteStoreClient) *RemoteStoreContentGarbageCollector {
	return &RemoteStoreContentGarbageCollector{
		remoteStoreClient: remoteStoreClient,
	}
}

// DeleteStaleFiles garbage collects all stale files we've uploaded. It's
// intended to be run as a go rountine at a regular cadence.
func (r *RemoteStoreContentGarbageCollector) DeleteStaleFiles(cutoffTime time.Time) error {
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
