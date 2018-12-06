#!/bin/bash

set -ex

source hack/common.sh
source cluster/$CLUSTER_PROVIDER/provider.sh

echo "Deploying ..."

_kubectl create -f ${OUT_DIR}/manifests -R $i

if [[ "$CLUSTER_PROVIDER" =~ os-* ]]; then
    _kubectl adm policy add-scc-to-user privileged -z cluster-api-provider-external -n cluster-api-provider-external
    # Helpful for development. Allows admin to access everything KubeVirt creates in the web console
    _kubectl adm policy add-scc-to-user privileged admin
fi

echo "Done"
