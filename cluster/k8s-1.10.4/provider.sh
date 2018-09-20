#!/bin/bash

set -e

image="k8s-1.10.4@sha256:e697f59d5b3b09130a8854b08417959eb70197fcc1b47cf64fe83c5164f00cfa"

source cluster/provider-common.sh

function up() {
    ${_cli} run $(_add_common_params) --registry-port 33000

    # Copy k8s config and kubectl
    ${_cli} scp --prefix $provider_prefix /usr/bin/kubectl - >${REPO_PATH}cluster/$CLUSTER_PROVIDER/.kubectl
    chmod u+x ${REPO_PATH}cluster/$CLUSTER_PROVIDER/.kubectl
    ${_cli} scp --prefix $provider_prefix /etc/kubernetes/admin.conf - >${REPO_PATH}cluster/$CLUSTER_PROVIDER/.kubeconfig

    # Set server and disable tls check
    export KUBECONFIG=${REPO_PATH}cluster/$CLUSTER_PROVIDER/.kubeconfig
    ${REPO_PATH}cluster/$CLUSTER_PROVIDER/.kubectl config set-cluster kubernetes --server=https://$(_main_ip):$(_port k8s)
    ${REPO_PATH}cluster/$CLUSTER_PROVIDER/.kubectl config set-cluster kubernetes --insecure-skip-tls-verify=true
}
