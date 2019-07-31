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
CMD_OUT_DIR=${OUT_DIR}/cmd
MANIFESTS_OUT_DIR=${OUT_DIR}/manifests
TESTS_OUT_DIR=${OUT_DIR}/tests

BUILDER_REPO_DIR=/root/go/src/kubevirt.io/machine-remediation-operator
BUILDER_OUT_DIR=$BUILDER_REPO_DIR/_out

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
