NAME := commitbot
IMAGE := docker.artifactory.rb/$(USER)/$(NAME)
IMAGETAG := $(IMAGE):$(shell git show-ref --head --hash=7 | head -n1)

build:
	docker build -t $(IMAGETAG) .

push:
	docker push $(IMAGETAG)

run:
	docker run -it $(IMAGETAG) bash

.PHONY: build push run
