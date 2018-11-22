
.PHONY: image push 

GCR_BUCKET = beekhof
PREFIX = quay.io/$(GCR_BUCKET)
DEV_PREFIX ?= quay.io/beekhof
NAME = fence-agents
TAG = 0.0.11

image:
	docker build -t "$(PREFIX)/$(NAME):$(TAG)" -f ./Dockerfile .

push: image
	docker push "$(PREFIX)/$(NAME):$(TAG)"

