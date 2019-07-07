#!/usr/bin/env bash

if [ -f cluster-up/hack/common.sh ]; then
    source cluster-up/hack/common.sh
fi

REPO_DIR="$(
    cd "$(dirname "$BASH_SOURCE[0]")/../"
    pwd
)"
OUT_DIR=$REPO_DIR/_out
VENDOR_DIR=$REPO_DIR/vendor
CMD_OUT_DIR=$OUT_DIR/cmd
MANIFESTS_OUT_DIR=$OUT_DIR/manifests

BUILDER_REPO_DIR=/root/go/src/github.com/openshift/machine-remediation-operator
BUILDER_OUT_DIR=$BUILDER_REPO_DIR/_out
