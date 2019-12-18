unit:
	go test -v -short -count=1 ./...

integration:
	go test -v -count=1 -run Integration ./...

test:
	go test -v -count=1 ./...

run:
	go build -o main && ./main -local -v 3
