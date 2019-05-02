load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["main.go"],
    importpath = "k8s.io/kubernetes/test/instrumentation",
    visibility = ["//visibility:private"],
    deps = [
        "//third_party/go-srcimporter:go_default_library",
    ],
)

go_binary(
    name = "instrumentation",
    embed = [":go_default_library"],
    visibility = ["//visibility:public"],
)

filegroup(
    name = "package-srcs",
    srcs = glob(["**"]),
    tags = ["automanaged"],
    visibility = ["//visibility:private"],
)

filegroup(
    name = "all-srcs",
    srcs = [":package-srcs"],
    tags = ["automanaged"],
    visibility = ["//visibility:public"],
)

genrule(
    name = "list_stable_metrics",
    srcs = [
        "//pkg/util/metrics:all-srcs",
    ],
    outs = ["stable_metrics.txt"],
    cmd = "./$(location :instrumentation) $(locations //pkg/util/metrics:all-srcs) > $@",
    message = "Listing all stable metrics.",
    tools = [":instrumentation"],
)

sh_test(
    name = "stable_metric_test",
    srcs = ["stable_metric_test.sh"],
    data = [
        "testdata/stable_metrics.txt",
        ":list_stable_metrics",
    ],
)
