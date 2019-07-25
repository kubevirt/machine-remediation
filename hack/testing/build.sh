#!/usr/bin/env bash

set -e

source $(dirname "$0")/../common.sh

mkdir -p ${TESTS_OUT_DIR}/
ginkgo build ${REPO_DIR}/tests
mv ${REPO_DIR}/tests/tests.test ${TESTS_OUT_DIR}/
