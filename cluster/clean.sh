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

    current_time=0
    sample=10
    timeout=120
    echo "Waiting for noderecovery namespace to dissappear ..."
    while  [ "$(_kubectl get ns | grep noderecovery)" ]; do
        sleep $sample
        current_time=$((current_time + sample))
        if [ $current_time -gt $timeout ]; then
            exit 1
        fi
    done
fi
