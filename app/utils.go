package main

import (
	"context"
	"math/rand"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	dockerclient "github.com/docker/docker/client"
)

// Use for generating random string. S3 bucket names can't container uppercase
// letters.
const lowercaseLetters = "abcdefghijklmnopqrstuvxyz"
const awsRegion = "us-east-1"

// TODO: Use throughout the program.
const defaultRandomStringLength = 24

func generateRandomString(randomStringLength int) string {
	ensureRandomness()

	b := make([]byte, randomStringLength)
	for i := range b {
		b[i] = lowercaseLetters[rand.Intn(len(lowercaseLetters))]
	}

	return string(b)
}

// ensureRandomness sets the `rand.Seed` with a new value (i.e.
// `time.Now.UnixNano()`) to guarantee new random values. See
// https://golang.org/pkg/math/rand/ for more info.
//
// Without this function, we will always use the same "random" name for our s3
// bucket.
func ensureRandomness() {
	rand.Seed(time.Now().UnixNano())
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

func createTmpS3Bucket() (string, error) {
	svc, err := rawS3Client(awsRegion)
	if err != nil {
		return "", err
	}

	bucketNameLength := 16
	randomBucketName := generateRandomString(bucketNameLength)

	_, err = svc.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(randomBucketName),
	})
	if err != nil {
		return "", err
	}

	err = svc.WaitUntilBucketExists(&s3.HeadBucketInput{
		Bucket: aws.String(randomBucketName),
	})
	if err != nil {
		return "", err
	}

	return randomBucketName, nil
}

func deleteTmpS3Bucket(tmpS3BucketName string) error {
	svc, err := rawS3Client(awsRegion)
	if err != nil {
		return err
	}

	if err := prepareTmpS3BucketForDeletion(svc, tmpS3BucketName); err != nil {
		return err
	}

	_, err = svc.DeleteBucket(&s3.DeleteBucketInput{
		Bucket: aws.String(tmpS3BucketName),
	})
	if err != nil {
		return err
	}

	// We do not wait for the bucket to actually not appear anymore, because
	// that operation can sometimes take a really long time... since we are
	// using a unique name for each bucket, we aren't super concerned about
	// starting a new test suite before the old bucket has been completely
	// cleaned up.

	return nil
}

func prepareTmpS3BucketForDeletion(svc *s3.S3, tmpS3BucketName string) error {
	return deleteAllObjectsInBucket(svc, tmpS3BucketName)
}

// deleteAllObjectsInBucket exists because s3 requires a bucket be empty before
// we delete it.
func deleteAllObjectsInBucket(svc *s3.S3, tmpS3BucketName string) error {
	deleteIter := s3manager.NewDeleteListIterator(svc, &s3.ListObjectsInput{
		Bucket: aws.String(tmpS3BucketName),
	})

	return s3manager.NewBatchDeleteWithClient(svc).Delete(aws.BackgroundContext(), deleteIter)
}
