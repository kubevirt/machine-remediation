#!/usr/bin/env bash

set -e

source hack/config-default.sh

bazel run \
    --host_force_python=PY2 \
    --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 \
    --workspace_status_command=./hack/print-workspace-status.sh \
    --define container_prefix=${container_prefix} \
    --define container_tag=${container_tag} \
    //:push-images
