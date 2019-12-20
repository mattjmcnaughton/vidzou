unit:
	go test -v -short -count=1 ./...

integration:
	go test -v -count=1 -run Integration ./...

test:
	go test -v -count=1 ./...

build:
	go build -o main

run: build
	./main -local -v 3
