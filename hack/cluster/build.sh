#!/usr/bin/env bash

set -e

source $(dirname "$0")/../common.sh
source $(dirname "$0")/../config.sh
source ${REPO_DIR}/cluster-up/cluster/${KUBEVIRT_PROVIDER}/provider.sh

echo "Building ..."

docker_tag=devel

# Build everyting and publish it
${REPO_DIR}/hack/dockerized "CONTAINER_PREFIX=${docker_prefix} CONTAINER_TAG=${docker_tag} ./hack/bazel/push-images.sh"
${REPO_DIR}/hack/dockerized "MR_IMAGE=${manifest_docker_prefix}/machine-remediation:${docker_tag} CONTAINER_TAG=${docker_tag} IMAGE_PULL_POLICY=${IMAGE_PULL_POLICY} VERBOSITY=${VERBOSITY} ./hack/generate/manifests.sh"

# Make sure that all nodes use the newest images
images=""
for image in ${container_images}; do
    name=$(basename $image)
    images="${images} ${manifest_docker_prefix}/${name}:${docker_tag}"
done

nodes=("master-0" "worker-0")
for node in ${nodes[@]}; do
    ${REPO_DIR}/cluster-up/ssh.sh ${node} "echo \"${images}\" | xargs \-\-max-args=1 sudo podman pull"
done

echo "Done"
