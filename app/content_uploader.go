package main

import (
	"fmt"
	"github.com/go-logr/logr"
	"os"
	"path"
)

type ContentUploader interface {
	// For now, we do not give the user any control over what we name the
	// file remotely.
	UploadContentPublicly(hostLocation string) (string, error)
}

type RemoteStoreContentUploader struct {
	remoteStoreClient RemoteStoreClient
	logger            logr.Logger
}

var _ ContentUploader = (*RemoteStoreContentUploader)(nil)

func NewRemoteStoreContentUploader(remoteStoreClient RemoteStoreClient, logger logr.Logger) *RemoteStoreContentUploader {
	return &RemoteStoreContentUploader{
		remoteStoreClient: remoteStoreClient,
		logger:            logger,
	}
}

func (r *RemoteStoreContentUploader) UploadContentPublicly(hostLocation string) (string, error) {
	r.logger.V(3).Info("Publicly uploading content from local file system", "hostLocation", hostLocation)

	if _, err := os.Stat(hostLocation); os.IsNotExist(err) || os.IsPermission(err) {
		return "", fmt.Errorf("Error accessing file prior to upload: %s", err)
	}

	remoteFileName := path.Base(hostLocation)
	return r.remoteStoreClient.UploadFilePublicly(hostLocation, remoteFileName)
}
