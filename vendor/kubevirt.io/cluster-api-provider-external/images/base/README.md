### How to update base images

- edit base image
- run `make bazel-base-images-build`
- run `make bazel-base-images-push`
- get image digest via `docker images --digests`
- update `WORKSPACE` file with updated images digest

**NOTE**: Change this images only if you want to use different fedora image or to use specific package version.
