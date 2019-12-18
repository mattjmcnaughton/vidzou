package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	dockerclient "github.com/docker/docker/client"
	"github.com/go-logr/logr"
)

// ContainerClient defines an interface, implementable by a number of different
// container runtimes (i.e. docker, containerd, etc...),
// for the containerized options we need in this application.
type ContainerClient interface {
	EnsureImageAvailableOnHost(imageName string) error
	RunContainer(imageName string, cmd []string, runContainerOpts *runContainerOptions) error
}

// runContainerOptions aggregates common options for running containers. We
// restrict the options only to one's our application actually needs.
type runContainerOptions struct {
	binds []string
	uid   string
}

// DockerClient defines a wrapper around the Docker Golang SDK.
type DockerClient struct {
	cli *dockerclient.Client
	ctx context.Context

	logger logr.Logger
}

// Ensure all client implementations fulfill the ContainerClient interface.
var _ ContainerClient = (*DockerClient)(nil)

// NewDockerClient creates a new Docker client and returns it.
func NewDockerClient(logger logr.Logger) (*DockerClient, error) {
	ctx := context.Background()

	logger.V(3).Info("Creating new raw docker client")
	cli, err := dockerclient.NewClientWithOpts(dockerclient.FromEnv, dockerclient.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	dockerClient := &DockerClient{
		cli:    cli,
		ctx:    ctx,
		logger: logger,
	}

	return dockerClient, nil
}

// EnsureImageAvailableOnHost ensures that a container image exists on the host
// (i.e. could be used to run a container). We return an error only if we were
// unable to ensure the image exists on the host.
func (dc *DockerClient) EnsureImageAvailableOnHost(imageName string) error {
	dc.logger.V(3).Info("Ensuring image exists on host", "imageName", imageName)

	_, _, err := dc.cli.ImageInspectWithRaw(dc.ctx, imageName)

	imageExistsOnHost := err == nil
	if imageExistsOnHost {
		return nil
	}

	dc.logger.V(3).Info("Pulling image onto host", "imageName", imageName)
	reader, err := dc.cli.ImagePull(dc.ctx, imageName, types.ImagePullOptions{})
	if err != nil {
		return err
	}

	// We need the `io.Copy`, or else will not actually download the
	// image...
	io.Copy(os.Stdout, reader)
	return nil
}

// RunContainer runs a container. We return a non-nil error either if there is
// an error running the container or the exit code of the containerized process
// is non-zero.
func (dc *DockerClient) RunContainer(imageName string, cmd []string, runContainerOpts *runContainerOptions) error {
	dc.logger.V(3).Info("Running container with following settings", "imageName", imageName, "cmd", cmd, "runContainerOptions", runContainerOpts)

	containerConfig := &container.Config{
		Image: imageName,
		Cmd:   cmd,
	}
	hostConfig := &container.HostConfig{
		AutoRemove: true,
	}

	if runContainerOpts.uid != "" {
		containerConfig.User = runContainerOpts.uid
	}
	if runContainerOpts.binds != nil {
		hostConfig.Binds = runContainerOpts.binds
	}

	createContainerResp, err := dc.cli.ContainerCreate(dc.ctx, containerConfig, hostConfig, nil, "")
	if err != nil {
		return err
	}

	dc.logger.V(3).Info("Starting container")
	err = dc.cli.ContainerStart(dc.ctx, createContainerResp.ID, types.ContainerStartOptions{})
	if err != nil {
		return err
	}

	dc.logger.V(3).Info("Waiting for container to finish executing")
	statusCh, errCh := dc.cli.ContainerWait(dc.ctx, createContainerResp.ID, container.WaitConditionNotRunning)

	statusCodesIndicatingSuccess := map[int]bool{0: true}

	select {
	case err := <-errCh:
		dc.logger.V(3).Info("No longer waiting on container")

		// errCh passes an error if there was an issue waiting for the
		// container... NOT if the container had an error while
		// executing.
		if err != nil {
			return err
		}
	case resp := <-statusCh:
		dc.logger.V(3).Info("No longer waiting on container")

		if _, ok := statusCodesIndicatingSuccess[int(resp.StatusCode)]; !ok {
			return fmt.Errorf("Container %s finished with non-zero exit code.", createContainerResp.ID)
		}
	}

	dc.logger.V(3).Info("No longer waiting on container")
	return nil
}
