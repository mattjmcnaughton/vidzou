package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/user"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

/**
* Development notes:
*
* Need to run `go get github.com/docker/docker@master` to ensure using the
* version of docker which matches the documentation.
*
* When running the container, be sure we are mounting the directory and running
* with the proper user arg.
 */

func main() {
	// http.HandleFunc("/", homePage)
	// http.ListenAndServe(":5000", nil)

	// uploadToS3()
	// generateS3PresignedURL()
	// cleanUpS3Bucket()
}

func homePage(res http.ResponseWriter, req *http.Request) {
	go downloadVideo()
}

func downloadVideo() {
	ctx := context.Background()

	// Should eventually be injected for easier testing?
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	// Will eventually need to add docker pull support...

	// Will eventually use to run as the proper user.
	_, err = user.Current()
	if err != nil {
		panic(err)
	}

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: "mattjmcnaughton/youtube-dl:0.0.1.a",
		Cmd:   []string{"-x", "--audio-format", "mp3", "https://www.youtube.com/watch?v=6QJAevd9uO4"},

		// Currently, we receive a permissions error when we try and run
		// as a specific user. This functionality isn't critical for
		// now, so we can try and address it later.
		//User:  currentUser.Uid,
	}, &container.HostConfig{
		Binds: []string{"/tmp/downloads:/downloads"},
	}, nil, "")

	if err != nil {
		panic(err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	// Wait until the container isn't running anymore.
	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)

	select {
	case err := <-errCh:
		if err != nil {
			panic(err)
		}
	case <-statusCh:
	}

	fmt.Println("Done")
}

func uploadToS3() {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	})

	tmpBucketName := "/web-youtube-dl-test"

	fileToUpload := "/tmp/downloads/pLAyed_Out.mp3"
	file, err := os.Open(fileToUpload)
	if err != nil {
		panic(err)
	}
	// Will eventually want to delete the file.
	defer file.Close()

	uploader := s3manager.NewUploader(sess)
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(tmpBucketName),
		Key:    aws.String("test.mp3"),
		Body:   file,
	})

	fmt.Printf("Upload successful")
}

func generateS3PresignedURL() (string, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	})

	tmpBucketName := "/web-youtube-dl-test"

	svc := s3.New(sess)

	req, _ := svc.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(tmpBucketName),
		Key:    aws.String("test.mp3"),
	})
	urlStr, err := req.Presign(5 * time.Minute)

	if err != nil {
		return "", err
	}

	return urlStr, nil
}

func cleanUpS3Bucket() {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	})

	tmpBucketName := "/web-youtube-dl-test"

	svc := s3.New(sess)

	keysToDelete := []string{}

	resp, err := svc.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: aws.String(tmpBucketName),
	})
	if err != nil {
		panic(err)
	}

	now := time.Now()
	for _, item := range resp.Contents {
		if item.LastModified.Before(now.Add(0 * time.Minute)) {
			fmt.Printf("Deleting %s\n", *item.Key)
			keysToDelete = append(keysToDelete, *item.Key)
		} else {
			fmt.Printf("Not Deleting %s\n", *item.Key)
		}
	}

	for _, key := range keysToDelete {
		_, err = svc.DeleteObject(&s3.DeleteObjectInput{
			Bucket: aws.String(tmpBucketName),
			Key:    aws.String(key),
		})

		if err != nil {
			panic(err)
		}

		err = svc.WaitUntilObjectNotExists(&s3.HeadObjectInput{
			Bucket: aws.String(tmpBucketName),
			Key:    aws.String(key),
		})
	}

	fmt.Println("Finished deleting objects")
}
