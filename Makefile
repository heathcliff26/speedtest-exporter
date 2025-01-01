SHELL := bash

REPOSITORY ?= localhost
CONTAINER_NAME ?= speedtest-exporter
SLIM_TAG ?= slim
CLI_TAG ?= cli

default: build

build:
	hack/build.sh

build-slim:
	podman build -t $(REPOSITORY)/$(CONTAINER_NAME):$(SLIM_TAG) .

build-cli:
	podman build -f Dockerfile.cli -t $(REPOSITORY)/$(CONTAINER_NAME):$(CLI_TAG) .

test:
	go test -v ./...

lint:
	golangci-lint run -v

fmt:
	gofmt -s -w ./cmd ./pkg

validate:
	hack/validate.sh

update-deps:
	hack/update-deps.sh

.PHONY: \
	default \
	build \
	build-slim \
	build-cli \
	test \
	lint \
	fmt \
	validate \
	update-deps \
	$(NULL)
