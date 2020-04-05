package main

import (
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	logrtesting "github.com/go-logr/logr/testing"
)

// We use this url for testing our downloaders... video I own :)
const youtubeURL = "https://www.youtube.com/watch?v=hLswuIQ5Tjk"

// We may not want this to be a NullLogger long term...
var testLogger = logrtesting.NullLogger{}

// markIntegrationTest indicates a test is an integration test, and should be
// skipped when we run go test with `-short`. We classify integration tests as
// tests with ANY external dependencies (i.e. docker, s3, etc...).
func markIntegrationTest(t *testing.T) {
	t.Helper()

	if testing.Short() {
		t.Skip("skipping integration test")
	}
}

func retryWithTimeout(attempts int, sleep time.Duration, fn func() error) error {
	if err := fn(); err != nil {
		attempts--
		if attempts > 0 {
			time.Sleep(sleep)
			return retryWithTimeout(attempts, sleep, fn)
		}

		return err
	}

	return nil
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
