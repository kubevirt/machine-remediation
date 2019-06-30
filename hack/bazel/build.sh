#!/usr/bin/env bash

set -e

source hack/common.sh

rm -rf ${CMD_OUT_DIR}

# Build all binaries for amd64
bazel build \
    --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 \
    --workspace_status_command=./hack/print-workspace-status.sh \
    //cmd/...
