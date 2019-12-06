package main

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	dockerclient "github.com/docker/docker/client"
)

// markIntegrationTest indicates a test is an integration test, and should be
// skipped when we run go test with `-short`. We classify integration tests as
// tests with ANY external dependencies (i.e. docker, s3, etc...).
func markIntegrationTest(t *testing.T) {
	t.Helper()

	if testing.Short() {
		t.Skip("skipping integration test")
	}
}

// rawDockerClient creates a "full" docker client (i.e. all functionalities of
// Golang Docker SDK). We use it for test purposes only.
func rawDockerClient() (*dockerclient.Client, context.Context, error) {
	ctx := context.Background()

	cli, err := dockerclient.NewClientWithOpts(dockerclient.FromEnv, dockerclient.WithAPIVersionNegotiation())
	if err != nil {
		return nil, nil, err
	}

	return cli, ctx, nil
}

// rawS3Client creates a "full" s3 client (i.e. all functionalities of Golang S3
// SDK). We use it for test purposes only.
func rawS3Client(awsRegion string) (*s3.S3, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(awsRegion),
	})
	if err != nil {
		return nil, err
	}

	svc := s3.New(sess)

	return svc, nil
}
