#!/usr/bin/env bash

# Generate BUILD files
bazel run \
    --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 \
    --workspace_status_command=./hack/print-workspace-status.sh \
    //:gazelle

# Allign BAZEL files to a single format
bazel run \
    --incompatible_disable_deprecated_attr_params=false \
    --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 \
    --workspace_status_command=./hack/print-workspace-status.sh \
    //:buildifier
