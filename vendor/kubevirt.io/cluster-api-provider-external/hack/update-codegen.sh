#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

APIS_PKG="kubevirt.io/cluster-api-provider-external/pkg/apis"
APIS_VERSIONS="${APIS_PKG}/providerconfig/v1alpha1"

echo "Generating deepcopy funcs"
deepcopy-gen --input-dirs ${APIS_VERSIONS} -O zz_generated.deepcopy --bounding-dirs ${APIS_PKG} --go-header-file hack/boilerplate.go.txt
