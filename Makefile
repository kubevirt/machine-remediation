bazel-generate:
	SYNC_VENDOR=true hack/dockerized "bazel run :gazelle"

bazel-generate-manifests-dev:
	SYNC_MANIFESTS=true hack/dockerized "bazel build //manifests:generate_manifests --define dev=true"

bazel-generate-manifests-release:
	SYNC_MANIFESTS=true hack/dockerized "bazel build //manifests:generate_manifests --define release=true --define container_tag=${CONTAINER_TAG}"

bazel-generate-manifests-tests:
	SYNC_MANIFESTS=true hack/dockerized "bazel build //manifests/testing:generate_manifests"

bazel-base-images-build:
	./hack/dockerized "bazel build //images/base:build_images"

bazel-base-images-push:
	./hack/dockerized "bazel build //images/base:push_images"

bazel-push-images-k8s-1.11.0:
	hack/dockerized "bazel run //:push_images --define dev=true --define cluster_provider=k8s_1_11_0"

bazel-push-images-os-3.11.0:
	hack/dockerized "bazel run //:push_images --define dev=true --define cluster_provider=os_3_11_0"

bazel-push-images-release:
	hack/dockerized "bazel run //:push_images --define release=true --define container_tag=${CONTAINER_TAG}"

bazel-tests:
	./hack/dockerized "bazel test --test_output=all -- //pkg/... "

cluster-build:
	./cluster/build.sh

cluster-clean:
	./cluster/clean.sh

cluster-deploy: cluster-clean
	./cluster/deploy.sh

cluster-down:
	./cluster/down.sh

cluster-sync: cluster-build cluster-deploy

cluster-up:
	./cluster/up.sh

deps-install:
	SYNC_VENDOR=true hack/dockerized "dep ensure -v"

deps-update:
	SYNC_VENDOR=true hack/dockerized "dep ensure -v -update"

distclean:
	hack/dockerized "rm -rf vendor/ && rm -f Gopkg.lock"
	rm -rf vendor/

functests-build:
	SYNC_OUT=true hack/dockerized "hack/functests-build.sh"

functests-run-devel: functests-build
	CONTAINERS_PREFIX="registry:5000/kubevirt" CONTAINERS_TAG="devel" hack/functests-run.sh

functests-run-release:
	CONTAINERS_PREFIX="registry:5000/kubevirt" CONTAINERS_TAG="devel" hack/functests-run.sh > .release-functest 2>&1

generate:
	hack/dockerized "hack/update-codegen.sh"

release-announce: functests-run-release
	./hack/release-announce.sh $(RELREF) $(PREREF)

.PHONY: bazel-generate \
	bazel-base-images-build \
	bazel-base-images-push \
	bazel-generate-manifests-dev \
	bazel-generate-manifests-release \
	bazel-generate-manifests-tests \
	bazel-push-images-k8s-1.11.0 \
	bazel-push-images-os-3.11.0 \
	bazel-push-images-release \
	bazel-tests \
	cluster-build \
	cluster-clean \
	cluster-deploy \
	cluster-down \
	cluster-sync \
	cluster-up \
	deps-install \
	deps-update \
	distclean \
	functests-build \
	functests-run-devel \
	functests-run-release \
	generate \
	release-announce
