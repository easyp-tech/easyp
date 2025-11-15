//go:build builtin_plugins
// +build builtin_plugins

package plugin

import _ "embed"

//go:embed wasm/protoc-gen-cpp.wasm
var protocGenCpp []byte

//go:embed wasm/protoc-gen-csharp.wasm
var protocGenCsharp []byte

//go:embed wasm/protoc-gen-java.wasm
var protocGenJava []byte

//go:embed wasm/protoc-gen-kotlin.wasm
var protocGenKotlin []byte

//go:embed wasm/protoc-gen-objc.wasm
var protocGenObjc []byte

//go:embed wasm/protoc-gen-php.wasm
var protocGenPhp []byte

//go:embed wasm/protoc-gen-pyi.wasm
var protocGenPyi []byte

//go:embed wasm/protoc-gen-python.wasm
var protocGenPython []byte

//go:embed wasm/protoc-gen-ruby.wasm
var protocGenRuby []byte

//go:embed wasm/protoc-gen-rust.wasm
var protocGenRust []byte

//go:embed wasm/protoc-gen-upb.wasm
var protocGenUPB []byte

//go:embed wasm/protoc-gen-upb_minitable.wasm
var protocGenUPBMinitable []byte

//go:embed wasm/protoc-gen-upbdefs.wasm
var protocGenUPBDefs []byte

//go:embed wasm/memory.wasm
var wasmMemory []byte
