package moduleconfig

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

const bufLockFile = "buf.lock"

type bufDependencyLock struct {
	Version string                `yaml:"version"`
	Deps    []bufLockedDependency `yaml:"deps"`
}

type bufLockedDependency struct {
	Name       string `yaml:"name"`
	Remote     string `yaml:"remote"`
	Owner      string `yaml:"owner"`
	Repository string `yaml:"repository"`
	Commit     string `yaml:"commit"`
	Digest     string `yaml:"digest"`
}

func readBufBSRDependencies(configPath string, declarations []string) ([]v1.BSRDependency, error) {
	var dependencies []v1.BSRDependency
	declared := make(map[string]string, len(declarations))
	for _, raw := range declarations {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			return nil, fmt.Errorf("%s: deps entry is empty", configPath)
		}
		module, reference, hasReference := strings.Cut(raw, ":")
		if hasReference && reference == "" {
			return nil, fmt.Errorf("%s: empty BSR reference for %s", configPath, module)
		}
		dependency := v1.BSRDependency{Module: module, Reference: reference, Config: bufModuleConfigFile}
		if err := dependency.Validate(); err != nil {
			return nil, fmt.Errorf("Validate: %s: %w", configPath, err)
		}
		if previous, ok := declared[module]; ok {
			if previous != reference {
				return nil, fmt.Errorf("%s: conflicting BSR references for %s: %q and %q", configPath, module, previous, reference)
			}
			continue
		}
		declared[module] = reference
		dependencies = append(dependencies, dependency)
	}
	lockPath := filepath.Join(filepath.Dir(configPath), bufLockFile)
	raw, err := readGitDependencyConfig(lockPath)
	if errors.Is(err, os.ErrNotExist) {
		return dependencies, nil
	}
	if err != nil {
		return nil, fmt.Errorf("readGitDependencyConfig: %w", err)
	}
	pins, err := parseBufDependencyLock(raw)
	if err != nil {
		return nil, fmt.Errorf("parseBufDependencyLock: %s: %w", lockPath, err)
	}
	if len(declarations) == 0 && len(pins) > 0 {
		return nil, fmt.Errorf("%s has dependencies but %s declares none", lockPath, configPath)
	}
	byModule := make(map[string]v1.BSRDependency, len(pins))
	for _, pin := range pins {
		byModule[pin.Module] = pin
	}
	for i, dependency := range dependencies {
		pin, ok := byModule[dependency.Module]
		if !ok {
			return nil, fmt.Errorf("%s: BSR dependency %s is missing from buf.lock", configPath, dependency.Module)
		}
		dependency.Commit, dependency.Digest = pin.Commit, pin.Digest
		if err := dependency.Validate(); err != nil {
			return nil, fmt.Errorf("Validate: %s: %w", configPath, err)
		}
		dependencies[i] = dependency
	}
	// buf.lock also contains transitive modules that have no declaration in buf.yaml.
	for _, pin := range pins {
		if _, ok := declared[pin.Module]; !ok {
			dependencies = append(dependencies, pin)
		}
	}
	return dependencies, nil
}

func parseBufDependencyLock(raw []byte) ([]v1.BSRDependency, error) {
	var lock bufDependencyLock
	decoder := yaml.NewDecoder(bytes.NewReader(raw))
	if err := decoder.Decode(&lock); err != nil {
		return nil, fmt.Errorf("Decode: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err != nil {
			return nil, fmt.Errorf("Decode: %w", err)
		}
		return nil, fmt.Errorf("buf.lock must contain one YAML document")
	}
	if lock.Version != "v1" && lock.Version != "v1beta1" && lock.Version != "v2" {
		return nil, fmt.Errorf("unsupported buf.lock version %q", lock.Version)
	}
	var pins []v1.BSRDependency
	seen := make(map[string]bool, len(lock.Deps))
	for _, entry := range lock.Deps {
		module := entry.Name
		if lock.Version != "v2" {
			module = entry.Remote + "/" + entry.Owner + "/" + entry.Repository
		}
		if entry.Commit == "" {
			return nil, fmt.Errorf("missing BSR commit for %s", module)
		}
		pin := v1.BSRDependency{Module: module, Commit: entry.Commit, Digest: entry.Digest, Config: bufModuleConfigFile}
		if err := pin.Validate(); err != nil {
			return nil, fmt.Errorf("Validate: %w", err)
		}
		if seen[module] {
			return nil, fmt.Errorf("duplicate BSR pin for %s", module)
		}
		seen[module] = true
		pins = append(pins, pin)
	}
	return pins, nil
}
