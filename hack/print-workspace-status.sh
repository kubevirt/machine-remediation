#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

export REPO_DIR=$(dirname "${BASH_SOURCE}")/..

source "${REPO_DIR}/hack/version.sh"
version::get_version_vars

# Prefix with STABLE_ so that these values are saved to stable-status.txt
# instead of volatile-status.txt.
# Stamped rules will be retriggered by changes to stable-status.txt, but not by
# changes to volatile-status.txt.
# IMPORTANT: the camelCase vars should match the lists in hack/version.sh
# and pkg/version/def.bzl.
cat <<EOF
STABLE_BUILD_GIT_COMMIT ${GIT_COMMIT-}
STABLE_BUILD_SCM_STATUS ${GIT_TREE_STATE-}
STABLE_BUILD_SCM_REVISION ${GIT_VERSION-}
STABLE_DOCKER_TAG ${GIT_VERSION/+/_}
gitCommit ${GIT_COMMIT-}
gitTreeState ${GIT_TREE_STATE-}
gitVersion ${GIT_VERSION-}
buildDate $(date \
    ${SOURCE_DATE_EPOCH:+"--date=@${SOURCE_DATE_EPOCH}"} \
    -u +'%Y-%m-%dT%H:%M:%SZ')
EOF
