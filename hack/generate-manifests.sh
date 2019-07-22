#!/usr/bin/env bash

set -e

source hack/common.sh
source hack/config.sh

manifest_container_prefix=${manifest_container_prefix-${container_prefix}}

rm -rf ${MANIFESTS_OUT_DIR}

(cd ${REPO_DIR}/tools/manifest-templator/ && go install)

# first process file includes only
args=$(cd ${REPO_DIR}/manifests && find . -type f -name "*.yaml.in" -not -path "./generated/*")
for arg in $args; do
    infile=${REPO_DIR}/manifests/${arg}
    outfile=${REPO_DIR}/manifests/${arg}.tmp

    manifest-templator \
        --process-files \
        --generated-manifests-dir=${REPO_DIR}/manifests/generated/ \
        --input-file=${infile} >${outfile}
done

# then process variables
args=$(cd ${REPO_DIR}/manifests && find . -type f -name "*.yaml.in.tmp")
for arg in $args; do
    infile=${REPO_DIR}/manifests/${arg}
    final_out_dir=$(dirname ${MANIFESTS_OUT_DIR}/${arg})
    mkdir -p ${final_out_dir}

    manifest=$(basename -s .in.tmp ${arg})
    outfile=${final_out_dir}/${manifest}

    manifest-templator \
        --process-vars \
        --namespace=${namespace} \
        --container-prefix=${manifest_container_prefix} \
        --container-tag=${container_tag} \
        --image-pull-policy=${image_pull_policy} \
        --verbosity=${verbosity} \
        --input-file=${infile} >${outfile}
done

# Remove tmp files
(cd ${REPO_DIR}/manifests && find . -type f -name "*.yaml.in.tmp" -exec rm {} \;)

# Remove empty lines at the end of files which are added by go templating
find ${MANIFESTS_OUT_DIR}/ -type f -exec sed -i {} -e '${/^$/d;}' \;
# Remove empty lines between resources
find ${MANIFESTS_OUT_DIR}/ -type f -exec sed -i '/^$/d' {} \;
