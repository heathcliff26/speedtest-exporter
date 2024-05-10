SHELL := bash

REPOSITORY ?= localhost
CONTAINER_NAME ?= speedtest-exporter
SLIM_TAG ?= slim
CLI_TAG ?= cli

GO_BUILD_FLAGS ?= -ldflags="-w -s"

default: build

build:
	go build $(GO_BUILD_FLAGS) -o bin/speedtest-exporter ./cmd/

build-slim:
	podman build -t $(REPOSITORY)/$(CONTAINER_NAME):$(SLIM_TAG) .

build-cli:
	podman build -f Dockerfile.cli -t $(REPOSITORY)/$(CONTAINER_NAME):$(CLI_TAG) .

test:
	go test -v ./...

update-deps:
	hack/update-deps.sh

.PHONY: \
	default \
	build \
	build-slim \
	build-cli \
	test \
	update-deps \
	$(NULL)
