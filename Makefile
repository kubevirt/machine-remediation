.PHONY: bazel-build-images
bazel-build:
	./hack/dockerized "./hack/bazel/build.sh"

.PHONY: bazel-generate
bazel-generate:
	SYNC_VENDOR=true ./hack/dockerized "./hack/bazel/generate.sh"

.PHONY: bazel-push-images
bazel-push-images:
	./hack/dockerized "CONTAINER_PREFIX=${CONTAINER_PREFIX} CONTAINER_TAG=${CONTAINER_TAG} ./hack/bazel/push-images.sh"

.PHONY: bazel-tests
bazel-tests:
	hack/dockerized "bazel test \
		--platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 \
		--workspace_status_command=./hack/print-workspace-status.sh \
        --test_output=errors -- //pkg/... //tools/..."

.PHONY: deps-update
deps-update:
	SYNC_VENDOR=true ./hack/dockerized "./hack/deps-update.sh"

.PHONY: distclean
distclean:
	hack/dockerized "rm -rf vendor/ && rm -f go.sum && GO111MODULE=on go clean -modcache"
	rm -rf vendor/


.PHONY: fmt
fmt:
	./hack/dockerized "./hack/bazel/fmt.sh"

.PHONY: generate
generate:
	./hack/dockerized "./hack/generate.sh"

.PHONY: generate-manifests
generate-manifests: generate
	./hack/dockerized "./hack/generate-manifests.sh"
