package main

import (
	"math/rand"
	"time"
)

// Use for generating random string. S3 bucket names can't container uppercase
// letters.
const lowercaseLetters = "abcdefghijklmnopqrstuvxyz"

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
