#!/usr/bin/env bash

set -Eeuo pipefail


export GITHUB_REPOSITORY="GongT/containers"
export GITHUB_TOKEN="96c35b58ded50bfdd3bd3ecd530c897d2a7e8815 "
export CANCEL_ALL=yes

go run ./runoverworkflows.go
