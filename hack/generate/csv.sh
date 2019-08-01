#!/usr/bin/env bash

set -e

source $(dirname "$0")/../common.sh

(cd ${REPO_DIR}/tools/csv-generator/ && go install)

mkdir -p ${GENERATED_MANIFESTS_DIR}

fake_csv_previous_version="1111.1111.1111"
fake_csv_version="2222.2222.2222"
csv-generator \
    --namespace={{.Namespace}} \
    --repository={{.ContainerPrefix}} \
    --version={{.ContainerTag}} \
    --pullPolicy={{.ImagePullPolicy}} \
    --verbosity={{.Verbosity}} \
    --csv-version=${fake_csv_version} \
    --csv-previous-version=${fake_csv_previous_version} \
    >${GENERATED_MANIFESTS_DIR}/machine-remediation-operator-csv.yaml.in
sed -i "s/$fake_csv_version/{{.CSVVersion}}/g" ${GENERATED_MANIFESTS_DIR}/machine-remediation-operator-csv.yaml.in
sed -i "s/$fake_csv_previous_version/{{.CSVPreviousVersion}}/g" ${GENERATED_MANIFESTS_DIR}/machine-remediation-operator-csv.yaml.in
