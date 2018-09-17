#!/bin/bash

set -e

_cli="docker run --privileged --net=host --rm ${USE_TTY} -v /var/run/docker.sock:/var/run/docker.sock kubevirtci/gocli@sha256:df958c060ca8d90701a1b592400b33852029979ad6d5c1d9b79683033704b690"

function _main_ip() {
    echo 127.0.0.1
}

function _port() {
    ${_cli} ports --prefix $provider_prefix "$@"
}

function _registry_volume() {
    echo ${job_prefix}_registry
}

function _add_common_params() {
    local params="--nodes ${CLUSTER_NUM_NODES} --memory 4096M --cpu 4 --random-ports --background --prefix $provider_prefix --registry-volume $(_registry_volume) kubevirtci/${image} ${CLUSTER_PROVIDER_EXTRA_ARGS}"
    echo $params
}

function build() {
    # Let's first prune old images, keep the last 5 iterations to improve the cache hit chance
    for arg in ${docker_images}; do
        local name=$(basename $arg)
        images_to_prune="$(docker images --filter "label=${job_prefix}" --filter "label=${name}" --format="{{.ID}} {{.Repository}}:{{.Tag}}" | cat -n | sort -uk2,2 | sort -k1 | tr -s ' ' | grep -v "<none>" | cut -d' ' -f3 | tail -n +6)"
        if [ -n "${images_to_prune}" ]; then
            docker rmi ${images_to_prune}
        fi
    done

    # Build everyting and publish it
    ${REPO_PATH}hack/dockerized "DOCKER_TAG=${DOCKER_TAG} CLUSTER_PROVIDER=${CLUSTER_PROVIDER} ./hack/build-manifests.sh"
    make push

    # Make sure that all nodes use the newest images
    container=""
    container_alias=""
    for arg in ${docker_images}; do
        local name=$(basename $arg)
        container="${container} ${manifest_docker_prefix}/${name}:${docker_tag}"
        container_alias="${container_alias} ${manifest_docker_prefix}/${name}:${docker_tag} kubevirt/${name}:${docker_tag}"
    done
    for i in $(seq 1 ${KUBEVIRT_NUM_NODES}); do
        ${_cli} ssh --prefix $provider_prefix "node$(printf "%02d" ${i})" "echo \"${container}\" | xargs \-\-max-args=1 sudo docker pull"
        ${_cli} ssh --prefix $provider_prefix "node$(printf "%02d" ${i})" "echo \"${container_alias}\" | xargs \-\-max-args=2 sudo docker tag"
    done
}

function _kubectl() {
    export KUBECONFIG=${REPO_PATH}cluster/$CLUSTER_PROVIDER/.kubeconfig
    ${REPO_PATH}cluster/$CLUSTER_PROVIDER/.kubectl "$@"
}

function down() {
    ${_cli} rm --prefix $provider_prefix
}
