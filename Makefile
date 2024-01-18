SHELL := bash

REPOSITORY ?= localhost
CONTAINER_NAME ?= speedtest-exporter
SLIM_TAG ?= slim
CLI_TAG ?= cli

GO_BUILD_FLAGS ?= -ldflags="-w -s"

default: build-slim

build: build-slim build-cli

build-slim:
	podman build -t $(REPOSITORY)/$(CONTAINER_NAME):$(SLIM_TAG) .

build-cli:
	podman build -f Dockerfile.cli -t $(REPOSITORY)/$(CONTAINER_NAME):$(CLI_TAG) .

go-build:
	go build $(GO_BUILD_FLAGS) -o bin/speedtest-exporter ./cmd/

go-test:
	go test -v ./...

.PHONY: \
	default \
	build \
	build-slim \
	build-cli \
	go-build \
	go-test \
	$(NULL)
