#!/bin/bash

set -o pipefail
echo 'running go test ...'
go test -coverprofile=coverage.out ./lib/... && go tool cover -func=coverage.out | grep total | awk '{print "total: " $3}' > coverage.txt && cat coverage.txt