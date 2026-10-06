package modules

// WalkProtoFiles visits source files while excluding hidden, vendor and nested-module directories.
func WalkProtoFiles(root string, visit func(string) error) error {
	return (SourceRoot{Path: root}).Walk(visit)
}
