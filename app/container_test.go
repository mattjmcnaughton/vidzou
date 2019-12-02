package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/docker/docker/api/types"
)

func TestDockerClientEnsureImageAvailableOnHostIntegration(t *testing.T) {
	markIntegrationTest(t)

	dockerClient, err := NewDockerClient()
	if err != nil {
		t.Fatalf("Error creating docker client: %s", err)
	}

	demoImageToPull := "alpine:edge"

	err = ensureImageNotOnHost(demoImageToPull)
	if err != nil {
		t.Fatalf("Error ensuring image didn't originally exist on host: %s", err)
	}

	err = dockerClient.EnsureImageAvailableOnHost(demoImageToPull)
	if err != nil {
		t.Fatalf("Error ensuring image available: %s", err)
	}

	testImageAvailableOnHost(t, demoImageToPull)
}

func ensureImageNotOnHost(imageName string) error {
	cli, ctx, err := rawDockerClient()
	if err != nil {
		return err
	}

	_, _, err = cli.ImageInspectWithRaw(ctx, imageName)

	imageDoesNotExistOnHost := err != nil
	if imageDoesNotExistOnHost {
		return nil
	}

	_, err = cli.ImageRemove(ctx, imageName, types.ImageRemoveOptions{})
	return err
}

func testImageAvailableOnHost(t *testing.T, imageName string) {
	t.Helper()

	cli, ctx, err := rawDockerClient()
	if err != nil {
		t.Fatalf("Error creating raw docker client: %s", err)
	}

	_, _, err = cli.ImageInspectWithRaw(ctx, imageName)
	if err != nil {
		t.Fatalf("Image does not exist on host: %s", err)
	}
}

func TestDockerClientEnsureImageAvailableOnHostImageAlreadyExistsIntegration(t *testing.T) {
	markIntegrationTest(t)

	dockerClient, err := NewDockerClient()
	if err != nil {
		t.Fatalf("Error creating docker client: %s", err)
	}

	demoImageToPull := "alpine:edge"

	err = ensureImageNotOnHost(demoImageToPull)
	if err != nil {
		t.Fatalf("Error ensuring image didn't originally exist on host: %s", err)
	}

	for i := 0; i < 2; i++ {
		err = dockerClient.EnsureImageAvailableOnHost(demoImageToPull)
		if err != nil {
			t.Fatalf("Error ensuring image available: %s", err)
		}
		testImageAvailableOnHost(t, demoImageToPull)
	}
}

func TestDockerClientEnsureImageAvailableOnHostFailsOnNonExistentImageIntegration(t *testing.T) {
	markIntegrationTest(t)

	dockerClient, err := NewDockerClient()
	if err != nil {
		t.Fatalf("Error creating docker client: %s", err)
	}

	nonExistentImageToTryAndPull := "alpine:blahblah"
	err = dockerClient.EnsureImageAvailableOnHost(nonExistentImageToTryAndPull)
	if err == nil {
		t.Fatal("Pulling non-existent image should've raised an error")
	}
}

func TestDockerClientRunContainerIntegration(t *testing.T) {
	markIntegrationTest(t)

	dockerClient, err := NewDockerClient()
	if err != nil {
		t.Fatalf("Error creating docker client: %s", err)
	}

	demoImage := "alpine:edge"
	err = dockerClient.EnsureImageAvailableOnHost(demoImage)
	if err != nil {
		t.Fatalf("Error ensuring image available: %s", err)
	}

	alwaysSucceedCmd := []string{
		"/bin/true",
	}

	err = dockerClient.RunContainer(demoImage, alwaysSucceedCmd, &runContainerOptions{})
	if err != nil {
		t.Fatalf("Error running container: %s", err)
	}
}

