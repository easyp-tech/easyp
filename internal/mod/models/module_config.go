package models

// ModuleConfig contains module config such as dirs from buf.work.yaml
type ModuleConfig struct {
	// Directories contains dirs with proto files (buf.work.yaml)
	Directories []string
}
