bazel-generate:
	SYNC_VENDOR=true hack/dockerized "bazel run :gazelle"

bazel-generate-manifests-dev:
	SYNC_MANIFESTS=true hack/dockerized "bazel build //manifests:generate_manifests --define dev=true"

bazel-generate-manifests-release:
	SYNC_MANIFESTS=true hack/dockerized "bazel build //manifests:generate_manifests --define release=true"

bazel-push-images-k8s-1.10.4:
	hack/dockerized "bazel run //:push_images --define dev=true --define cluster_provider=k8s_1_10_4"

bazel-push-images-os-3.10.0:
	hack/dockerized "bazel run //:push_images --define dev=true --define cluster_provider=os_3_10_0"

cluster-up:
	./cluster/up.sh

cluster-down:
	./cluster/down.sh

deps-install:
	SYNC_VENDOR=true hack/dockerized "dep ensure"
	hack/dep-prune.sh

deps-update:
	SYNC_VENDOR=true hack/dockerized "dep ensure -update"
	hack/dep-prune.sh

distclean: clean
	hack/dockerized "rm -rf vendor/ && rm -f Gopkg.lock"
	rm -rf vendor/

.PHONY: bazel-generate bazel-generate-manifests-dev bazel-generate-manifests-release bazel-push-images-k8s-1.10.4 bazel-push-images-os-3.10.0 cluster-up cluster-down deps-install deps-update distclean
