//go:build builtin_plugins
// +build builtin_plugins

package plugin

import _ "embed"

//go:embed wasm/protoc-gen-cpp.wasm
var protocGenCpp []byte

//go:embed wasm/protoc-gen-csharp.wasm
var protocGenCsharp []byte

//go:embed wasm/protoc-gen-objc.wasm
var protocGenObjc []byte

//go:embed wasm/protoc-gen-php.wasm
var protocGenPhp []byte

//go:embed wasm/protoc-gen-python.wasm
var protocGenPython []byte

//go:embed wasm/protoc-gen-ruby.wasm
var protocGenRuby []byte

//go:embed wasm/memory.wasm
var wasmMemory []byte

// gRPC plugins
//
//go:embed wasm/grpc_cpp_plugin.wasm
var grpcCppPlugin []byte

//go:embed wasm/grpc_csharp_plugin.wasm
var grpcCsharpPlugin []byte

//go:embed wasm/grpc_node_plugin.wasm
var grpcNodePlugin []byte

//go:embed wasm/grpc_objective_c_plugin.wasm
var grpcObjectiveCPlugin []byte

//go:embed wasm/grpc_php_plugin.wasm
var grpcPhpPlugin []byte

//go:embed wasm/grpc_python_plugin.wasm
var grpcPythonPlugin []byte

//go:embed wasm/grpc_ruby_plugin.wasm
var grpcRubyPlugin []byte
