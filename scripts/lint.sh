#!/bin/bash

set -o pipefail
echo 'running golangci-lint ...'
golangci-lint run --config .golangci.yml --out-format code-climate | tee "gl-code-quality-report.json" | jq -r '.[] | "\(.location.path):\(.location.lines.begin) \(.description)"'