container_prefix=${CONTAINER_PREFIX:-index.docker.io/kubevirt}
container_tag=${CONTAINER_TAG:-latest}
image_pull_policy=${IMAGE_PULL_POLICY:-IfNotPresent}
csv_version=${CSV_VERSION:-0.0.0}
csv_previous_version=${CSV_PREVIOUS_VERSION:-0.0.0}
kubevirtci_git_hash="45cdf80da619866cdab36295de25a8aeea7eccbc"
namespace=openshift-machine-api
verbosity=${VERBOSITY:-2}
