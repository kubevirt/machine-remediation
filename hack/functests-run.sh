#!/bin/bash

set -e

source hack/common.sh

${TESTS_OUT_DIR}/tests.test \
    --kubeconfig=${CLUSTER_PROVIDER_KUBECONFIG} \
    --container-prefix=${CONTAINERS_PREFIX} \
    --container-tag=${CONTAINERS_TAG} \
    --test.timeout 60m \
    ${FUNC_TEST_ARGS}
