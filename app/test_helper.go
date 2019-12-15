package main

import (
	"testing"
	"time"
)

// We use this url for testing our downloaders... video I own :)
const youtubeURL = "https://www.youtube.com/watch?v=hLswuIQ5Tjk"

// markIntegrationTest indicates a test is an integration test, and should be
// skipped when we run go test with `-short`. We classify integration tests as
// tests with ANY external dependencies (i.e. docker, s3, etc...).
func markIntegrationTest(t *testing.T) {
	t.Helper()

	if testing.Short() {
		t.Skip("skipping integration test")
	}
}

func retryWithTimeout(attempts int, sleep time.Duration, fn func() error) error {
	if err := fn(); err != nil {
		attempts--
		if attempts > 0 {
			time.Sleep(sleep)
			return retryWithTimeout(attempts, sleep, fn)
		}

		return err
	}

	return nil
}
