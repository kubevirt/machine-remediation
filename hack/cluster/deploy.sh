#!/usr/bin/env bash

set -e

source $(dirname "$0")/../common.sh
source $(dirname "$0")/../config.sh
source ${REPO_DIR}/cluster-up/cluster/${KUBEVIRT_PROVIDER}/provider.sh

CONTAINER_TAG=${CONTAINER_TAG:-devel}

# Deploy MRO operator
_kubectl create -f ${MANIFESTS_OUT_DIR}/release/machine-remediation-operator.${CONTAINER_TAG}.yaml

until [[ $(_kubectl get crd machineremediationoperators.machineremediation.kubevirt.io --no-headers) != "" ]]; do
    sleep 5
done

# Deploy MRO
_kubectl create -f ${MANIFESTS_OUT_DIR}/release/machine-remediation-operator-cr.${CONTAINER_TAG}.yaml

until [[ $(_kubectl -n ${namespace} get mro mro --no-headers) != "" ]]; do
    sleep 5
done

# wait until MRO is ready
_kubectl wait -n ${namespace} mro mro --for condition=Available --timeout 2m || (echo "MRO not ready in time" && exit 1)

echo "Done"
