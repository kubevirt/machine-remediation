#!/bin/bash

set -e

source hack/common.sh
source cluster/$CLUSTER_PROVIDER/provider.sh

echo "Cleaning up ..."

_kubectl -n noderecovery delete ds -l 'kubevirt.io=noderecovery'
_kubectl -n noderecovery delete pods -l 'kubevirt.io=noderecovery'
_kubectl -n noderecovery delete clusterrolebinding -l 'kubevirt.io=noderecovery'
_kubectl -n noderecovery delete clusterroles -l 'kubevirt.io=noderecovery'
_kubectl -n noderecovery delete serviceaccounts -l 'kubevirt.io=noderecovery'
_kubectl -n noderecovery delete configmaps -l 'kubevirt.io=noderecovery'
_kubectl -n noderecovery delete customresourcedefinitions -l 'kubevirt.io=noderecovery'

if [ "$(_kubectl get ns | grep noderecovery)" ]; then
    echo "Clean noderecovery namespace"
    _kubectl delete ns noderecovery
fi
