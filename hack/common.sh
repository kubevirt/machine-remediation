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

CLUSTER_PROVIDER=${CLUSTER_PROVIDER:-k8s-1.11.0}
CLUSTER_NUM_NODES=${CLUSTER_NUM_NODES:-1}
CLUSTER_PROVIDER_KUBECONFIG=$REPO_DIR/cluster/$CLUSTER_PROVIDER/.kubeconfig

CONTAINERS_PREFIX=${CONTAINERS_PREFIX:-docker.io/kubevirt}
CONTAINERS_TAG=${CONTAINERS_TAG:-latest}

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
