#!/usr/bin/env bash

set -e

source $(dirname "$0")/common.sh

CRDS_GENERATORS_CMD_DIR=${VENDOR_DIR}/sigs.k8s.io/controller-tools/cmd

(
    # To support running this script from anywhere, we have to first cd into this directory
    # so we can install the tools.
    cd "$(dirname "${0}")"
    go install ${CRDS_GENERATORS_CMD_DIR}/controller-gen
)

echo "Generating CRD's"
controller-gen crd --domain kubevirt.io --output-dir=manifests/generated/

# add --- in the head of the file
args=$(cd ${REPO_DIR}/manifests/generated && find . -type f -name "*.yaml")
for arg in $args; do
    file=${REPO_DIR}/manifests/generated/${arg}
    sed -i '1i ---' ${file}
done
