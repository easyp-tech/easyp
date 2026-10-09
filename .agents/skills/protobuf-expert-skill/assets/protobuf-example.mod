// EasyP v1 module manifest — copy as protobuf.mod and edit the identity.
// Then run `easyp mod tidy` to resolve versions and write protobuf.lock.
module github.com/acme/contracts

roots (
	proto
)

require (
	github.com/googleapis/googleapis
	github.com/grpc-ecosystem/grpc-gateway
	github.com/bufbuild/protovalidate
)
