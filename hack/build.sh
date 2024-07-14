#!/bin/bash

set -e

base_dir="$(dirname "${BASH_SOURCE[0]}" | xargs realpath)/.."

bin_dir="${base_dir}/bin"

GOOS="${GOOS:-$(go env GOOS)}"
GOARCH="${GOARCH:-$(go env GOARCH)}"

GO_LD_FLAGS="${GO_LD_FLAGS:-"-s"}"

if [ "${RELEASE_VERSION}" != "" ]; then
    echo "Building release version ${RELEASE_VERSION}"
    GO_LD_FLAGS+=" -X github.com/heathcliff26/speedtest-exporter/pkg/version.version=${RELEASE_VERSION}"
fi

output_name="${bin_dir}/speedtest-exporter"
if [ "${GOOS}" == "windows" ]; then
    output_name="${output_name}.exe"
fi

pushd "${base_dir}" >/dev/null

echo "Building $(basename "${output_name}")"
GOOS="${GOOS}" GOARCH="${GOARCH}" CGO_ENABLED=0 go build -ldflags="${GO_LD_FLAGS}" -o "${output_name}" ./cmd/...