func TestDockerClientRunContainerFailsWhenContainerFailsIntegration(t *testing.T) {
	markIntegrationTest(t)

	dockerClient, err := NewDockerClient()
	if err != nil {
		t.Fatalf("Error creating docker client: %s", err)
	}

	demoImage := "alpine:edge"
	err = dockerClient.EnsureImageAvailableOnHost(demoImage)
	if err != nil {
		t.Fatalf("Error ensuring image available: %s", err)
	}

	alwaysFailCmd := []string{
		"/bin/false",
	}

	err = dockerClient.RunContainer(demoImage, alwaysFailCmd, &runContainerOptions{})
	if err == nil {
		t.Fatalf("Expected error running container with always fail command.")
	}
}

func TestDockerClientRunContainerWithBindsIntegration(t *testing.T) {
	markIntegrationTest(t)

	dockerClient, err := NewDockerClient()
	if err != nil {
		t.Fatalf("Error creating docker client: %s", err)
	}

	demoImage := "alpine:edge"
	err = dockerClient.EnsureImageAvailableOnHost(demoImage)
	if err != nil {
		t.Fatalf("Error ensuring image available: %s", err)
	}

	useDefaultTempDirectory := ""
	tmpDirectoryPath, err := ioutil.TempDir(useDefaultTempDirectory, "integration-test")
	if err != nil {
		t.Fatalf("Error creating temp dir: %s", err)
	}
	defer os.RemoveAll(tmpDirectoryPath)

	containerDirectoryPath := "/tmp"
	testFileName := "fake-file.txt"
	writeTmpFileCmd := []string{
		"touch",
		fmt.Sprintf("%s/%s", containerDirectoryPath, testFileName),
	}

	runOpts := &runContainerOptions{
		binds: []string{fmt.Sprintf("%s:%s", tmpDirectoryPath, containerDirectoryPath)},
	}

	err = dockerClient.RunContainer(demoImage, writeTmpFileCmd, runOpts)
	if err != nil {
		t.Fatalf("Error running container: %s", err)
	}

	localTestFilePath := filepath.Join(tmpDirectoryPath, testFileName)
	if _, err := os.Stat(localTestFilePath); err != nil {
		t.Fatalf("Shoud be able to stat file on host.")
	}
}

func TestDockerClientRunContainerWithUidIntegration(t *testing.T) {
	markIntegrationTest(t)

	dockerClient, err := NewDockerClient()
	if err != nil {
		t.Fatalf("Error creating docker client: %s", err)
	}

	demoImage := "alpine:edge"
	err = dockerClient.EnsureImageAvailableOnHost(demoImage)
	if err != nil {
		t.Fatalf("Error ensuring image available: %s", err)
	}

	useDefaultTempDirectory := ""
	tmpDirectoryPath, err := ioutil.TempDir(useDefaultTempDirectory, "integration-test")
	if err != nil {
		t.Fatalf("Error creating temp dir: %s", err)
	}
	defer os.RemoveAll(tmpDirectoryPath)

	containerDirectoryPath := "/tmp"
	testFileName := "fake-file.txt"
	writeTmpFileCmd := []string{
		"touch",
		fmt.Sprintf("%s/%s", containerDirectoryPath, testFileName),
	}

	uid := strconv.Itoa(os.Getuid())
	runOpts := &runContainerOptions{
		binds: []string{fmt.Sprintf("%s:%s", tmpDirectoryPath, containerDirectoryPath)},
		uid:   uid,
	}

	err = dockerClient.RunContainer(demoImage, writeTmpFileCmd, runOpts)
	if err != nil {
		t.Fatalf("Error running container: %s", err)
	}

	localTestFilePath := filepath.Join(tmpDirectoryPath, testFileName)
	if _, err := os.Stat(localTestFilePath); err != nil {
		t.Fatalf("Shoud be able to stat file on host.")
	}

	testFile, err := os.OpenFile(localTestFilePath, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		t.Fatalf("Should be able to open file on localhost")
	}
	defer testFile.Close()

	if _, err = testFile.WriteString("hi"); err != nil {
		t.Fatalf("Because we created file in containerized process running as host user, should be able to write to the file: %s", err)
	}
}
