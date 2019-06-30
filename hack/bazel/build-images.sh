#!/usr/bin/env bash

set -e

bazel build \
    --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 \
    --workspace_status_command=./hack/print-workspace-status.sh \
    --define container_prefix=${CONTAINER_PREFIX} \
    --define container_tag=${CONTAINER_TAG} \
    //:build-images
