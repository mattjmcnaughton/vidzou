VERSION=2020.05.29
IMAGE="mattjmcnaughton/youtube-dl:$(VERSION)"

build_image:
	docker build -t $(IMAGE) .

publish_image: build_image
	docker push $(IMAGE)
