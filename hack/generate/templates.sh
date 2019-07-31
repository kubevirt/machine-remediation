#!/usr/bin/env bash

set -e

source $(dirname "$0")/../common.sh

(cd ${REPO_DIR}/tools/resource-generator/ && go install)

mkdir -p ${GENERATED_MANIFESTS_DIR}
resource-generator \
    --type=machine-remediation-operator \
    --namespace={{.Namespace}} \
    --repository={{.ContainerPrefix}} \
    --version={{.ContainerTag}} \
    --pullPolicy={{.ImagePullPolicy}} \
    --verbosity={{.Verbosity}} \
    >${GENERATED_MANIFESTS_DIR}/machine-remediation-operator.yaml.in
resource-generator \
    --type=machine-remediation-operator-cr \
    --namespace={{.Namespace}} \
    --repository={{.ContainerPrefix}} \
    --version={{.ContainerTag}} \
    --pullPolicy={{.ImagePullPolicy}} \
    --verbosity={{.Verbosity}} \
    >${GENERATED_MANIFESTS_DIR}/machine-remediation-operator-cr.yaml.in

# Generate CSV
fake_csv_previous_version="1111.1111.1111"
fake_csv_version="2222.2222.2222"

resource-generator \
    --type=csv \
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

#rm -rf cluster-up
#curl -L https://github.com/kubevirt/kubevirtci/archive/${kubevirtci_git_hash}/kubevirtci.tar.gz | tar xz kubevirtci-${kubevirtci_git_hash}/cluster-up --strip-component 1
