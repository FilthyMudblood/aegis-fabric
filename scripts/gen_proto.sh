#!/usr/bin/env bash
set -euo pipefail

if ! command -v buf >/dev/null 2>&1; then
  echo "error: buf is not installed" >&2
  exit 1
fi

if ! command -v protoc-gen-go >/dev/null 2>&1; then
  echo "error: protoc-gen-go is not installed" >&2
  echo "hint: go install google.golang.org/protobuf/cmd/protoc-gen-go@latest" >&2
  exit 1
fi

if ! command -v protoc-gen-go-grpc >/dev/null 2>&1; then
  echo "error: protoc-gen-go-grpc is not installed" >&2
  echo "hint: go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest" >&2
  exit 1
fi

buf generate
