#!/bin/bash

set -e

source hack/common.sh
source cluster/$CLUSTER_PROVIDER/provider.sh

echo "Cleaning up ..."

_kubectl -n kube-system delete ds -l 'kubevirt.io=noderecovery'
_kubectl -n kube-system delete pods -l 'kubevirt.io=noderecovery'
_kubectl -n kube-system delete clusterrolebinding -l 'kubevirt.io=noderecovery'
_kubectl -n kube-system delete clusterroles -l 'kubevirt.io=noderecovery'
_kubectl -n kube-system delete serviceaccounts -l 'kubevirt.io=noderecovery'
_kubectl -n kube-system delete configmaps -l 'kubevirt.io=noderecovery'
_kubectl -n kube-system delete customresourcedefinitions -l 'kubevirt.io=noderecovery'
