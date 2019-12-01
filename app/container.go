package main

import (
	"context"
	"io"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	dockerclient "github.com/docker/docker/client"
)

type ContainerClient interface {
	EnsureImageAvailableOnHost(imageName string) error
	RunContainer(imageName string, cmd []string, runContainerOpts *runContainerOpts) error
}

type runContainerOpts struct {
	binds []string
	uid   string
}

type DockerClient struct {
	cli *dockerclient.Client
	ctx context.Context
}

// Allows me to easily swap the container runtime engine.
var _ ContainerClient = (*DockerClient)(nil)

func NewDockerClient() (*DockerClient, error) {
	ctx := context.Background()

	cli, err := dockerclient.NewClientWithOpts(dockerclient.FromEnv, dockerclient.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	dockerClient := &DockerClient{
		cli: cli,
		ctx: ctx,
	}

	return dockerClient, nil
}

func (dc *DockerClient) EnsureImageAvailableOnHost(imageName string) error {
	reader, err := dc.cli.ImagePull(dc.ctx, imageName, types.ImagePullOptions{})
	if err != nil {
		return err
	}

	// We need the `io.Copy`, or else will not actually download the
	// image...
	io.Copy(os.Stdout, reader)
	return nil
}

func (dc *DockerClient) RunContainer(imageName string, cmd []string, runContainerOpts *runContainerOpts) error {
	containerConfig := &container.Config{
		Image: imageName,
		Cmd:   cmd,
	}
	hostConfig := &container.HostConfig{}

	if runContainerOpts.uid != "" {
		containerConfig.User = runContainerOpts.uid
	}
	if runContainerOpts.binds != nil {
		hostConfig.Binds = runContainerOpts.binds
	}

	// Use better descriptive variable names here.
	createContainerResp, err := dc.cli.ContainerCreate(dc.ctx, containerConfig, hostConfig, nil, "")
	if err != nil {
		return err
	}

	err = dc.cli.ContainerStart(dc.ctx, createContainerResp.ID, types.ContainerStartOptions{})
	if err != nil {
		return err
	}

	statusCh, errCh := dc.cli.ContainerWait(dc.ctx, createContainerResp.ID, container.WaitConditionNotRunning)

	select {
	case err := <-errCh:
		if err != nil {
			return err
		}
	case <-statusCh:
		// No-opt when container is finished executing. Our code
		// blocking until the container is done is the important
		// behavior.
	}

	return nil
}
