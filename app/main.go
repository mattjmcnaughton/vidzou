package main

import (
	"flag"
	"k8s.io/klog/v2"
	"k8s.io/klog/v2/klogr"
	"os"
	"strconv"
)

// High level logging guidelines... use log level 0 for information which MUST
// be known. Use 2 for generally useful information we want to show by default.
// Use 3 for additional info which is helpful for development/debugging.
const defaultLogLevel = 2

var runningLocally = flag.Bool("local", false, "run app locally")
var s3Bucket = flag.String("s3_bucket", os.Getenv("VIDZOU_S3_BUCKET"), "s3 bucket in which to store info")

func main() {
	initAndParseFlags()
	logger := klogr.New()

	var s3BucketName string
	var s3CleanUp func() error
	var err error

	if *runningLocally {
		logger.V(2).Info("Running locally... create tmp s3 bucket")

		s3BucketName, err = createTmpS3Bucket()
		if err != nil {
			panic(err)
		}

		s3CleanUp = func() error {
			logger.V(2).Info("Removing tmp s3 bucket")
			return deleteTmpS3Bucket(s3BucketName)
		}
	} else {
		logger.V(2).Info("Running remotely... do not attempt create s3 bucket")
		if len(*s3Bucket) == 0 {
			panic("Must pass valid s3_bucket")
		}

		s3BucketName = *s3Bucket
		s3CleanUp = func() error {
			return nil
		}
	}
	logger.V(2).Info("S3 bucket exists", "bucketName", s3BucketName)

	logger.V(3).Info("Creating foundational clients")
	fsClient, err := NewTmpFsClient()
	if err != nil {
		panic(err)
	}

	cleanUpFunc := func() error {
		if err := s3CleanUp(); err != nil {
			return err
		}

		return fsClient.CleanUp()
	}

	s3ConfigOptions := &s3ConfigurationOptions{
		awsRegion: awsRegion,
		awsBucket: s3BucketName,
	}
	s3Client, err := NewS3Client(s3ConfigOptions, logger)
	if err != nil {
		panic(err)
	}

	logger.V(3).Info("Creating all content managers")
	downloader, err := NewDockerYoutubeDlContentDownloader(fsClient, logger)
	if err != nil {
		panic(err)
	}
	uploader := NewRemoteStoreContentUploader(s3Client, logger)
	_ = NewRemoteStoreContentGarbageCollector(s3Client, logger)

	// @TODO(mattjmcnaughton) Launch garbage collector in a separate goroutine.

	server := NewServer(8080, downloader, uploader, logger)
	err = server.ListenAndServe(cleanUpFunc)

	logger.V(2).Info("Terminating program")
	if err != nil {
		panic(err)
	}
}

func initAndParseFlags() {
	klog.InitFlags(nil)
	flag.Set("v", strconv.Itoa(defaultLogLevel))
	flag.Parse()
}
