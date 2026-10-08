package modules

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/sourceview"
)

// ReadManifest returns parsed metadata and the original bytes for comment-preserving edits.
func ReadManifest(root string) ([]byte, v1.Module, error) {
	original, err := sourceview.ReadLocal(context.Background(), root, v1.ModuleFile)
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

func writeV1ResolvedFiles(root string, original, updated []byte, lock v1.Lock) (resultErr error) {
	tx, err := newResolvedFilesTransaction(root)
	if err != nil {
		return fmt.Errorf("newResolvedFilesTransaction: %w", err)
	}
	defer func() { resultErr = errors.Join(resultErr, tx.close()) }()
	manifest, err := tx.capture(v1.ModuleFile, nil)
	if err != nil {
		return fmt.Errorf("capture: %w", err)
	}
	if !manifest.exists || !bytes.Equal(manifest.data, original) {
		return fmt.Errorf("protobuf.mod changed since resolution: %w", sourceview.ErrChanged)
	}
	_, err = tx.capture(v1.LockFile, nil)
	if err != nil {
		return fmt.Errorf("capture: %w", err)
	}
	return writeV1ResolvedFilesTransaction(tx, updated, lock)
}

func writeV1ResolvedFilesTransaction(tx *resolvedFilesTransaction, updated []byte, lock v1.Lock) error {
	raw, err := marshalV1Lock(lock)
	if err != nil {
		return fmt.Errorf("marshalV1Lock: %w", err)
	}
	err = tx.plan(v1.ModuleFile, updated, 0o644)
	if err != nil {
		return fmt.Errorf("plan: %w", err)
	}
	err = tx.plan(v1.LockFile, raw, 0o644)
	if err != nil {
		return fmt.Errorf("plan: %w", err)
	}
	err = tx.apply()
	if err != nil {
		return fmt.Errorf("apply: %w", err)
	}
	return nil
}

func marshalV1Lock(lock v1.Lock) ([]byte, error) {
	raw, err := yaml.Marshal(lock)
	if err != nil {
		return nil, fmt.Errorf("Marshal: %w", err)
	}
	raw = append([]byte("# protobuf.lock - GENERATED FILE, DO NOT EDIT MANUALLY\n"), raw...)
	return raw, nil
}
