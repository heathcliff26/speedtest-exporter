#!/bin/bash

base_dir="$(dirname "${0}" | xargs realpath)"

cat "${base_dir}/speedtest-cli-result.json"
