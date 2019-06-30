#!/usr/bin/env bash

set -e

source hack/common.sh

shfmt -i 4 -w ${REPO_DIR}/hack/
bazel run \
    --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 \
    --workspace_status_command=./hack/print-workspace-status.sh \
    //:gazelle -- pkg/ tools/ cmd/
bazel run \
    --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 \
    --workspace_status_command=./hack/print-workspace-status.sh \
    //:goimports
# allign BAZEL files to a single format
bazel run \
    --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 \
    --workspace_status_command=./hack/print-workspace-status.sh \
    //:buildifier
