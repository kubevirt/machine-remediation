#!/usr/bin/env bash

set -e

source $(dirname "$0")/../common.sh
source $(dirname "$0")/../config.sh
source ${REPO_DIR}/cluster-up/cluster/${KUBEVIRT_PROVIDER}/provider.sh

# delete all operator related objects under the namespace
_kubectl -n ${namespace} delete deployment -l machineremediation.kubevirt.io
_kubectl -n ${namespace} delete rs -l machineremediation.kubevirt.io
_kubectl -n ${namespace} delete pods -l machineremediation.kubevirt.io
_kubectl -n ${namespace} delete serviceaccounts -l machineremediation.kubevirt.io

# delete all operator related objects under the cluster
_kubectl delete clusterrolebinding -l machineremediation.kubevirt.io
_kubectl delete clusterroles -l machineremediation.kubevirt.io

# delete MR CRD
_kubectl delete crd machineremediations.machineremediation.kubevirt.io
