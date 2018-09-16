bazel-generate:
	SYNC_VENDOR=true hack/dockerized "bazel run :gazelle"

bazel-docker-images:
	hack/dockerized "bazel run --define docker_prefix=localhost:5000/kubevirt --define docker_tag=devel :noderecovery_images"

deps-install:
	SYNC_VENDOR=true hack/dockerized "dep ensure"
	hack/dep-prune.sh

deps-update:
	SYNC_VENDOR=true hack/dockerized "dep ensure -update"
	hack/dep-prune.sh

distclean: clean
	hack/dockerized "rm -rf vendor/ && rm -f Gopkg.lock"
	rm -rf vendor/

.PHONY: bazel-generate bazel-docker-images deps-install deps-update distclean
