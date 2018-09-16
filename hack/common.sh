#!/bin/bash

REPO_DIR="$(
    cd "$(dirname "$BASH_SOURCE[0]")/../"
    pwd
)"
OUT_DIR=$REPO_DIR/_out
VENDOR_DIR=$REPO_DIR/vendor
CMD_OUT_DIR=$OUT_DIR/cmd
TESTS_OUT_DIR=$OUT_DIR/tests
MANIFESTS_OUT_DIR=$OUT_DIR/manifests

function build_func_tests() {
    mkdir -p ${TESTS_OUT_DIR}/
    ginkgo build ${REPO_DIR}/tests
    mv ${REPO_DIR}/tests/tests.test ${TESTS_OUT_DIR}/
}

CLUSTER_PROVIDER=${CLUSTER_PROVIDER:-k8s-1.10.4}
CLUSTER_NUM_NODES=${CLUSTER_NUM_NODES:-1}

# If on a developer setup, expose ocp on 8443, so that the openshift web console can be used (the port is important because of auth redirects)
if [ -z "${JOB_NAME}" ]; then
    CLUSTER_PROVIDER_EXTRA_ARGS="${KUBEVIRT_PROVIDER_EXTRA_ARGS} --ocp-port 8443"
fi

#If run on jenkins, let us create isolated environments based on the job and
# the executor number
provider_prefix=${JOB_NAME:-${CLUSTER_PROVIDER}}${EXECUTOR_NUMBER}
job_prefix=${JOB_NAME:-noderecovery}${EXECUTOR_NUMBER}

# Populate an environment variable with the version info needed.
# It should be used for everything which needs a version when building (not generating)
# IMPORTANT:
# RIGHT NOW ONLY RELEVANT FOR BUILDING, GENERATING CODE OUTSIDE OF GIT
# IS NOT NEEDED NOR RECOMMENDED AT THIS STAGE.

function noderecovery_version() {
    if [ -n "${NODERECOVERY_VERSION}" ]; then
        echo ${NODERECOVERY_VERSION}
    elif [ -d ${REPO_DIR}/.git ]; then
        echo "$(git describe --always --tags)"
    else
        echo "undefined"
    fi
}
NODERECOVERY_VERSION="$(noderecovery_version)"
