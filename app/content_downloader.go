package main

import (
	"fmt"
	"io/ioutil"
	"path"
	"strings"
)

const defaultAudioFormat = "mp3"

type ContentDownloader interface {
	DownloadContent(remotePath string, downloadOptions *DownloadOptions) (string, error)
}

type DownloadOptions struct {
	audioOnly bool
}

type ContainerYoutubeDlContentDownloader struct {
	containerClient ContainerClient
	fsClient        FsClient
}

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

func (c *ContainerYoutubeDlContentDownloader) DownloadContent(remotePath string, downloadOptions *DownloadOptions) (string, error) {
	// Image should be specified as constant elsewhere, probably?
	image := "mattjmcnaughton/youtube-dl:0.0.1.a"
	// Mount directory should be specified as contant? Must match the
	// published container.
	youtubeDlMountDirectory := "/downloads"

	// We will use this unique prefix for identifying the file on the file
	// system (as we can't predict the title, extension, etc...)
	uniqueOutputFilePrefix := generateRandomString(8)
	fileNameTemplate := fmt.Sprintf("%s/%s-%%(title)s.%%(ext)s", youtubeDlMountDirectory, uniqueOutputFilePrefix)

	cmd := []string{"-o", fileNameTemplate, remotePath}

	if downloadOptions.audioOnly {
		audioOnlyYoutubeDlOptions := []string{"-x", "--audio-format", defaultAudioFormat}
		cmd = append(audioOnlyYoutubeDlOptions, cmd...)
	}

	binds := []string{
		fmt.Sprintf("%s:%s", c.fsClient.GetMountDirectory(), youtubeDlMountDirectory),
	}

	runContainerOpts := &runContainerOptions{
		binds: binds,
	}

	if err := c.containerClient.EnsureImageAvailableOnHost(image); err != nil {
		return "", err
	}

	if err := c.containerClient.RunContainer(image, cmd, runContainerOpts); err != nil {
		return "", err
	}

	return c.findFileUsingUniqueIdentifier(uniqueOutputFilePrefix)
}

// We will use this unique prefix for identifying the file on the file
// system (as we can't predict the title, extension, etc...)
func (c *ContainerYoutubeDlContentDownloader) findFileUsingUniqueIdentifier(uniqueOutputFilePrefix string) (string, error) {
	filesInDir, err := ioutil.ReadDir(c.fsClient.GetMountDirectory())
	if err != nil {
		return "", err
	}

	for _, f := range filesInDir {
		if strings.HasPrefix(f.Name(), uniqueOutputFilePrefix) {
			return path.Join(c.fsClient.GetMountDirectory(), f.Name()), nil
		}
	}

	return "", fmt.Errorf("Cannot identify file with unique prefix: %s", uniqueOutputFilePrefix)
}
