VERSION=0.0.1.a
IMAGE="mattjmcnaughton/youtube-dl:$(VERSION)"

build_image:
	docker build -t $(IMAGE) .

publish_image: build_image
	docker push $(IMAGE)
