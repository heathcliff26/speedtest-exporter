#!/bin/bash

set -e

script_dir="$(dirname "${BASH_SOURCE[0]}" | xargs realpath)/.."

pushd "${script_dir}" >/dev/null

OUT_DIR="${script_dir}/coverprofiles"
APP="speedtest-exporter"

if [ ! -d "${OUT_DIR}" ]; then
    mkdir "${OUT_DIR}"
fi

go test -coverprofile="${OUT_DIR}/cover-${APP}.out" "./..."
go tool cover -html "${OUT_DIR}/cover-${APP}.out" -o "${OUT_DIR}/index.html"

popd >/dev/null
