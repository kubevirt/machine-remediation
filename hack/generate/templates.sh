#!/usr/bin/env bash

set -e

source $(dirname "$0")/../common.sh

(cd ${REPO_DIR}/tools/resource-generator/ && go install)

mkdir -p ${GENERATED_MANIFESTS_DIR}
resource-generator \
    --mdb-image={{.ImageMachineDisruptionBudget}} \
    --mhc-image={{.ImageMachineHealthCheck}} \
    --mr-image={{.ImageMachineRemediation}} \
    --mro-image={{.ImageOperator}} \
    --type=machine-remediation-operator \
    --namespace={{.Namespace}} \
    --version={{.OperatorVersion}} \
    --pullPolicy={{.ImagePullPolicy}} \
    --verbosity={{.Verbosity}} \
    >${GENERATED_MANIFESTS_DIR}/machine-remediation-operator.yaml.in
resource-generator \
    --type=machine-remediation-operator-cr \
    --namespace={{.Namespace}} \
    --version={{.OperatorVersion}} \
    --pullPolicy={{.ImagePullPolicy}} \
    --verbosity={{.Verbosity}} \
    >${GENERATED_MANIFESTS_DIR}/machine-remediation-operator-cr.yaml.in

#rm -rf cluster-up
#curl -L https://github.com/kubevirt/kubevirtci/archive/${kubevirtci_git_hash}/kubevirtci.tar.gz | tar xz kubevirtci-${kubevirtci_git_hash}/cluster-up --strip-component 1
