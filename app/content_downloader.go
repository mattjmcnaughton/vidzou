package main

type ContentDownloader interface {
	DownloadContent(remotePath string, hostLocation string, downloadOptions *DownloadOpts) error
}

type DownloadOpts struct {
	audioOnly bool
}

type ContainerYoutubeDlContentDownloader struct {
	containerClient ContainerClient
	fsClient        FsClient
}

// Where is the most appropriate place for this interface enforcement to live?
var _ ContentDownloader = (*ContainerYoutubeDlContentDownloader)(nil)

func NewDockerYoutubeDlContentDownloader(fsClient FsClient) (*ContainerYoutubeDlContentDownloader, error) {
	dockerClient, err := NewDockerClient()
	if err != nil {
		return nil, err
	}

	return NewContainerYoutubeDlContentDownloader(dockerClient, fsClient), nil
}

func NewContainerYoutubeDlContentDownloader(containerClient ContainerClient, fsClient FsClient) *ContainerYoutubeDlContentDownloader {
	return &ContainerYoutubeDlContentDownloader{
		containerClient: containerClient,
		fsClient:        fsClient,
	}
}

func (c *ContainerYoutubeDlContentDownloader) DownloadContent(remotePath string, hostLocation string, downloadOptions *DownloadOpts) error {
	return nil
}
