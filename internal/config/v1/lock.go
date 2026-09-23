package v1

// Lock is the v1 protobuf.lock document for one protobuf module.
type Lock struct {
	Version int            `yaml:"version"`
	Modules []LockedModule `yaml:"modules"`
}

type LockedModule struct {
	Source  string `yaml:"source"`
	Version string `yaml:"version"`
	Commit  string `yaml:"commit"`
	Hash    string `yaml:"hash"`
}
