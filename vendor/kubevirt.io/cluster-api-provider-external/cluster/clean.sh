#!/bin/bash

set -e

source hack/common.sh
source cluster/$CLUSTER_PROVIDER/provider.sh

echo "Cleaning up ..."

# Remove finalizers from all machines, to not block the cleanup
_kubectl -n clusterapi-external-provider get machines -o=custom-columns=NAME:.metadata.name,FINALIZERS:.metadata.finalizers --no-headers | grep "machine.cluster.k8s.io" | while read p; do
    arr=($p)
    name="${arr[0]}"
    _kubectl -n clusterapi-external-provider patch machine $name --type=json -p '[{ "op": "remove", "path": "/metadata/finalizers" }]'
done

# Remove finalizers from all cluster, to not block the cleanup
_kubectl -n clusterapi-external-provider get clusters -o=custom-columns=NAME:.metadata.name,FINALIZERS:.metadata.finalizers --no-headers | grep "cluster.cluster.k8s.io" | while read p; do
    arr=($p)
    name="${arr[0]}"
    _kubectl -n clusterapi-external-provider patch cluster $name --type=json -p '[{ "op": "remove", "path": "/metadata/finalizers" }]'
done

_kubectl -n clusterapi-external-provider delete deployment -l 'kubevirt.io=clusterapi-external-provider'
_kubectl -n clusterapi-external-provider delete configmap -l 'kubevirt.io=clusterapi-external-provider'
_kubectl -n clusterapi-external-provider delete rs -l 'kubevirt.io=clusterapi-external-provider'
_kubectl -n clusterapi-external-provider delete pods -l 'kubevirt.io=clusterapi-external-provider'
_kubectl -n clusterapi-external-provider delete clusterrolebinding -l 'kubevirt.io=clusterapi-external-provider'
_kubectl -n clusterapi-external-provider delete clusterroles -l 'kubevirt.io=clusterapi-external-provider'
_kubectl -n clusterapi-external-provider delete serviceaccounts -l 'kubevirt.io=clusterapi-external-provider'
_kubectl -n clusterapi-external-provider delete customresourcedefinitions -l 'kubevirt.io=clusterapi-external-provider'

# Remove clusterapi-external-provider namespace
if [ "$(_kubectl get ns | grep clusterapi-external-provider)" ]; then
    echo "Clean clusterapi-external-provider namespace"
    _kubectl delete ns clusterapi-external-provider

    current_time=0
    sample=10
    timeout=120
    echo "Waiting for clusterapi-external-provider namespace to dissappear ..."
    while  [ "$(_kubectl get ns | grep clusterapi-external-provider)" ]; do
        sleep $sample
        current_time=$((current_time + sample))
        if [ $current_time -gt $timeout ]; then
            exit 1
        fi
    done
fi
