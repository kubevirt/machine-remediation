#!/usr/bin/env bash

set -e

source $(dirname "$0")/../common.sh
source $(dirname "$0")/../config.sh
source ${REPO_DIR}/cluster-up/cluster/${KUBEVIRT_PROVIDER}/provider.sh

# delete machine-remediation operator object and CRD
if [[ $(_kubectl get crd machineremediationoperators.machineremediation.kubevirt.io --no-headers) != "" ]]; then
    if  [[ $(_kubectl -n ${namespace} get mro mro --no-headers) != "" ]]; then
        _kubectl -n ${namespace} delete mro mro
        until [[ $(_kubectl -n ${namespace} get mro mro --no-headers) == "" ]]; do
            sleep 5
        done
    fi

    # delete MRO CRD
    _kubectl delete crd machineremediationoperators.machineremediation.kubevirt.io
fi

# delete all operator related objects under the namespace
_kubectl -n ${namespace} delete deployment -l machineremediation.kubevirt.io
_kubectl -n ${namespace} delete rs -l machineremediation.kubevirt.io
_kubectl -n ${namespace} delete pods -l machineremediation.kubevirt.io
_kubectl -n ${namespace} delete serviceaccounts -l machineremediation.kubevirt.io

# delete all operator related objects under the cluster
_kubectl delete clusterrolebinding -l machineremediation.kubevirt.io
_kubectl delete clusterroles -l machineremediation.kubevirt.io
