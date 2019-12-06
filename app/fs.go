package main

import (
	"io/ioutil"
	"os"
	"path"
)

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
