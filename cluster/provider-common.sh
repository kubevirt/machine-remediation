#!/bin/bash

set -e

_cli="docker run --privileged --net=host --rm ${USE_TTY} -v /var/run/docker.sock:/var/run/docker.sock kubevirtci/gocli@sha256:df958c060ca8d90701a1b592400b33852029979ad6d5c1d9b79683033704b690"

function _main_ip() {
    echo 127.0.0.1
}

function _port() {
    ${_cli} ports --prefix $provider_prefix "$@"
}

function _registry_volume() {
    echo ${job_prefix}_registry
}

function _add_common_params() {
    local params="--nodes ${CLUSTER_NUM_NODES} --memory 4096M --cpu 4 --random-ports --background --prefix $provider_prefix --registry-volume $(_registry_volume) kubevirtci/${image} ${CLUSTER_PROVIDER_EXTRA_ARGS}"
    echo $params
}

function build() {
    # Build everyting and publish it
    make bazel-generate
    make bazel-generate-manifests-dev
    make bazel-push-images-$CLUSTER_PROVIDER
}

function _kubectl() {
    export KUBECONFIG=${REPO_PATH}cluster/$CLUSTER_PROVIDER/.kubeconfig
    ${REPO_PATH}cluster/$CLUSTER_PROVIDER/.kubectl "$@"
}

function down() {
    ${_cli} rm --prefix $provider_prefix
}
