#!/usr/bin/env bash

set -e

source $(dirname "$0")/../common.sh
source $(dirname "$0")/../config.sh

version_file="${REPO_DIR}/hack/kubevirtci/version.txt"
sha_file="${REPO_DIR}/hack/kubevirtci/sha.txt"

# check if we got a new cluster-up git commit hash
if [[ -f "${version_file}" ]] && [[ $(cat ${version_file}) == ${kubevirtci_git_hash} ]]; then
    current_sha=$(find ${REPO_DIR}/cluster-up -type f | sort | xargs sha256sum | sha256sum | awk '{print $1}')
    if [[ -f "${sha_file}" ]] && [[ $(cat ${sha_file}) == ${current_sha} ]]; then
        exit 0
    fi
fi

# download updated cluster-up from kubevirtci
echo "downloading cluster-up"
rm -rf ${REPO_DIR}/cluster-up
curl -L https://github.com/kubevirt/kubevirtci/archive/${kubevirtci_git_hash}/kubevirtci.tar.gz | tar xz kubevirtci-${kubevirtci_git_hash}/cluster-up --strip-component 1

# remove unneeded providers
find ${REPO_DIR}/cluster-up/cluster -maxdepth 1 -mindepth 1 -type d | grep -v okd | xargs rm -rf

echo ${kubevirtci_git_hash} >${version_file}

new_sha=$(find ${REPO_DIR}/cluster-up -type f | sort | xargs sha256sum | sha256sum | awk '{print $1}')
echo ${new_sha} >${sha_file}
