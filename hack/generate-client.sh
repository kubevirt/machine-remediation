#!/usr/bin/env bash

set -e

source $(dirname "$0")/common.sh

OUTPUT_PKG="github.com/openshift/machine-remediation-operator/pkg/client"
APIS_PKG="github.com/openshift/machine-remediation-operator/pkg/apis"
APIS_VERSIONS="${APIS_PKG}/machineremediation/v1alpha1"
CODE_GENERATORS_CMD_DIR=${VENDOR_DIR}/k8s.io/code-generator/cmd

(
    # To support running this script from anywhere, we have to first cd into this directory
    # so we can install the tools.
    cd "$(dirname "${0}")"
    go install ${CODE_GENERATORS_CMD_DIR}/{client-gen,deepcopy-gen}
)

echo "Generating deepcopy funcs"
deepcopy-gen --input-dirs ${APIS_VERSIONS} -O zz_generated.deepcopy --bounding-dirs ${APIS_PKG} --go-header-file ${REPO_DIR}/hack/boilerplate.go.txt

echo "Generating clientset for ${APIS_VERSIONS} at ${OUTPUT_PKG}/clientset"
client-gen --clientset-name ${CLIENTSET_NAME_VERSIONED:-versioned} --input-base "" --input ${APIS_VERSIONS} --output-package ${OUTPUT_PKG}/clientset --go-header-file ${REPO_DIR}/hack/boilerplate.go.txt
