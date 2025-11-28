package plugin

type builtinPlugin string

const (
	// Protobuf base plugins (only for languages with gRPC support)
	builtinPluginCpp    builtinPlugin = "cpp"
	builtinPluginCsharp builtinPlugin = "csharp"
	builtinPluginObjc   builtinPlugin = "objc"
	builtinPluginPhp    builtinPlugin = "php"
	builtinPluginPython builtinPlugin = "python"
	builtinPluginRuby   builtinPlugin = "ruby"

	// gRPC plugins
	builtinPluginGrpcCpp        builtinPlugin = "grpc_cpp"
	builtinPluginGrpcCsharp     builtinPlugin = "grpc_csharp"
	builtinPluginGrpcNode       builtinPlugin = "grpc_node"
	builtinPluginGrpcObjectiveC builtinPlugin = "grpc_objective_c"
	builtinPluginGrpcPhp        builtinPlugin = "grpc_php"
	builtinPluginGrpcPython     builtinPlugin = "grpc_python"
	builtinPluginGrpcRuby       builtinPlugin = "grpc_ruby"
)

var builtinPlugins = map[builtinPlugin]bool{
	// Protobuf base plugins (only for languages with gRPC support)
	builtinPluginCpp:    true,
	builtinPluginCsharp: true,
	builtinPluginObjc:   true,
	builtinPluginPhp:    true,
	builtinPluginPython: true,
	builtinPluginRuby:   true,

	// gRPC plugins
	builtinPluginGrpcCpp:        true,
	builtinPluginGrpcCsharp:     true,
	builtinPluginGrpcNode:       true,
	builtinPluginGrpcObjectiveC: true,
	builtinPluginGrpcPhp:        true,
	builtinPluginGrpcPython:     true,
	builtinPluginGrpcRuby:       true,
}

// IsBuiltinPlugin checks if the plugin is builtin (supported via go-protobuf-gen-builtins)
func IsBuiltinPlugin(pluginName string) bool {

	return builtinPluginExecutorEnabled() && builtinPlugins[builtinPlugin(pluginName)]
}
