#!/usr/bin/env bash

set -Eeuo pipefail

if [[ $# -ne 1 ]]; then
	echo "Usage: $0 <github/repo>
" >&2
	exit 1
fi

export GITHUB_REPOSITORY="$1"

if [[ "${GITHUB_TOKEN+found}" != found ]]; then
	echo "need GITHUB_TOKEN" >&2
	exit 1
fi
export GITHUB_TOKEN

go run ./cmd/restart-failed/*.go
