package gitmodules

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"

	"golang.org/x/mod/sumdb/dirhash"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

const cacheVerificationVersion = 2

type cacheVerificationStamp struct {
	Version     int    `json:"version"`
	Hash        string `json:"hash"`
	Fingerprint string `json:"fingerprint"`
	Roots       string `json:"roots,omitempty"`
}

func verifyInstalledV1Module(installed string, entry v1.LockedModule, writeStamp bool) error {
	files, err := snapshotV1Files(installed)
	if err != nil {
		return fmt.Errorf("snapshotV1Files: %w", err)
	}
	for _, name := range files {
		if path.Ext(name) != ".proto" && !snapshotConfigFile(name) {
			return fmt.Errorf("unrelated snapshot file %q; regenerate the unreleased v1 lock with easyp mod tidy", name)
		}
	}
	fingerprint, fast, err := cacheTreeFingerprint(installed)
	if err != nil {
		return fmt.Errorf("cacheTreeFingerprint: %w", err)
	}
	if fast {
		stamp, readErr := readCacheVerificationStamp(installed)
		if readErr == nil && stamp.Version == cacheVerificationVersion && stamp.Hash == entry.Hash && stamp.Roots == v1RootSelectionKey(entry.Roots) && stamp.Fingerprint == fingerprint {
			return nil
		}
	}

	actual, err := dirhash.HashDir(installed, "", dirhash.Hash1)
	if err != nil {
		return fmt.Errorf("HashDir: %w", err)
	}
	if actual != entry.Hash {
		return fmt.Errorf("cached %s@%s hash mismatch: got %s, want %s; inspect or quarantine only cache directory %q, then rerun easyp mod download; keep protobuf.lock unchanged", entry.Source, entry.Commit, actual, entry.Hash, installed)
	}

	if fast {
		after, afterFast, fingerprintErr := cacheTreeFingerprint(installed)
		if fingerprintErr != nil {
			return fmt.Errorf("cacheTreeFingerprint after hash: %w", fingerprintErr)
		}
		if !afterFast || after != fingerprint {
			return fmt.Errorf("cached %s@%s changed while its lock hash was being verified; retry after concurrent cache writes stop", entry.Source, entry.Commit)
		}
		if writeStamp {
			_ = writeCacheVerificationStamp(installed, cacheVerificationStamp{
				Version:     cacheVerificationVersion,
				Hash:        entry.Hash,
				Fingerprint: after,
				Roots:       v1RootSelectionKey(entry.Roots),
			})
		}
	}
	return nil
}

func cacheVerificationStampPath(installed string) string {
	return installed + ".verified.json"
}

func readCacheVerificationStamp(installed string) (cacheVerificationStamp, error) {
	raw, err := os.ReadFile(cacheVerificationStampPath(installed))
	if err != nil {
		return cacheVerificationStamp{}, err
	}
	var stamp cacheVerificationStamp
	if err := json.Unmarshal(raw, &stamp); err != nil {
		return cacheVerificationStamp{}, fmt.Errorf("Unmarshal: %w", err)
	}
	return stamp, nil
}

func writeCacheVerificationStamp(installed string, stamp cacheVerificationStamp) error {
	raw, err := json.Marshal(stamp)
	if err != nil {
		return fmt.Errorf("Marshal: %w", err)
	}
	path := cacheVerificationStampPath(installed)
	temp, err := os.CreateTemp(filepath.Dir(path), ".verify-*")
	if err != nil {
		return fmt.Errorf("CreateTemp: %w", err)
	}
	tempPath := temp.Name()
	defer func() { _ = os.Remove(tempPath) }()
	if err := temp.Chmod(0o600); err != nil {
		_ = temp.Close()
		return fmt.Errorf("Chmod: %w", err)
	}
	if _, err := temp.Write(append(raw, '\n')); err != nil {
		_ = temp.Close()
		return fmt.Errorf("Write: %w", err)
	}
	if err := temp.Close(); err != nil {
		return fmt.Errorf("Close: %w", err)
	}
	if err := os.Rename(tempPath, path); err != nil {
		return fmt.Errorf("Rename: %w", err)
	}
	return nil
}

func cacheTreeFingerprint(root string) (string, bool, error) {
	hash := sha256.New()
	supported := true
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		info, err := entry.Info()
		if err != nil {
			return fmt.Errorf("Info: %w", err)
		}
		if path != root && !info.IsDir() && !info.Mode().IsRegular() {
			return fmt.Errorf("unsupported non-regular cached path %q", path)
		}
		identity, ok := cacheFileIdentity(info)
		if !ok {
			supported = false
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return fmt.Errorf("Rel: %w", err)
		}
		if relative == "." {
			relative = ""
		}
		name := filepath.ToSlash(relative)
		if err := binary.Write(hash, binary.LittleEndian, uint32(len(name))); err != nil {
			return err
		}
		if _, err := hash.Write([]byte(name)); err != nil {
			return err
		}
		for _, value := range []uint64{
			uint64(info.Mode()),
			uint64(info.Size()),
			identity.Device,
			identity.Inode,
			identity.Links,
			uint64(identity.ChangeSeconds),
			uint64(identity.ChangeNanoseconds),
		} {
			if err := binary.Write(hash, binary.LittleEndian, value); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return "", false, err
	}
	return "m1:" + base64.StdEncoding.EncodeToString(hash.Sum(nil)), supported, nil
}
