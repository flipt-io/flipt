workspace(name = "flipt")

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

http_archive(
    name = "build_stack_rules_proto",
    urls = ["https://github.com/stackb/rules_proto/archive/d86ca6bc56b1589677ec59abfa0bed784d6b7767.tar.gz"],
    sha256 = "36f11f56f6eb48a81eb6850f4fb6c3b4680e3fc2d3ceb9240430e28d32c47009",
    strip_prefix = "rules_proto-d86ca6bc56b1589677ec59abfa0bed784d6b7767",
)

## Go

load("@build_stack_rules_proto//go:deps.bzl", "go_grpc_compile")

go_grpc_compile()

load("@io_bazel_rules_go//go:def.bzl", "go_register_toolchains", "go_rules_dependencies")

go_rules_dependencies()

go_register_toolchains()

## Java

load("@build_stack_rules_proto//java:deps.bzl", "java_grpc_compile")

java_grpc_compile()

## Ruby

load("@build_stack_rules_proto//ruby:deps.bzl", "ruby_grpc_compile")

ruby_grpc_compile()

load("@com_github_grpc_grpc//bazel:grpc_deps.bzl", "grpc_deps")

grpc_deps()

## Python

load("@build_stack_rules_proto//python:deps.bzl", "python_proto_compile")

python_proto_compile()
