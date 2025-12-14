package plugin

import _ "embed"

//go:embed wasm/memory.wasm
var wasmMemory []byte

// Universal WASM module containing all plugins (protobuf and gRPC)
// Base protobuf plugins: cpp, csharp, java, kotlin, objc, php, python, ruby
// gRPC plugins: grpc_cpp, grpc_csharp, grpc_node, grpc_objective_c, grpc_php, grpc_python, grpc_ruby
//
//go:embed wasm/protoc_gen_universal.wasm
var protocGenUniversal []byte
