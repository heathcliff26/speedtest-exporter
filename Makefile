SHELL := bash

REPOSITORY ?= localhost
CONTAINER_NAME ?= speedtest-exporter
SLIM_TAG ?= slim
CLI_TAG ?= cli

# Build the binary
build:
	hack/build.sh

# Build the slim container image
image-slim:
	podman build -t $(REPOSITORY)/$(CONTAINER_NAME):$(SLIM_TAG) .

# Build the CLI container image
image-cli:
	podman build -f Dockerfile.cli -t $(REPOSITORY)/$(CONTAINER_NAME):$(CLI_TAG) .

# Run unit tests
test:
	go test -v -coverprofile=coverprofile.out ./...

# Generate cover profile
coverprofile:
	hack/coverprofile.sh

# Run linter
lint:
	golangci-lint run -v

# Format code
fmt:
	gofmt -s -w ./cmd ./pkg

# Validate that all generated files are up to date
validate:
	hack/validate.sh

# Update dependencies
update-deps:
	hack/update-deps.sh

# Scan code for vulnerabilities using gosec
gosec:
	gosec ./...

# Clean build artifacts
clean:
	rm -rf bin coverprofiles coverprofile.out

# Show this help message
help:
	@echo "Available targets:"
	@echo ""
	@awk '/^#/{c=substr($$0,3);next}c&&/^[[:alpha:]][[:alnum:]_-]+:/{print substr($$1,1,index($$1,":")),c}1{c=0}' $(MAKEFILE_LIST) | column -s: -t
	@echo ""
	@echo "Run 'make <target>' to execute a specific target."

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
	gosec \
	clean \
	help \
	$(NULL)
