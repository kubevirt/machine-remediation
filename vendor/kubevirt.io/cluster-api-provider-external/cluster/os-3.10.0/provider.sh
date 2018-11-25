#!/bin/bash

set -e

image="os-3.10.0@sha256:faa467495207af8faa9214b1bf8adabf6161fab7f4da11b63efa41610a3ff0ab"

source cluster/provider-common.sh

function up() {
    ${_cli} run --reverse $(_add_common_params) --registry-port 33001
    ${_cli} ssh --prefix $provider_prefix node01 -- sudo cp /etc/origin/master/admin.kubeconfig ~vagrant/
    ${_cli} ssh --prefix $provider_prefix node01 -- sudo chown vagrant:vagrant ~vagrant/admin.kubeconfig

    # Copy oc tool and configuration file
    ${_cli} scp --prefix $provider_prefix /usr/bin/oc - >${REPO_PATH}cluster/$CLUSTER_PROVIDER/.kubectl
    chmod u+x ${REPO_PATH}cluster/$CLUSTER_PROVIDER/.kubectl
    ${_cli} scp --prefix $provider_prefix /etc/origin/master/admin.kubeconfig - >${REPO_PATH}cluster/$CLUSTER_PROVIDER/.kubeconfig

    # Update Kube config to support unsecured connection
    export KUBECONFIG=${REPO_PATH}cluster/$CLUSTER_PROVIDER/.kubeconfig
    ${REPO_PATH}cluster/$CLUSTER_PROVIDER/.kubectl config set-cluster node01:8443 --server=https://$(_main_ip):$(_port ocp)
    ${REPO_PATH}cluster/$CLUSTER_PROVIDER/.kubectl config set-cluster node01:8443 --insecure-skip-tls-verify=true
}
