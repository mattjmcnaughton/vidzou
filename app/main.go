package main

import (
	"flag"
	"k8s.io/klog/v2"
	"k8s.io/klog/v2/klogr"
)

func main() {
	// Initialize a `klog` logger which I'll inject everywhere else.
	klog.InitFlags(nil)
	flag.Set("v", "2")
	logger := klogr.New()

	// TODO: Add prometheus metrics server.

	tmpS3Bucket, err := createTmpS3Bucket()
	if err != nil {
		panic(err)
	}

	fsClient, err := NewTmpFsClient()
	if err != nil {
		panic(err)
	}

	cleanUpFunc := func() error {
		if err := deleteTmpS3Bucket(tmpS3Bucket); err != nil {
			return err
		}

		return fsClient.CleanUp()
	}

	s3ConfigOptions := &s3ConfigurationOptions{
		awsRegion: awsRegion,
		awsBucket: tmpS3Bucket,
	}
	s3Client, err := NewS3Client(s3ConfigOptions)
	if err != nil {
		panic(err)
	}

	downloader, err := NewDockerYoutubeDlContentDownloader(fsClient)
	if err != nil {
		panic(err)
	}
	uploader := NewRemoteStoreContentUploader(s3Client)
	_ = NewRemoteStoreContentGarbageCollector(s3Client)

	server := NewServer(8080, downloader, uploader, logger)
	// POC logger
	logger.V(2).Info("hi")
	err = server.ListenAndServe(cleanUpFunc)

	if err != nil {
		panic(err)
	}

	// Launch garbage collector
}
