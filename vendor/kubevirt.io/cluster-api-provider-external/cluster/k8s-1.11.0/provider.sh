#!/bin/bash

set -e

image="k8s-1.11.0@sha256:3412f158ecad53543c9b0aa8468db84dd043f01832a66f0db90327b7dc36a8e8"

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
