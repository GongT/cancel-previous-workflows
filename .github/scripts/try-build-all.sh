#!/usr/bin/env bash

set -Eeuo pipefail

x() {
	echo " + $*"
	"$@"
}

for I in cmd/*/; do
	x go build -o "$(mktemp --dry-run)" "$I/"*.go
done
