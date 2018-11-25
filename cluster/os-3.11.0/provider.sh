#!/bin/bash

set -e

image="os-3.11.0@sha256:09c667db028e40a3646ba070a0de78c09ba6ccbabf6df4937f064688da0745ee"

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
