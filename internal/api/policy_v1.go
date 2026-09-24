package api

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

type v1PolicyFile struct {
	policy   v1.Policy
	sections map[string]yaml.Node
}

func readV1PolicyFile(path string) (v1PolicyFile, bool, error) {
	raw, found, err := readOptionalFile(path)
	if err != nil {
		return v1PolicyFile{}, false, fmt.Errorf("readOptionalFile: %w", err)
	}
	if !found {
		return v1PolicyFile{}, false, nil
	}
	policy, err := v1.ParsePolicy(bytes.NewReader(raw))
	if err != nil {
		return v1PolicyFile{}, true, fmt.Errorf("ParsePolicy: %w", &os.PathError{Op: "parse", Path: path, Err: err})
	}
	var sections map[string]yaml.Node
	if err := yaml.Unmarshal(raw, &sections); err != nil {
		return v1PolicyFile{}, true, fmt.Errorf("Unmarshal: %w", err)
	}
	return v1PolicyFile{policy: policy, sections: sections}, true, nil
}

func (file v1PolicyFile) has(section string) bool {
	_, exists := file.sections[section]
	return exists
}

func v1PolicyPath(directory, projectRoot, configPath string) string {
	if directory == projectRoot {
		return configPath
	}
	return filepath.Join(directory, v1.PolicyFile)
}

// ancestorDirs includes start and stops at root or the filesystem root.
func ancestorDirs(start, root string) []string {
	var directories []string
	for directory := start; ; directory = filepath.Dir(directory) {
		directories = append(directories, directory)
		if directory == root || directory == filepath.Dir(directory) {
			return directories
		}
	}
}
