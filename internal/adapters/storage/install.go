package storage

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/codeclysm/extract/v3"
	"golang.org/x/mod/sumdb/dirhash"

	"github.com/easyp-tech/easyp/internal/core/models"
)

func (s *Storage) Install(
	ctx context.Context,
	cacheDownloadPaths models.CacheDownloadPaths,
	module models.Module,
	revision models.Revision,
	moduleConfig models.ModuleConfig,
) (models.ModuleHash, error) {
	s.logger.Info(
		ctx,
		"Install package",
		slog.String("package", module.Name),
		slog.String("version", revision.Version),
		slog.String("commit", revision.CommitHash),
	)

	version := sanitizePath(revision.Version)
	installedDirPath := s.GetInstallDir(module.Name, version)

	fp, err := os.Open(cacheDownloadPaths.ArchiveFile)
	if err != nil {
		return "", fmt.Errorf("os.Open: %w", err)
	}
	defer func() { _ = fp.Close() }()

	tempExtractDir, err := os.MkdirTemp("", "easyp-extract-*")
	if err != nil {
		return "", fmt.Errorf("os.MkdirTemp extract: %w", err)
	}
	defer os.RemoveAll(tempExtractDir)

	s.logger.Debug(ctx, "Extracting archive to temp dir", slog.String("tempExtractDir", tempExtractDir))

	if err := extract.Archive(ctx, fp, tempExtractDir, nil); err != nil {
		return "", fmt.Errorf("extract.Archive: %w", err)
	}

	parentDir := filepath.Dir(installedDirPath)
	if err := os.MkdirAll(parentDir, dirPerm); err != nil {
		return "", fmt.Errorf("os.MkdirAll parent: %w", err)
	}

	tempInstallDir, err := os.MkdirTemp(parentDir, "easyp-install-*")
	if err != nil {
		return "", fmt.Errorf("os.MkdirTemp install: %w", err)
	}
	defer os.RemoveAll(tempInstallDir)

	if err := os.Chmod(tempInstallDir, dirPerm); err != nil {
		return "", fmt.Errorf("os.Chmod tempInstallDir: %w", err)
	}

	renamer := getRenamer(moduleConfig)

	if err := buildInstallTree(tempExtractDir, tempInstallDir, renamer); err != nil {
		return "", fmt.Errorf("buildInstallTree: %w", err)
	}

	installedPackageHash, err := dirhash.HashDir(tempInstallDir, "", dirhash.DefaultHash)
	if err != nil {
		return "", fmt.Errorf("dirhash.HashDir: %w", err)
	}

	var backupPath string
	if _, err := os.Stat(installedDirPath); err != nil {
		if !os.IsNotExist(err) {
			return "", fmt.Errorf("os.Stat installed dir: %w", err)
		}
	} else {
		backupPath, err = os.MkdirTemp(parentDir, "easyp-backup-*")
		if err != nil {
			return "", fmt.Errorf("os.MkdirTemp backup: %w", err)
		}
		if err := os.RemoveAll(backupPath); err != nil {
			return "", fmt.Errorf("os.RemoveAll backup: %w", err)
		}
		if err := os.Rename(installedDirPath, backupPath); err != nil {
			return "", fmt.Errorf("os.Rename backup: %w", err)
		}
	}

	if err := os.Rename(tempInstallDir, installedDirPath); err != nil {
		if backupPath != "" {
			_ = os.Rename(backupPath, installedDirPath)
		}
		return "", fmt.Errorf("os.Rename: %w", err)
	}

	if backupPath != "" {
		_ = os.RemoveAll(backupPath)
	}

	return models.ModuleHash(installedPackageHash), nil
}

func buildInstallTree(srcDir, dstDir string, renamer func(string) string) error {
	return filepath.Walk(srcDir, func(srcPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(srcDir, srcPath)
		if err != nil {
			return fmt.Errorf("filepath.Rel: %w", err)
		}

		if relPath == "." {
			return nil
		}

		relPath = filepath.ToSlash(relPath)

		rewrittenRel := renamer(relPath)
		if rewrittenRel == "" {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		dstPath := filepath.Join(dstDir, filepath.FromSlash(rewrittenRel))

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode().Perm()|0700)
		}

		if info.Mode()&os.ModeSymlink != 0 {
			return rewriteSymlink(srcPath, srcDir, dstPath, dstDir, renamer)
		}

		return copyFile(srcPath, dstPath, info.Mode())
	})
}

func rewriteSymlink(srcPath, srcDir, dstPath, dstDir string, renamer func(string) string) error {
	target, err := os.Readlink(srcPath)
	if err != nil {
		return fmt.Errorf("os.Readlink %s: %w", srcPath, err)
	}

	if filepath.IsAbs(target) {
		return fmt.Errorf("absolute symlink target not allowed: %s -> %s", srcPath, target)
	}

	linkDir := filepath.Dir(srcPath)
	resolvedTarget := filepath.Join(linkDir, target)
	resolvedTarget, err = filepath.Abs(resolvedTarget)
	if err != nil {
		return fmt.Errorf("filepath.Abs: %w", err)
	}

	relFromSrcDir, err := filepath.Rel(srcDir, resolvedTarget)
	if err != nil {
		return fmt.Errorf("filepath.Rel resolved: %w", err)
	}

	clean := relFromSrcDir == ".." || strings.HasPrefix(relFromSrcDir, ".."+string(os.PathSeparator))
	if clean {
		return fmt.Errorf("symlink target escapes source tree: %s -> %s", srcPath, target)
	}

	relFromSrcDir = filepath.ToSlash(relFromSrcDir)
	rewrittenTarget := renamer(relFromSrcDir)
	if rewrittenTarget == "" {
		return nil
	}

	dstLinkDir := filepath.Dir(dstPath)
	if err := os.MkdirAll(dstLinkDir, dirPerm); err != nil {
		return fmt.Errorf("os.MkdirAll symlink parent: %w", err)
	}
	newTarget, err := filepath.Rel(dstLinkDir, filepath.Join(dstDir, filepath.FromSlash(rewrittenTarget)))
	if err != nil {
		return fmt.Errorf("filepath.Rel new target: %w", err)
	}

	return os.Symlink(newTarget, dstPath)
}

func copyFile(srcPath, dstPath string, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(dstPath), dirPerm); err != nil {
		return fmt.Errorf("os.MkdirAll: %w", err)
	}

	src, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("os.Open src: %w", err)
	}
	defer src.Close()

	dst, err := os.OpenFile(dstPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return fmt.Errorf("os.OpenFile dst: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("io.Copy: %w", err)
	}

	return nil
}

func getRenamer(moduleConfig models.ModuleConfig) func(string) string {
	return func(file string) string {
		for _, dir := range moduleConfig.Directories {
			dir := dir + "/"

			if strings.HasPrefix(file, dir) {
				return strings.TrimPrefix(file, dir)
			}
		}
		return file
	}
}
