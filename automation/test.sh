#!/bin/bash

set -ex

export WORKSPACE="${WORKSPACE:-$PWD}"
export CLUSTER_NUM_NODES=2

if [[ $TARGET =~ openshift-.* ]]; then
    export CLUSTER_PROVIDER="os-3.11.0"
else
    export CLUSTER_PROVIDER="k8s-1.11.0"
fi

kubectl() { cluster/kubectl.sh "$@"; }

# Make sure that the VM is properly shut down on exit
trap '{ make cluster-down; }' EXIT SIGINT SIGTERM SIGSTOP

make cluster-down
make cluster-up

# Wait for nodes to become ready
set +e
kubectl get nodes --no-headers
kubectl_rc=$?
while [ $kubectl_rc -ne 0 ] || [ -n "$(kubectl get nodes --no-headers | grep NotReady)" ]; do
    echo "Waiting for all nodes to become ready ..."
    kubectl get nodes --no-headers
    kubectl_rc=$?
    sleep 10
done
set -e

echo "Nodes are ready:"
kubectl get nodes

make cluster-sync

# OpenShift is running important containers under default namespace
namespaces=(kube-system default noderecovery cluster-api-provider-external)

timeout=300
sample=30

for i in ${namespaces[@]}; do
  # Wait until kubevirt pods are running
  current_time=0
  while [ -n "$(kubectl get pods -n $i --no-headers | grep -v Running)" ]; do
    echo "Waiting for kubevirt pods to enter the Running state ..."
    kubectl get pods -n $i --no-headers | >&2 grep -v Running || true
    sleep $sample

    current_time=$((current_time + sample))
    if [ $current_time -gt $timeout ]; then
      exit 1
    fi
  done

  # Make sure all containers are ready
  current_time=0
  while [ -n "$(kubectl get pods -n $i -o'custom-columns=status:status.containerStatuses[*].ready' --no-headers | grep false)" ]; do
    echo "Waiting for KubeVirt containers to become ready ..."
    kubectl get pods -n $i -o'custom-columns=status:status.containerStatuses[*].ready' --no-headers | grep false || true
    sleep $sample

    current_time=$((current_time + sample))
    if [ $current_time -gt $timeout ]; then
      exit 1
    fi
  done
  kubectl get pods -n $i
done

kubectl version

ginko_params="--ginkgo.noColor"

# Run functional tests
FUNC_TEST_ARGS=$ginko_params make functests-run-devel
