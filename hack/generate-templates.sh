#!/usr/bin/env bash

set -e

source $(dirname "$0")/common.sh

(cd ${REPO_DIR}/tools/resource-generator/ && go install)
resource-generator --type=machine-remediation-operator --namespace={{.Namespace}} --repository={{.ContainerPrefix}} --version={{.ContainerTag}} --pullPolicy={{.ImagePullPolicy}} --verbosity={{.Verbosity}} >${REPO_DIR}/manifests/generated/machine-remediation-operator.yaml.in

#rm -rf cluster-up
#curl -L https://github.com/kubevirt/kubevirtci/archive/${kubevirtci_git_hash}/kubevirtci.tar.gz | tar xz kubevirtci-${kubevirtci_git_hash}/cluster-up --strip-component 1
