#!/usr/bin/env bash

set -ex

source $(dirname "$0")/../common.sh

CRDS_GENERATORS_CMD_DIR=${VENDOR_DIR}/sigs.k8s.io/controller-tools/cmd

(
    # To support running this script from anywhere, we have to first cd into this directory
    # so we can install the tools.
    cd "$(dirname "${0}")"
    go install ${CRDS_GENERATORS_CMD_DIR}/controller-gen
)

echo "Generating CRD's"
mkdir -p ${GENERATED_MANIFESTS_DIR}/crds
controller-gen crd --domain kubevirt.io --output-dir=${GENERATED_MANIFESTS_DIR}/crds

# add --- in the head of the file
args=$(cd ${GENERATED_MANIFESTS_DIR}/crds && find . -type f -name "*.yaml")
for arg in $args; do
    file=${GENERATED_MANIFESTS_DIR}/crds/${arg}
    sed -i '1i ---' ${file}
done
