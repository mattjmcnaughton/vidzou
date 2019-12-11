package main

import (
	"fmt"
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
}

var _ ContentUploader = (*RemoteStoreContentUploader)(nil)

func NewRemoteStoreContentUploader(remoteStoreClient RemoteStoreClient) *RemoteStoreContentUploader {
	return &RemoteStoreContentUploader{
		remoteStoreClient: remoteStoreClient,
	}
}

func (r *RemoteStoreContentUploader) UploadContentPublicly(hostLocation string) (string, error) {
	if _, err := os.Stat(hostLocation); os.IsNotExist(err) || os.IsPermission(err) {
		return "", fmt.Errorf("Error accessing file prior to upload: %s", err)
	}

	remoteFileName := path.Base(hostLocation)
	return r.remoteStoreClient.UploadFilePublicly(hostLocation, remoteFileName)
}
