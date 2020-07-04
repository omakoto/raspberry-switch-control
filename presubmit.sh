#!/bin/bash

set -e

cd "${0%/*}/nscontroller/"

go get -v -t honnef.co/go/tools/cmd/...
go get -v -t golang.org/x/lint/golint

gofmt -s -d $(find . -type f -name '*.go') |& perl -pe 'END{exit($. > 0 ? 1 : 0)}'

go test -v -race ./...                   # Run all the tests with the race detector enabled

echo "Running extra checks..."
go vet ./...                             # go vet is the official Go static analyzer
staticcheck ./...
golint $(go list ./...) | grep -Pv '(exported .* should have)' | perl -pe 'END{exit($. > 0 ? 1 : 0)}'
