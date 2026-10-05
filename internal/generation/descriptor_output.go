package generation

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"

	"github.com/easyp-tech/easyp/internal/fs/fs"
)

type descriptorOutput struct {
	path string
	raw  []byte
	mode os.FileMode
}

func planDescriptorOutputs(request Request, targets []preparedDescriptorTarget) ([]descriptorOutput, error) {
	var outputs []descriptorOutput
	if request.DescriptorSetOut != "" {
		set, err := combineDescriptorTargets(targets, request.IncludeImports)
		if err != nil {
			return nil, err
		}
		output, err := prepareDescriptorOutput(request.DescriptorSetOut, set)
		if err != nil {
			return nil, err
		}
		return []descriptorOutput{output}, nil
	}
	paths := make(map[string]string)
	for _, target := range targets {
		outputPath := descriptorTargetPath(request, target)
		// Resolve existing ancestor symlinks before checking aliases: two projects
		// can otherwise overwrite one export through different directory paths.
		canonical, err := canonicalDescriptorOutputPath(outputPath)
		if err != nil {
			return nil, fmt.Errorf("canonicalDescriptorOutputPath: %w", err)
		}
		// Reject case aliases consistently, including on case-sensitive hosts.
		key := strings.ToLower(canonical)
		if previous, ok := paths[key]; ok {
			return nil, fmt.Errorf("descriptor output collision at %q between %s and %s", outputPath, previous, target.label)
		}
		paths[key] = target.label
		output, err := prepareDescriptorOutput(outputPath, target.plan.DescriptorSet(request.IncludeImports))
		if err != nil {
			return nil, err
		}
		outputs = append(outputs, output)
	}
	for key := range paths {
		for parent := filepath.Dir(key); parent != filepath.Dir(parent); parent = filepath.Dir(parent) {
			if _, exists := paths[parent]; exists {
				return nil, fmt.Errorf("descriptor output path %q is also a parent directory of %q", parent, key)
			}
		}
	}
	slices.SortFunc(outputs, func(a, b descriptorOutput) int { return strings.Compare(a.path, b.path) })
	return outputs, nil
}

func descriptorTargetPath(request Request, target preparedDescriptorTarget) string {
	project := relativeDescriptorPath(request.WorkDir, filepath.Dir(target.target.configPath))
	if !filepath.IsLocal(filepath.FromSlash(project)) {
		project = filepath.Join("_external", descriptorFileStem(filepath.Base(filepath.Dir(target.target.configPath)), project))
	}
	identity := target.selected.module.Name
	name := path.Base(identity)
	if identity == "" {
		directory := relativeDescriptorPath(request.WorkDir, target.selected.directory)
		identity, name = "local:"+directory, path.Base(directory)
		if name == "." {
			name = "root"
		}
	}
	return filepath.Join(request.DescriptorSetOutDir, filepath.FromSlash(project), descriptorFileStem(name, identity)+".pb")
}

// The identity suffix makes names stable even when another selected module has
// the same basename. It does not depend on module order, versions or cache paths.
func descriptorFileStem(name, identity string) string {
	slug := strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '-' || r == '_' || r == '.' {
			return r
		}
		return '-'
	}, strings.ToLower(name))
	slug = strings.Trim(slug, "-.")
	if len(slug) > 48 {
		slug = slug[:48]
	}
	if slug == "" {
		slug = "module"
	}
	digest := sha256.Sum256([]byte(identity))
	return fmt.Sprintf("%s-%x", slug, digest[:6])
}

func prepareDescriptorOutput(path string, set *descriptorpb.FileDescriptorSet) (descriptorOutput, error) {
	raw, err := proto.MarshalOptions{Deterministic: true}.Marshal(set)
	if err != nil {
		return descriptorOutput{}, fmt.Errorf("Marshal: %w", err)
	}
	mode := os.FileMode(0o644)
	info, err := os.Lstat(path)
	if err == nil {
		if !info.Mode().IsRegular() {
			return descriptorOutput{}, fmt.Errorf("descriptor output %q is not a regular file", path)
		}
		mode = info.Mode().Perm()
	} else if !os.IsNotExist(err) {
		return descriptorOutput{}, fmt.Errorf("Lstat: %w", err)
	}
	if _, err := canonicalDescriptorOutputPath(path); err != nil {
		return descriptorOutput{}, fmt.Errorf("canonicalDescriptorOutputPath: %w", err)
	}
	for dir := filepath.Dir(path); ; dir = filepath.Dir(dir) {
		info, err := os.Stat(dir)
		if err == nil {
			if !info.IsDir() {
				return descriptorOutput{}, fmt.Errorf("descriptor output parent %q is not a directory", dir)
			}
			break
		}
		if !os.IsNotExist(err) || filepath.Dir(dir) == dir {
			return descriptorOutput{}, fmt.Errorf("Stat: %w", err)
		}
	}
	return descriptorOutput{path: path, raw: raw, mode: mode}, nil
}

func (output descriptorOutput) write() error {
	if err := os.MkdirAll(filepath.Dir(output.path), 0o755); err != nil {
		return fmt.Errorf("MkdirAll: %w", err)
	}
	if err := fs.WriteAtomicFile(output.path, output.raw, output.mode); err != nil {
		return fmt.Errorf("WriteAtomicFile: %w", err)
	}
	return nil
}

// canonicalDescriptorOutputPath resolves the existing part of an output path.
// Missing descendants are appended without creating directories during preflight.
func canonicalDescriptorOutputPath(outputPath string) (string, error) {
	directory := filepath.Dir(outputPath)
	var missing []string
	for {
		resolved, err := filepath.EvalSymlinks(directory)
		if err == nil {
			for i := len(missing) - 1; i >= 0; i-- {
				resolved = filepath.Join(resolved, missing[i])
			}
			return filepath.Join(resolved, filepath.Base(outputPath)), nil
		}
		if !os.IsNotExist(err) || directory == filepath.Dir(directory) {
			return "", fmt.Errorf("EvalSymlinks: %w", err)
		}
		// An existing symlink with an unresolved target is not a missing
		// directory that MkdirAll can safely create later.
		if _, statErr := os.Lstat(directory); statErr == nil {
			return "", fmt.Errorf("EvalSymlinks: %w", err)
		} else if !os.IsNotExist(statErr) {
			return "", fmt.Errorf("Lstat: %w", statErr)
		}
		missing = append(missing, filepath.Base(directory))
		directory = filepath.Dir(directory)
	}
}
