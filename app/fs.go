package main

import (
	"io/ioutil"
	"os"
	"path"
)

// TODO: Currently only ContentDownloader uses the `FsClient`. The
// `ContentUploader` just reads directly from the file system using the full
// path. I should think more about if `ContentUploader` should also be using
// `FsClient`...

type FsClient interface {
	GetMountDirectory() string
	GeneratePathForFile(fileName string) string
}

type TmpFsClient struct {
	baseDirectory string
}

func NewTmpFsClient() (*TmpFsClient, error) {
	useDefaultTempDirectory := ""
	prefixForDirName := ""

	tmpDir, err := ioutil.TempDir(useDefaultTempDirectory, prefixForDirName)
	if err != nil {
		return nil, err
	}

	tmpFsClient := &TmpFsClient{
		baseDirectory: tmpDir,
	}
	return tmpFsClient, nil
}

func (t *TmpFsClient) GetMountDirectory() string {
	return t.baseDirectory
}

func (t *TmpFsClient) GeneratePathForFile(fileName string) string {
	return path.Join(t.baseDirectory, fileName)
}

func (t *TmpFsClient) CleanUp() error {
	return os.RemoveAll(t.baseDirectory)
}
