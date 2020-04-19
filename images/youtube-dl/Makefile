VERSION=2020.03.24
IMAGE="mattjmcnaughton/youtube-dl:$(VERSION)"

build_image:
	docker build -t $(IMAGE) .

publish_image: build_image
	docker push $(IMAGE)
