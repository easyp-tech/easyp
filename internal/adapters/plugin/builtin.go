package plugin

type builtinPlugin string

const (
	builtinPluginCpp          builtinPlugin = "cpp"
	builtinPluginCsharp       builtinPlugin = "csharp"
	builtinPluginJava         builtinPlugin = "java"
	builtinPluginKotlin       builtinPlugin = "kotlin"
	builtinPluginObjc         builtinPlugin = "objc"
	builtinPluginPhp          builtinPlugin = "php"
	builtinPluginPyi          builtinPlugin = "pyi"
	builtinPluginPython       builtinPlugin = "python"
	builtinPluginRuby         builtinPlugin = "ruby"
	builtinPluginRust         builtinPlugin = "rust"
	builtinPluginUpb          builtinPlugin = "upb"
	builtinPluginUpbMinitable builtinPlugin = "upb_minitable"
	builtinPluginUpbDefs      builtinPlugin = "upbdefs"
)

// IsBuiltinPlugin проверяет, является ли плагин базовым (поддерживается через go-protobuf-gen-builtins)
func IsBuiltinPlugin(pluginName string) bool {
	builtinPlugins := map[builtinPlugin]bool{
		builtinPluginCpp:          true,
		builtinPluginCsharp:       true,
		builtinPluginJava:         true,
		builtinPluginKotlin:       true,
		builtinPluginObjc:         true,
		builtinPluginPhp:          true,
		builtinPluginPyi:          true,
		builtinPluginPython:       true,
		builtinPluginRuby:         true,
		builtinPluginRust:         true,
		builtinPluginUpb:          true,
		builtinPluginUpbMinitable: true,
		builtinPluginUpbDefs:      true,
	}

	return builtinPluginExecutorEnabled() && builtinPlugins[builtinPlugin(pluginName)]
}
