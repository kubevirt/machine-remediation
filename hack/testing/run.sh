#!/usr/bin/env bash

set -e

source $(dirname "$0")/../common.sh
source $(dirname "$0")/../config.sh

${TESTS_OUT_DIR}/tests.test -kubeconfig=${KUBECONFIG} -container-prefix=${container_prefix} -container-tag=${container_tag}
