package main

import (
	"context"
	dockerclient "github.com/docker/docker/client"
	"testing"
)

func markIntegrationTest(t *testing.T) {
	t.Helper()

	if testing.Short() {
		t.Skip("skipping integration test")
	}
}

// A raw docker client for test only operations.
func rawDockerClient() (*dockerclient.Client, context.Context, error) {
	ctx := context.Background()

	cli, err := dockerclient.NewClientWithOpts(dockerclient.FromEnv, dockerclient.WithAPIVersionNegotiation())
	if err != nil {
		return nil, nil, err
	}

	return cli, ctx, nil
}
