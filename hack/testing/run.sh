#!/usr/bin/env bash

set -e

source $(dirname "$0")/../common.sh
source $(dirname "$0")/../config.sh

${TESTS_OUT_DIR}/tests.test -kubeconfig=${kubeconfig} -container-tag=${docker_tag} -container-prefix=${functest_docker_prefix}
