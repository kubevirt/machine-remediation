#!/usr/bin/env bash

set -e

source $(dirname "$0")/../common.sh

# remove generate manifests
rm -rf ${REPO_DIR}/manifests/generated/*

# remove generated client
rm -rf ${REPO_DIR}/pkg/client/*

# remove generate files
find ${REPO_DIR}/pkg/ -name "*generated*.go" -exec rm {} -f \;
