#!/bin/bash

set -e

source hack/common.sh

mkdir -p ${TESTS_OUT_DIR}/
ginkgo build ${REPO_DIR}/tests
mv ${REPO_DIR}/tests/tests.test ${TESTS_OUT_DIR}/
