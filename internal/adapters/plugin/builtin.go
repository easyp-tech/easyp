package plugin

type builtinPlugin string

const (
	// Protobuf base plugins
	builtinPluginCpp    builtinPlugin = "cpp"
	builtinPluginCsharp builtinPlugin = "csharp"
	builtinPluginJava   builtinPlugin = "java"
	builtinPluginKotlin builtinPlugin = "kotlin"
	builtinPluginObjc   builtinPlugin = "objc"
	builtinPluginPhp    builtinPlugin = "php"
	builtinPluginPython builtinPlugin = "python"
	builtinPluginRuby   builtinPlugin = "ruby"

	// gRPC plugins
	builtinPluginGrpcCpp        builtinPlugin = "grpc_cpp"
	builtinPluginGrpcCsharp     builtinPlugin = "grpc_csharp"
	builtinPluginGrpcObjectiveC builtinPlugin = "grpc_objc"
	builtinPluginGrpcPhp        builtinPlugin = "grpc_php"
	builtinPluginGrpcPython     builtinPlugin = "grpc_python"
	builtinPluginGrpcRuby       builtinPlugin = "grpc_ruby"
	builtinPluginGrpcJava       builtinPlugin = "grpc_java"
)

var builtinPlugins = map[builtinPlugin]bool{
	// Protobuf base plugins
	builtinPluginCpp:    true,
	builtinPluginCsharp: true,
	builtinPluginJava:   true,
	builtinPluginKotlin: true,
	builtinPluginObjc:   true,
	builtinPluginPhp:    true,
	builtinPluginPython: true,
	builtinPluginRuby:   true,

	// gRPC plugins
	builtinPluginGrpcCpp:        true,
	builtinPluginGrpcCsharp:     true,
	builtinPluginGrpcObjectiveC: true,
	builtinPluginGrpcPhp:        true,
	builtinPluginGrpcPython:     true,
	builtinPluginGrpcRuby:       true,
	builtinPluginGrpcJava:       true,
}

// IsBuiltinPlugin checks if the plugin is builtin (supported via go-protobuf-gen-builtins)
func IsBuiltinPlugin(pluginName string) bool {

	return builtinPlugins[builtinPlugin(pluginName)]
}
