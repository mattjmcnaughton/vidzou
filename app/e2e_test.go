package main

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func TestE2E(t *testing.T) {
	// Eventually should include some random element.
	testBucket := "web-youtube-dl-test"

	createTestBucket(testBucket)

	downloadVideo()
	uploadToS3()

	presignedURL, err := generateS3PresignedURL()

	if err != nil {
		t.Fatalf("Presigning url failed: %s", err)
	}

	fmt.Println(presignedURL)
	resp, err := http.Get(presignedURL)
	if err != nil {
		t.Fatalf("Error calling GET against presigned URL: %s", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("Request for signed url not successful: %d", resp.StatusCode)
	}

	cleanUpS3Bucket()

	if err := deleteTestBucket(testBucket); err != nil {
		t.Fatalf("Failed to delete test bucket: %s", err)
	}
}

func createTestBucket(bucketName string) error {
	sess, err := getNewSession()
	if err != nil {
		return err
	}

	svc := s3.New(sess)
	_, err = svc.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		return err
	}

	fmt.Printf("Waiting for bucket to be created")

	err = svc.WaitUntilBucketExists(&s3.HeadBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		return err
	}

	return nil
}

func getNewSession() (*session.Session, error) {
	return session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	})
}

func deleteTestBucket(bucketName string) error {
	sess, err := getNewSession()
	if err != nil {
		return err
	}

	svc := s3.New(sess)
	_, err = svc.DeleteBucket(&s3.DeleteBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		return err
	}

	// Assume we delete the bucket successfully... we don't necessarily need
	// to wait for it to be deleted.

	return nil
}
