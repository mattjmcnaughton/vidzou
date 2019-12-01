package main

type ContentUploader interface {
	// For now, we do not give the user any control over what we name the
	// file remotely.
	UploadContentPublicly(hostLocation string) (string, error)
}

type RemoteStoreContentUploader struct {
	remoteStoreClient RemoteStoreClient
	fsClient          FsClient
}

var _ ContentUploader = (*RemoteStoreContentUploader)(nil)

func NewS3RemoteStoreContentUploader(fsClient FsClient, s3ConfigOptions *s3ConfigurationOptions) (*RemoteStoreContentUploader, error) {
	s3Client, err := NewS3Client(s3ConfigOptions)
	if err != nil {
		return nil, err
	}

	return NewRemoteStoreContentUploader(s3Client, fsClient), nil
}

func NewRemoteStoreContentUploader(remoteStoreClient RemoteStoreClient, fsClient FsClient) *RemoteStoreContentUploader {
	return &RemoteStoreContentUploader{
		remoteStoreClient: remoteStoreClient,
		fsClient:          fsClient,
	}
}

func (r *RemoteStoreContentUploader) UploadContentPublicly(hostLocation string) (string, error) {
	return "", nil
}
