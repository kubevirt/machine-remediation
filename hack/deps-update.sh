#!/usr/bin/env bash

GO111MODULE=on go mod tidy
GO111MODULE=on go mod vendor
./hack/bazel/generate.sh
