package main

import (
	"time"
)

type ContentGarbageCollector interface {
	DeleteStaleFiles(cutoffTime *time.Time) error
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

func (*RemoteStoreContentGarbageCollector) DeleteStaleFiles(cutoffTime *time.Time) error {
	return nil
}
