#!/bin/bash

set -e

source hack/common.sh
source cluster/$CLUSTER_PROVIDER/provider.sh

echo "Cleaning up ..."

# Remove finalizers from all machines, to not block the cleanup
_kubectl -n cluster-api-provider-external get machines -o=custom-columns=NAME:.metadata.name,FINALIZERS:.metadata.finalizers --no-headers | grep "machine.cluster.k8s.io" | while read p; do
    arr=($p)
    name="${arr[0]}"
    _kubectl -n cluster-api-provider-external delete machine $name
    _kubectl -n cluster-api-provider-external patch machine $name --type=json -p '[{ "op": "remove", "path": "/metadata/finalizers" }]'
done

# Remove finalizers from all cluster, to not block the cleanup
_kubectl -n cluster-api-provider-external get clusters -o=custom-columns=NAME:.metadata.name,FINALIZERS:.metadata.finalizers --no-headers | grep "cluster.cluster.k8s.io" | while read p; do
    arr=($p)
    name="${arr[0]}"
    _kubectl -n cluster-api-provider-external delete cluster $name
    _kubectl -n cluster-api-provider-external patch cluster $name --type=json -p '[{ "op": "remove", "path": "/metadata/finalizers" }]'
done

_kubectl -n cluster-api-provider-external delete deployment -l 'kubevirt.io=cluster-api-provider-external'
_kubectl -n cluster-api-provider-external delete configmap -l 'kubevirt.io=cluster-api-provider-external'
_kubectl -n cluster-api-provider-external delete rs -l 'kubevirt.io=cluster-api-provider-external'
_kubectl -n cluster-api-provider-external delete pods -l 'kubevirt.io=cluster-api-provider-external'
_kubectl -n cluster-api-provider-external delete clusterrolebinding -l 'kubevirt.io=cluster-api-provider-external'
_kubectl -n cluster-api-provider-external delete clusterroles -l 'kubevirt.io=cluster-api-provider-external'
_kubectl -n cluster-api-provider-external delete serviceaccounts -l 'kubevirt.io=cluster-api-provider-external'
_kubectl -n cluster-api-provider-external delete customresourcedefinitions -l 'kubevirt.io=cluster-api-provider-external'

# Remove cluster-api-provider-external namespace
if [ "$(_kubectl get ns | grep cluster-api-provider-external)" ]; then
    echo "Clean cluster-api-provider-external namespace"
    _kubectl delete ns cluster-api-provider-external

    current_time=0
    sample=10
    timeout=120
    echo "Waiting for cluster-api-provider-external namespace to dissappear ..."
    while  [ "$(_kubectl get ns | grep cluster-api-provider-external)" ]; do
        sleep $sample
        current_time=$((current_time + sample))
        if [ $current_time -gt $timeout ]; then
            exit 1
        fi
    done
fi
