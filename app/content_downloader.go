package main

import (
	"fmt"
	"github.com/go-logr/logr"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

const defaultAudioFormat = "mp3"

type ContentDownloader interface {
	DownloadContent(remotePath string, downloadOptions *DownloadOptions) (string, error)

	// BestEffortInit contains non-critical operations which, if run before
	// the first call of `DownloadContent`, improve performance.
	BestEffortInit() error
}

type DownloadOptions struct {
	audioOnly bool
}

type ContainerYoutubeDlContentDownloader struct {
	containerClient    ContainerClient
	fsClient           FsClient
	logger             logr.Logger
	YoutubeDlImageName string
}

type FakeContentDownloader struct {
	fsClient FsClient
}

var _ ContentDownloader = (*ContainerYoutubeDlContentDownloader)(nil)
var _ ContentDownloader = (*FakeContentDownloader)(nil)

func NewDockerYoutubeDlContentDownloader(fsClient FsClient, logger logr.Logger) (*ContainerYoutubeDlContentDownloader, error) {
	dockerClient, err := NewDockerClient(logger)
	if err != nil {
		return nil, err
	}

	return NewContainerYoutubeDlContentDownloader(dockerClient, fsClient, logger), nil
}

func NewContainerYoutubeDlContentDownloader(containerClient ContainerClient, fsClient FsClient, logger logr.Logger) *ContainerYoutubeDlContentDownloader {
	return &ContainerYoutubeDlContentDownloader{
		containerClient: containerClient,
		fsClient:        fsClient,
		logger:          logger,

		// Can set via constructor/setting later, should we find the need.
		YoutubeDlImageName: "mattjmcnaughton/youtube-dl:2020.03.24",
	}
}

func (c *ContainerYoutubeDlContentDownloader) DownloadContent(remotePath string, downloadOptions *DownloadOptions) (string, error) {
	c.logger.V(3).Info("Downloading content using ContainerYoutubeDl", "remotePath", remotePath)

	// When we launch the server, we kick off a background go routine to
	// ensure the image is available on the host. As a result, this call
	// should be a no-op the majority of the time. Still, there's no harm to
	// having it for additional protection.
	if err := c.containerClient.EnsureImageAvailableOnHost(c.YoutubeDlImageName); err != nil {
		return "", err
	}

	// Mount directory should be specified as contant? Must match the
	// published container.
	youtubeDlMountDirectory := "/downloads"

	// We will use this unique prefix for identifying the file on the file
	// system (as we can't predict the title, extension, etc...)
	uniqueOutputFilePrefix := generateRandomString(8)
	fileNameTemplate := fmt.Sprintf("%s/%s-%%(title)s.%%(ext)s", youtubeDlMountDirectory, uniqueOutputFilePrefix)
	c.logger.V(3).Info("Generated fileNameTemplate", "fileNameTemplate", fileNameTemplate)

	cmd := []string{"-o", fileNameTemplate, remotePath}
	if downloadOptions.audioOnly {
		c.logger.V(2).Info("Restricting download to audio only")
		audioOnlyYoutubeDlOptions := []string{"-x", "--audio-format", defaultAudioFormat}
		cmd = append(audioOnlyYoutubeDlOptions, cmd...)
	}
	c.logger.V(3).Info("Issuing the following args to the containerized youtube dl process", "args", cmd)

	binds := []string{
		fmt.Sprintf("%s:%s", c.fsClient.GetMountDirectory(), youtubeDlMountDirectory),
	}

	runContainerOpts := &runContainerOptions{
		binds: binds,
	}

	if err := c.containerClient.RunContainer(c.YoutubeDlImageName, cmd, runContainerOpts); err != nil {
		return "", err
	}

	return c.findFileUsingUniqueIdentifier(uniqueOutputFilePrefix)
}

// We will use this unique prefix for identifying the file on the file
// system (as we can't predict the title, extension, etc...)
func (c *ContainerYoutubeDlContentDownloader) findFileUsingUniqueIdentifier(uniqueOutputFilePrefix string) (string, error) {
	c.logger.V(3).Info("Identifying file using unique id", "uniqueIdentifier", uniqueOutputFilePrefix)

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

func (c *ContainerYoutubeDlContentDownloader) BestEffortInit() error {
	return c.containerClient.EnsureImageAvailableOnHost(c.YoutubeDlImageName)
}

func NewFakeContentDownloader(fsClient FsClient) *FakeContentDownloader {
	return &FakeContentDownloader{
		fsClient: fsClient,
	}
}

func (f *FakeContentDownloader) DownloadContent(remotePath string, downloadOptions *DownloadOptions) (string, error) {
	fakeFileDownloadPath := path.Join(f.fsClient.GetMountDirectory(), generateRandomString(16))
	fakeFileContents := []byte("hi everyone\n")
	var defaultFilePerm os.FileMode = 0644

	err := ioutil.WriteFile(fakeFileDownloadPath, fakeFileContents, defaultFilePerm)
	if err != nil {
		return "", err
	}

	return fakeFileDownloadPath, nil
}

func (f *FakeContentDownloader) BestEffortInit() error {
	return nil
}
