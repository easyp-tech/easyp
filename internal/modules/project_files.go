package modules

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	disk "github.com/easyp-tech/easyp/internal/fs/fs"
)

// ReadManifest returns parsed metadata and the original bytes for comment-preserving edits.
func ReadManifest(root string) ([]byte, v1.Module, error) {
	original, err := os.ReadFile(filepath.Join(root, v1.ModuleFile))
	if err != nil {
		return nil, v1.Module{}, fmt.Errorf("ReadFile: %w", err)
	}
	module, err := v1.ParseModule(bytes.NewReader(original))
	if err != nil {
		return nil, v1.Module{}, fmt.Errorf("ParseModule: %w", err)
	}
	return original, module, nil
}

// ReadModuleOrDefault treats a directory without protobuf.mod as one local
// source root without dependencies or a module identity.
func ReadModuleOrDefault(root string) (v1.Module, error) {
	_, module, err := ReadManifest(root)
	if errors.Is(err, os.ErrNotExist) {
		return v1.Module{Roots: []string{"."}}, nil
	}
	if err != nil {
		return v1.Module{}, fmt.Errorf("ReadManifest: %w", err)
	}
	return module, nil
}

func writeV1ResolvedFiles(root string, original, updated []byte, lock v1.Lock) error {
	changed := !bytes.Equal(original, updated)
	if changed {
		if err := writeV1Manifest(root, updated); err != nil {
			return err
		}
	}
	if err := writeV1Lock(root, lock); err != nil {
		if changed {
			if rollbackErr := writeV1Manifest(root, original); rollbackErr != nil {
				return errors.Join(err, fmt.Errorf("restore protobuf.mod: %w", rollbackErr))
			}
		}
		return err
	}
	return nil
}

func writeV1Lock(root string, lock v1.Lock) error {
	raw, err := yaml.Marshal(lock)
	if err != nil {
		return fmt.Errorf("Marshal: %w", err)
	}

	raw = append([]byte("# protobuf.lock - GENERATED FILE, DO NOT EDIT MANUALLY\n"), raw...)
	if err := disk.WriteAtomicFile(filepath.Join(root, v1.LockFile), raw, 0o600); err != nil {
		return fmt.Errorf("WriteAtomicFile: %w", err)
	}
	return nil
}
