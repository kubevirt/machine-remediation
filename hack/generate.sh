#!/usr/bin/env bash

set -e

source $(dirname "$0")/common.sh

find ${REPO_DIR}/pkg/ -name "*generated*.go" -exec rm {} -f \;

(cd ${REPO_DIR}/tools/resource-generator/ && go build)
rm -f ${REPO_DIR}/manifests/generated/*
${REPO_DIR}/tools/resource-generator/resource-generator --type=machine-health-check-operator --namespace={{.Namespace}} --repository={{.ContainerPrefix}} --version={{.ContainerTag}} --pullPolicy={{.ImagePullPolicy}} --verbosity={{.Verbosity}} >${REPO_DIR}/manifests/generated/machine-health-check-operator.yaml.in

#rm -rf cluster-up
#curl -L https://github.com/kubevirt/kubevirtci/archive/${kubevirtci_git_hash}/kubevirtci.tar.gz | tar xz kubevirtci-${kubevirtci_git_hash}/cluster-up --strip-component 1
