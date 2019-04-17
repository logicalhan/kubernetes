package(default_visibility = ["//visibility:public"])

load(
    "@io_bazel_rules_go//go:def.bzl",
    "go_library",
    "go_test",
)

go_library(
    name = "go_default_library",
    srcs = [
        "util.go",
        "registry.go",
        "metric.go",
        "opts.go",
        "counter.go",
        "histogram.go",
        "summary.go",
        "gauge.go"
    ],
    importpath = "k8s.io/kubernetes/pkg/util/metrics",
    deps = [
        "//staging/src/k8s.io/client-go/util/flowcontrol:go_default_library",
        "//vendor/github.com/prometheus/client_golang/prometheus:go_default_library",
        "//vendor/github.com/blang/semver:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = [
        "util_test.go",
        "counter_test.go",
        "gauge_test.go",
        "summary_test.go",
        "histogram_test.go",
    ],
    embed = [":go_default_library"],
    deps = [
        "//staging/src/k8s.io/client-go/util/flowcontrol:go_default_library",
        "//vendor/github.com/blang/semver:go_default_library",
    ],
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
)
