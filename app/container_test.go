package main

import (
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

func TestRunContainerIntegration(t *testing.T) {
	markIntegrationTest(t)
	t.Skip("Need to write")
}

func TestRunContainerFailsWhenContainerFails(t *testing.T) {
	markIntegrationTest(t)
	t.Skip("Need to write")
}
