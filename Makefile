SHELL := bash

REPOSITORY ?= localhost
CONTAINER_NAME ?= speedtest-exporter
SLIM_TAG ?= slim
CLI_TAG ?= cli

build:
	hack/build.sh

image-slim:
	podman build -t $(REPOSITORY)/$(CONTAINER_NAME):$(SLIM_TAG) .

image-cli:
	podman build -f Dockerfile.cli -t $(REPOSITORY)/$(CONTAINER_NAME):$(CLI_TAG) .

test:
	go test -v -coverprofile=coverprofile.out ./...

coverprofile:
	hack/coverprofile.sh

lint:
	golangci-lint run -v

fmt:
	gofmt -s -w ./cmd ./pkg

validate:
	hack/validate.sh

update-deps:
	hack/update-deps.sh

clean:
	rm -rf bin coverprofiles coverprofile.out

.PHONY: \
	default \
	build \
	image-slim \
	image-cli \
	test \
	coverprofile \
	lint \
	fmt \
	validate \
	update-deps \
	clean \
	$(NULL)
