#!/bin/bash

set -e

source $(dirname "$0")/../hack/common.sh

source ${REPO_DIR}/cluster/$CLUSTER_PROVIDER/provider.sh

_kubectl "$@"
