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

BUILDER_IMAGE="docker.io/alukiano/mr-builder@sha256:2da4dffd5622caea5f904d88ad0b403f07f54acecedffe42891573f7c01be94f"
BUILDER_REPO_DIR=/root/go/src/kubevirt.io/machine-remediation
BUILDER_OUT_DIR=$BUILDER_REPO_DIR/_out

OUTPUT_CLIENT_PKG="kubevirt.io/machine-remediation/pkg/client"

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
