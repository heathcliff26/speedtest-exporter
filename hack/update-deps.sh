#!/bin/bash

set -e

script_dir="$(dirname "${BASH_SOURCE[0]}" | xargs realpath)/.."

pushd "${script_dir}" >/dev/null
go get -u ./...
go mod tidy
go mod vendor
popd >/dev/null
