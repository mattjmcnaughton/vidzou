TIMESTAMP=$(shell date "+%s")
IMAGE = "mattjmcnaughton/vidzou:$(TIMESTAMP)"

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

build_image:
	docker build -t $(IMAGE) .

publish_image:
	docker build -t $(IMAGE) .
	docker push $(IMAGE)
