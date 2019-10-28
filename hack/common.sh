#!/usr/bin/env bash

if [ -f cluster-up/hack/common.sh ]; then
    source cluster-up/hack/common.sh
fi

REPO_DIR="$(
    cd "$(dirname "$BASH_SOURCE[0]")/../"
    pwd
)"
OUT_DIR=${REPO_DIR}/_out
VENDOR_DIR=${REPO_DIR}/vendor
GENERATED_MANIFESTS_DIR=${REPO_DIR}/manifests/generated
BIN_OUT_DIR=${OUT_DIR}/bin
MANIFESTS_OUT_DIR=${OUT_DIR}/manifests
TESTS_OUT_DIR=${OUT_DIR}/tests

BUILDER_IMAGE="docker.io/alukiano/mro-builder@sha256:56a4ebcbc5c43e64c5a65b4889dcfd1f3cb6e88739973feecc68e2ac0d321b83"
BUILDER_REPO_DIR=/root/go/src/kubevirt.io/machine-remediation-operator
BUILDER_OUT_DIR=$BUILDER_REPO_DIR/_out

OUTPUT_CLIENT_PKG="kubevirt.io/machine-remediation-operator/pkg/client"

function version() {
    if [ -n "${REPO_VERSION}" ]; then
        echo ${REPO_VERSION}
    elif [ -d ${REPO_DIR}/.git ]; then
        echo "$(git describe --always --tags)"
    else
        echo "undefined"
    fi
}
REPO_VERSION="$(version)"
