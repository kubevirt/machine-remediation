load("@com_github_atlassian_bazel_tools//goimports:def.bzl", "goimports")

goimports(
    name = "goimports",
    display_diffs = True,
    local = ["github.com"],
    prefix = "kubevirt.io/machine-remediation-operator",
    write = True,
)

load("@io_bazel_rules_go//go:def.bzl", "nogo")

nogo(
    name = "nogo_vet",
    config = "nogo_config.json",
    visibility = ["//visibility:public"],
    # These deps enable the analyses equivalent to running `go vet`.
    # Passing vet = True enables only a tiny subset of these (the ones
    # that are always correct).
    # You can see the what `go vet` does by running `go doc cmd/vet`.
    deps = [
        "@org_golang_x_tools//go/analysis/passes/asmdecl:go_tool_library",
        "@org_golang_x_tools//go/analysis/passes/assign:go_tool_library",
        "@org_golang_x_tools//go/analysis/passes/atomic:go_tool_library",
        "@org_golang_x_tools//go/analysis/passes/bools:go_tool_library",
        "@org_golang_x_tools//go/analysis/passes/buildtag:go_tool_library",
        # Fails on a vendored dependency, disabling for now.
        # "@org_golang_x_tools//go/analysis/passes/cgocall:go_tool_library",
        "@org_golang_x_tools//go/analysis/passes/composite:go_tool_library",
        "@org_golang_x_tools//go/analysis/passes/copylock:go_tool_library",
        "@org_golang_x_tools//go/analysis/passes/httpresponse:go_tool_library",
        "@org_golang_x_tools//go/analysis/passes/loopclosure:go_tool_library",
        "@org_golang_x_tools//go/analysis/passes/lostcancel:go_tool_library",
        "@org_golang_x_tools//go/analysis/passes/nilfunc:go_tool_library",
        "@org_golang_x_tools//go/analysis/passes/printf:go_tool_library",
        "@org_golang_x_tools//go/analysis/passes/shift:go_tool_library",
        "@org_golang_x_tools//go/analysis/passes/stdmethods:go_tool_library",
        "@org_golang_x_tools//go/analysis/passes/structtag:go_tool_library",
        "@org_golang_x_tools//go/analysis/passes/tests:go_tool_library",
        "@org_golang_x_tools//go/analysis/passes/unreachable:go_tool_library",
        "@org_golang_x_tools//go/analysis/passes/unsafeptr:go_tool_library",
        "@org_golang_x_tools//go/analysis/passes/unusedresult:go_tool_library",
    ],
)

load("@bazel_gazelle//:def.bzl", "gazelle")

# gazelle:prefix kubevirt.io/machine-remediation-operator
gazelle(name = "gazelle")

load("@com_github_bazelbuild_buildtools//buildifier:def.bzl", "buildifier")

buildifier(name = "buildifier")

genrule(
    name = "get-version",
    srcs = [],
    outs = [".version"],
    cmd = "grep ^STABLE_BUILD_SCM_REVISION bazel-out/stable-status.txt | cut -d' ' -f2 >$@",
    stamp = 1,
    visibility = ["//visibility:public"],
)

load("@io_bazel_rules_docker//contrib:passwd.bzl", "passwd_entry", "passwd_file")
load("@bazel_tools//tools/build_defs/pkg:pkg.bzl", "pkg_tar")

passwd_entry(
    name = "nonroot-user",
    gid = 65534,
    home = "/home/nonroot-user",
    shell = "/bin/bash",
    uid = 65534,
    username = "nonroot-user",
)

passwd_file(
    name = "passwd",
    entries = [
        ":nonroot-user",
    ],
)

pkg_tar(
    name = "passwd-tar",
    srcs = [":passwd"],
    mode = "0644",
    package_dir = "etc",
    visibility = ["//visibility:public"],
)

load(
    "@io_bazel_rules_docker//container:container.bzl",
    "container_bundle",
    "container_image",
)

container_image(
    name = "passwd-image",
    base = "@fedora//image",
    tars = [":passwd-tar"],
    user = "65534",
    visibility = ["//visibility:public"],
)

load(
    "@io_bazel_rules_docker//repositories:repositories.bzl",
    container_repositories = "repositories",
)

config_setting(
    name = "release",
    values = {"define": "release=true"},
)

container_bundle(
    name = "build-images",
    images = {
        # cmd images
        "$(container_prefix)/machine-disruption-budget:$(container_tag)": "//cmd/machine-disruption-budget:machine-disruption-budget-image",
        "$(container_prefix)/machine-health-check:$(container_tag)": "//cmd/machine-health-check:machine-health-check-image",
        "$(container_prefix)/machine-remediation:$(container_tag)": "//cmd/machine-remediation:machine-remediation-image",
        "$(container_prefix)/machine-remediation-operator:$(container_tag)": "//cmd/machine-remediation-operator:machine-remediation-operator-image",
    },
)

load("@io_bazel_rules_docker//contrib:push-all.bzl", "docker_push")

docker_push(
    name = "push-images",
    bundle = ":build-images",
)
