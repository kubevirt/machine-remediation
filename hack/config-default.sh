container_images="cmd/machine-remediation"
container_prefix=${CONTAINER_PREFIX:-index.docker.io/kubevirt}
container_tag=${CONTAINER_TAG:-latest}
mr_image=${MR_IMAGE:-"${container_prefix}/machine-remediation:${container_tag}"}
image_pull_policy=${IMAGE_PULL_POLICY:-IfNotPresent}
kubevirtci_git_hash="5c28400e9b729435a05093f16a01d6e2aaab2432"
namespace=openshift-machine-api
verbosity=${VERBOSITY:-2}
