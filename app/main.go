package main

func main() {
	tmpS3Bucket, err := createTmpS3Bucket()
	if err != nil {
		panic(err)
	}
	defer deleteTmpS3Bucket(tmpS3Bucket)

	fsClient, err := NewTmpFsClient()
	if err != nil {
		panic(err)
	}
	defer fsClient.CleanUp()

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

	// The way we are killing the program in dev mode (with `ctrl-c`) is
	// causing the go process to termiante before we run the defers. As a
	// result, we need to manually clean up the s3 buckets. Add better
	// signal handling (maybe using `braintree/manners`).
	server := NewServer(8080, downloader, uploader)
	server.ListenAndServe()

	// Launch garbage collector
}
