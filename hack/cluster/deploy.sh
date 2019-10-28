#!/usr/bin/env bash

set -e

source $(dirname "$0")/../common.sh
source $(dirname "$0")/../config.sh
source ${REPO_DIR}/cluster-up/cluster/${KUBEVIRT_PROVIDER}/provider.sh

CONTAINER_TAG=${CONTAINER_TAG:-devel}

# Deploy MR operator
_kubectl create -f ${MANIFESTS_OUT_DIR}/release/machine-remediation.${CONTAINER_TAG}.yaml

# wait until MR is ready
_kubectl wait -n ${namespace} --for condition=Available --timeout 2m deployment/machine-remediation || (echo "MO not ready in time" && exit 1)

echo "Done"
