#!/usr/bin/env bash

set -e

source $(dirname "$0")/../common.sh

(cd ${REPO_DIR}/tools/resource-generator/ && go install)

mkdir -p ${GENERATED_MANIFESTS_DIR}
resource-generator \
    --mr-image={{.ImageMachineRemediation}} \
    --type=machine-remediation \
    --namespace={{.Namespace}} \
    --pullPolicy={{.ImagePullPolicy}} \
    --verbosity={{.Verbosity}} \
    >${GENERATED_MANIFESTS_DIR}/machine-remediation.yaml.in
