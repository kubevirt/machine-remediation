#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

OUTPUT_PKG="kubevirt.io/node-recovery/pkg/client"
APIS_PKG="kubevirt.io/node-recovery/pkg/apis"
APIS_VERSIONS="${APIS_PKG}/noderecovery/v1alpha1"

echo "Generating deepcopy funcs"
deepcopy-gen --input-dirs ${APIS_VERSIONS} -O zz_generated.deepcopy --bounding-dirs ${APIS_PKG} --go-header-file hack/boilerplate.go.txt

echo "Generating clientset for ${APIS_VERSIONS} at ${OUTPUT_PKG}/clientset"
client-gen --clientset-name ${CLIENTSET_NAME_VERSIONED:-versioned} --input-base "" --input ${APIS_VERSIONS} --output-package ${OUTPUT_PKG}/clientset --go-header-file hack/boilerplate.go.txt

echo "Generating listers for ${APIS_VERSIONS} at ${OUTPUT_PKG}/listers"
lister-gen --input-dirs ${APIS_VERSIONS} --output-package ${OUTPUT_PKG}/listers --go-header-file hack/boilerplate.go.txt

echo "Generating informers for ${APIS_VERSIONS} at ${OUTPUT_PKG}/informers"
informer-gen --input-dirs ${APIS_VERSIONS} \
  --versioned-clientset-package ${OUTPUT_PKG}/clientset/${CLIENTSET_NAME_VERSIONED:-versioned} \
  --listers-package ${OUTPUT_PKG}/listers \
  --output-package ${OUTPUT_PKG}/informers \
  --go-header-file hack/boilerplate.go.txt
