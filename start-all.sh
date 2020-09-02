#!/usr/bin/env bash

set -Eeuo pipefail

if [[ $# -lt 2 ]] || [[ $# -gt 3 ]]; then
	echo "Usage: $0 <github/repo> <branch|tag|sha> [yaml file name regex filter]
	Eg:	$0 user/repo master '^batch-run-.*\\.yaml$'
" >&2
	exit 1
fi

export GITHUB_REPOSITORY="$1"
export GITHUB_REF="$2"
if [[ $# -eq 3 ]]; then
	export FILTER_REGEX="$3"
fi

if [[ "${GITHUB_TOKEN+found}" != found ]]; then
	echo "need GITHUB_TOKEN" >&2
	exit 1
fi
export GITHUB_TOKEN

go run ./cmd/start-all/*.go
