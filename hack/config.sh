unset container_prefix container_tag namespace image_pull_policy verbosity

source ${REPO_DIR}/hack/config-default.sh
source ${REPO_DIR}/cluster-up/hack/config.sh

export container_prefix container_tag namespace image_pull_policy verbosity
