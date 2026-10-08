package modules

import (
	"bytes"
	"crypto/rand"
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"

	"github.com/easyp-tech/easyp/internal/sourceview"
)

type resolvedFileChange struct {
	name string
	data []byte
	mode os.FileMode
}

// resolvedFilesTransaction extends bounded observations with staged sources
// and metadata. Multiple renames are not crash atomic. State rechecks detect
// concurrent changes before commit, but cannot exclude a noncooperating writer
// racing the final check and rename.
type resolvedFilesTransaction struct {
	*resolvedFileObservations
	changes        map[string]resolvedFileChange
	validateInputs func() error
	beforeCommit   func() error
	rename         func(*os.Root, string, string) error
	stage          func(*os.Root, string, []byte, os.FileMode) error
}

func newResolvedFilesTransaction(directory string) (*resolvedFilesTransaction, error) {
	observations, err := newResolvedFileObservations(directory)
	if err != nil {
		return nil, fmt.Errorf("newResolvedFileObservations: %w", err)
	}
	return &resolvedFilesTransaction{
		resolvedFileObservations: observations,
		changes:                  make(map[string]resolvedFileChange), stage: stageResolvedFile,
		rename: func(root *os.Root, before, after string) error { return root.Rename(before, after) },
	}, nil
}

func (tx *resolvedFilesTransaction) plan(name string, data []byte, defaultMode os.FileMode) error {
	before, exists := tx.expected[name]
	if !exists {
		return fmt.Errorf("destination %q was not captured during planning", name)
	}
	if before.exists && bytes.Equal(before.data, data) {
		return nil
	}
	mode := defaultMode
	if before.exists {
		mode = before.resolution.Info.Mode()
	}
	target := filepath.FromSlash(before.resolution.Path)
	if existing, duplicate := tx.changes[target]; duplicate && !bytes.Equal(existing.data, data) {
		return fmt.Errorf("conflicting writes to physical destination %q", target)
	}
	tx.changes[target] = resolvedFileChange{name: target, data: bytes.Clone(data), mode: mode}
	return nil
}

func (tx *resolvedFilesTransaction) verify() error {
	err := tx.verifyCapturedInputs()
	if err != nil {
		return fmt.Errorf("verifyCapturedInputs: %w", err)
	}
	if tx.validateInputs != nil {
		err = tx.validateInputs()
		if err != nil {
			return fmt.Errorf("validateInputs: %w", err)
		}
		// A cache verifier may acquire or repair its own files. Recheck every
		// observed input so that a repaired race cannot be hidden by success.
		err = tx.verifyCapturedInputs()
		if err != nil {
			return fmt.Errorf("verifyCapturedInputs: %w", err)
		}
	}
	return nil
}

type stagedResolvedFile struct {
	change    resolvedFileChange
	output    string
	backup    string
	before    resolvedFileState
	installed resolvedFileState
}

func (tx *resolvedFilesTransaction) apply() (resultErr error) {
	err := tx.verify()
	if err != nil {
		return fmt.Errorf("verify: %w", err)
	}
	if len(tx.changes) == 0 {
		return nil
	}
	temporary := ".easyp-resolved-" + rand.Text()
	err = tx.root.Mkdir(temporary, 0o700)
	if err != nil {
		return fmt.Errorf("Mkdir: %w", err)
	}
	keepRecovery := false
	defer func() {
		if !keepRecovery {
			err := tx.root.RemoveAll(temporary)
			if err != nil {
				resultErr = errors.Join(resultErr, fmt.Errorf("RemoveAll: %w", err))
			}
		}
	}()
	var staged []stagedResolvedFile
	for _, name := range slices.Sorted(maps.Keys(tx.changes)) {
		change := tx.changes[name]
		before, err := tx.readState(name, nil)
		if err != nil {
			return fmt.Errorf("readState: %w", err)
		}
		item := stagedResolvedFile{change: change, before: before,
			output: filepath.Join(temporary, fmt.Sprintf("output-%d", len(staged))),
			backup: filepath.Join(temporary, fmt.Sprintf("original-%d", len(staged))),
		}
		err = tx.stage(tx.root, item.output, change.data, change.mode)
		if err != nil {
			return fmt.Errorf("stage: %w", err)
		}
		if before.exists {
			err = tx.stage(tx.root, item.backup, before.data, before.resolution.Info.Mode())
			if err != nil {
				return fmt.Errorf("stage: %w", err)
			}
		}
		item.installed, err = tx.readState(item.output, nil)
		if err != nil {
			return fmt.Errorf("readState: %w", err)
		}
		item.installed.name, item.installed.resolution.Path = change.name, filepath.ToSlash(change.name)
		staged = append(staged, item)
	}
	if tx.beforeCommit != nil {
		err = tx.beforeCommit()
		if err != nil {
			return fmt.Errorf("beforeCommit: %w", err)
		}
	}
	err = tx.verify()
	if err != nil {
		return fmt.Errorf("verify: %w", err)
	}
	for index, item := range staged {
		current, checkErr := tx.readState(item.change.name, nil)
		if checkErr == nil && !sameResolvedFile(item.before, current) {
			checkErr = fmt.Errorf("file %q changed before commit: %w", item.change.name, sourceview.ErrChanged)
		}
		err = checkErr
		if err == nil {
			if item.before.exists {
				err = tx.rename(tx.root, item.output, item.change.name)
			} else {
				// Link refuses a concurrently created destination, including a
				// dangling alias. Staging cleanup removes the extra link.
				err = tx.root.Link(item.output, item.change.name)
			}
		}
		if err == nil {
			continue
		}
		rollbackErr := tx.rollback(staged[:index])
		if rollbackErr != nil {
			keepRecovery = true
			rollbackErr = fmt.Errorf("rollback: %w; recovery copies remain in %q", rollbackErr, filepath.Join(tx.canonicalRoot, temporary))
		}
		return errors.Join(fmt.Errorf("install: %w", err), rollbackErr)
	}
	return nil
}

func (tx *resolvedFilesTransaction) rollback(applied []stagedResolvedFile) error {
	var failures []error
	for index := len(applied) - 1; index >= 0; index-- {
		item := applied[index]
		current, err := tx.readState(item.change.name, nil)
		if err == nil && !sameResolvedFile(item.installed, current) {
			err = fmt.Errorf("file %q changed after commit: %w", item.change.name, sourceview.ErrChanged)
		}
		if err != nil {
			failures = append(failures, fmt.Errorf("readState: %w", err))
			continue
		}
		if item.before.exists {
			err = tx.rename(tx.root, item.backup, item.change.name)
		} else {
			err = tx.root.Remove(item.change.name)
		}
		if err != nil {
			failures = append(failures, fmt.Errorf("restore: %w", err))
		}
	}
	return errors.Join(failures...)
}

func stageResolvedFile(root *os.Root, name string, data []byte, mode os.FileMode) (resultErr error) {
	file, err := root.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return fmt.Errorf("OpenFile: %w", err)
	}
	defer func() {
		err := file.Close()
		if err != nil {
			resultErr = errors.Join(resultErr, fmt.Errorf("Close: %w", err))
		}
	}()
	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("Write: %w", err)
	}
	err = file.Chmod(mode)
	if err != nil {
		return fmt.Errorf("Chmod: %w", err)
	}
	err = file.Sync()
	if err != nil {
		return fmt.Errorf("Sync: %w", err)
	}
	return nil
}
