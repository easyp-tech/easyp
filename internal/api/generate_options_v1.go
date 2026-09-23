package api

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

// inheritV1GenerateOptions reads only options from ancestor generator files.
// The nearest explicit value wins; generate and plugins belong to the consumer.
func inheritV1GenerateOptions(repoRoot, configPath string, gen *v1.Generate) error {
	configDir := filepath.Dir(configPath)
	rel, err := filepath.Rel(repoRoot, configDir)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return fmt.Errorf("generator %s is outside repository %s", configPath, repoRoot)
	}
	if configDir == repoRoot {
		return nil
	}
	for dir := filepath.Dir(configDir); ; dir = filepath.Dir(dir) {
		if dir == filepath.Dir(dir) {
			break
		}
		parentPath := filepath.Join(dir, "easyp.gen.yaml")
		fp, err := os.Open(parentPath)
		if err == nil {
			parent, parseErr := v1.ParseGenerate(fp)
			_ = fp.Close()
			if parseErr != nil {
				return fmt.Errorf("%s: %w", parentPath, parseErr)
			}
			if gen.Options.Go.PackagePrefix == nil {
				gen.Options.Go.PackagePrefix = parent.Options.Go.PackagePrefix
				if parent.Options.Go.PackagePrefix != nil {
					gen.InheritedGoPackagePrefix = true
				}
			}
		} else if !os.IsNotExist(err) {
			return err
		}
		if dir == repoRoot {
			break
		}
	}
	return nil
}
