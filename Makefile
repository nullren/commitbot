NAME := commitbot
IMAGE := docker.artifactory.rb/$(USER)/$(NAME)
IMAGETAG := $(IMAGE):$(shell git show-ref --head --hash=7 | head -n1)

build:
	docker build -t $(IMAGETAG) .

push:
	docker push $(IMAGETAG)

.PHONY: build push
