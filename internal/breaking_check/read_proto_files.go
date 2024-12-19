package breakingcheck

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/yoheimuta/go-protoparser/v4"
	"github.com/yoheimuta/go-protoparser/v4/interpret/unordered"

	"github.com/easyp-tech/easyp/internal/lint"
)

// read current proto files
// current version
// path: dir from should read proto files
// in buf example: buf breaking --against '.git#branch=master' --path no_deps
func ReadCurrentProtoFiles(ctx context.Context, path string) ([]lint.ProtoInfo, error) {
	protoFiles := make([]lint.ProtoInfo, 0)

	err := filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
		switch {
		case err != nil:
			return err
		case ctx.Err() != nil:
			return ctx.Err()
		case filepath.Ext(path) != ".proto":
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("disk.Open: %w", err)
		}
		defer func() {
			err := f.Close()
			if err != nil {
				// TODO: Handle error
			}
		}()

		proto, err := readProtoFile(f)
		if err != nil {
			return fmt.Errorf("readProtoFile: %w", err)
		}

		protoInfo := lint.ProtoInfo{
			Path: path,
			Info: proto,
			//ProtoFilesFromImport: protoFilesFromImport, // skip imports
		}
		protoFiles = append(protoFiles, protoInfo)

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("filepath.WalkDir: %w", err)
	}

	return protoFiles, nil
}

// the same as internal/lint/cmd.go:readProtoFile
func readProtoFile(f io.Reader) (*unordered.Proto, error) {
	got, err := protoparser.Parse(f)
	if err != nil {
		return nil, fmt.Errorf("protoparser.Parse: %w", err)
	}

	proto, err := unordered.InterpretProto(got)
	if err != nil {
		return nil, fmt.Errorf("unordered.InterpretProto: %w", err)
	}

	return proto, nil
}

// read like against buf parametr
// gitRef: branch name
// path full path where have to search proto files
func ReadAgainstProtoFiles(ctx context.Context, gitRef string, rootDir, path string) ([]lint.ProtoInfo, error) {
	gitRepo, err := GetGitRepository(rootDir)
	if err != nil {
		return nil, fmt.Errorf("GetGitRepository: %w", err)
	}

	refName := plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", gitRef))
	refAgainst, err := gitRepo.Reference(refName, false)
	if err != nil {
		return nil, fmt.Errorf("gitRepo.Reference: %w", err)
	}
	commitAgainst, err := gitRepo.CommitObject(refAgainst.Hash())
	if err != nil {
		return nil, fmt.Errorf("gitRepo.CommitObject: %w", err)
	}
	treeAgainst, err := commitAgainst.Tree()
	if err != nil {
		return nil, fmt.Errorf("commitAgainst.Tree: %w", err)
	}

	protoFiles := make([]lint.ProtoInfo, 0)
	err = treeAgainst.Files().ForEach(func(f *object.File) error {
		switch {
		case err != nil:
			return err
		case ctx.Err() != nil:
			return ctx.Err()
		case filepath.Ext(f.Name) != ".proto":
			return nil
		case !isTargetFile(path, f.Name):
			return nil
		}

		reader, err := f.Blob.Reader()
		if err != nil {
			return fmt.Errorf("f.Blob.Reader: %w", err)
		}
		proto, err := readProtoFile(reader)
		if err != nil {
			return fmt.Errorf("readProtoFile: %w", err)
		}
		protoInfo := lint.ProtoInfo{
			Path: filepath.Join(rootDir, f.Name),
			Info: proto,
			//ProtoFilesFromImport: protoFilesFromImport, // skip imports
		}
		protoFiles = append(protoFiles, protoInfo)

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("treeAgainst.Files().ForEach: %w", err)
	}

	return protoFiles, nil
}

func GetGitRepository(path string) (*git.Repository, error) {
	repository, err := git.PlainOpen(path)
	if err != nil {
		return nil, fmt.Errorf("git.PlainOpen: %w", err)
	}
	return repository, nil
}

// isTargetFile check if passed filePath is target
// it has to be in targetPath dir
func isTargetFile(targetPath, filePath string) bool {
	rel, err := filepath.Rel(targetPath, filePath)
	if err != nil {
		return false
	}
	if !filepath.IsLocal(rel) {
		return false
	}

	return true
}
