#!/bin/bash

set -e

source hack/common.sh
source cluster/$CLUSTER_PROVIDER/provider.sh

echo "Cleaning up ..."

namespaces=(noderecovery cluster-api-external-provider)
for namespace in ${namespaces[@]}; do
    # Remove finalizers from all machines, to not block the cleanup
    _kubectl -n ${namespace} get machines -o=custom-columns=NAME:.metadata.name,FINALIZERS:.metadata.finalizers --no-headers | grep "machine.cluster.k8s.io" | while read p; do
        arr=($p)
        name="${arr[0]}"
        _kubectl -n ${namespace} delete machine $name
        _kubectl -n ${namespace} patch machine $name --type=json -p '[{ "op": "remove", "path": "/metadata/finalizers" }]'
    done

    # Remove finalizers from all clusters, to not block the cleanup
    _kubectl -n ${namespace} get clusters -o=custom-columns=NAME:.metadata.name,FINALIZERS:.metadata.finalizers --no-headers | grep "cluster.cluster.k8s.io" | while read p; do
        arr=($p)
        name="${arr[0]}"
        _kubectl -n ${namespace} delete cluster $name
        _kubectl -n ${namespace} patch cluster $name --type=json -p '[{ "op": "remove", "path": "/metadata/finalizers" }]'
    done

    _kubectl -n ${namespace} delete ds -l "${namespace}.kubevirt.io"
    _kubectl -n ${namespace} delete pods -l "${namespace}.kubevirt.io"
    _kubectl -n ${namespace} delete clusterrolebinding -l "${namespace}.kubevirt.io"
    _kubectl -n ${namespace} delete clusterroles -l "${namespace}.kubevirt.io"
    _kubectl -n ${namespace} delete serviceaccounts -l "${namespace}.kubevirt.io"
    _kubectl -n ${namespace} delete configmaps -l "${namespace}.kubevirt.io"
    _kubectl -n ${namespace} delete secrets -l "${namespace}.kubevirt.io"
    _kubectl -n ${namespace} delete services -l "${namespace}.kubevirt.io"
    _kubectl -n ${namespace} delete customresourcedefinitions -l "${namespace}.kubevirt.io"

    if [ "$(_kubectl get ns | grep ${namespace})" ]; then
        echo "Clean ${namespace} namespace"
        _kubectl delete ns ${namespace}

        current_time=0
        sample=10
        timeout=120
        echo "Waiting for noderecovery namespace to dissappear ..."
        while  [ "$(_kubectl get ns | grep ${namespace})" ]; do
            sleep $sample
            current_time=$((current_time + sample))
            if [ $current_time -gt $timeout ]; then
                exit 1
            fi
        done
    fi
done
