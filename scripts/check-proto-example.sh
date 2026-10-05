#!/bin/sh
set -eu
repo=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT HUP INT TERM
cd "$repo"
go build -o "$work/easyp" ./cmd/easyp
cp easyp.yaml easyp.gen.yaml protobuf.mod protobuf.lock "$work/"
mkdir -p "$work/examples"
cp -R examples/proto "$work/examples/proto"
cd "$work"
export EASYPPATH="$work/cache"
./easyp validate-config
./easyp mod tidy
cmp protobuf.lock "$repo/protobuf.lock"
./easyp lint --root examples/proto
./easyp generate --descriptor_set_out first.pb --include_imports
cp -R examples/generated first-generated
./easyp generate --descriptor_set_out second.pb --include_imports
cmp first.pb second.pb
diff -r first-generated examples/generated
test -s examples/generated/python/hello/v1/hello_pb2.py
printf '%s\n' 'Native v1 example: validation, lint, lock and deterministic generation passed.'
