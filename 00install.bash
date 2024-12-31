#!/bin/bash

set -e
SCRIPT_DIR="${0%/*}"

cd "$SCRIPT_DIR"/nscontroller/cmd/
go install ./...
