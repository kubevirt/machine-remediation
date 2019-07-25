.PHONY: bazel-build-images
bazel-build: bazel-generate
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
		--cache_test_results=no \
        --test_output=errors -- //pkg/... //tools/utils/..."

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
generate: generate-clean generate-crds generate-client generate-templates generate-manifests bazel-generate

.PHONY: generate-clean
generate-clean:
	./hack/dockerized "./hack/generate/clean.sh"

.PHONY: generate-crds
generate-crds:
	./hack/dockerized "./hack/generate/crds.sh"

.PHONY: generate-client
generate-client:
	./hack/dockerized "./hack/generate/client.sh"

.PHONY: generate-manifests
generate-manifests: generate-templates
	./hack/dockerized "./hack/generate/manifests.sh"

.PHONY: generate-templates
generate-templates:
	./hack/dockerized "./hack/generate/templates.sh"
