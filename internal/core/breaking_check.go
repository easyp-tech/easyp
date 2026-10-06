package core

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"path/filepath"
	"slices"

	"github.com/yoheimuta/go-protoparser/v4/interpret/unordered"
	"github.com/yoheimuta/go-protoparser/v4/parser"

	"github.com/easyp-tech/easyp/internal/core/path_helpers"
	"github.com/easyp-tech/easyp/internal/fs/fs"
)

var ErrRootOutsideProject = fmt.Errorf("breaking check root must be inside the git repository")

const (
	BreakingCheckFilesCheck string = "FILE"
)

type BreakingCheckConfig struct {
	// branch name to compare with
	AgainstGitRef string
	// dirs should be ignored
	IgnoreDirs []string

	FilesCheck bool
	// Categories selects descriptor-backed compatibility profiles. Empty keeps legacy behavior.
	Categories []string
	// IgnoreUnstable skips packages whose final component is an unstable version.
	IgnoreUnstable bool
}

func (c *Core) BreakingCheck(ctx context.Context, projectRoot, workingDir, path string) ([]IssueInfo, error) {
	c.logger.Debug(
		ctx, "Paths for breaking check",
		slog.String("projectRoot", projectRoot),
		slog.String("workingDir", workingDir),
		slog.String("path", path),
	)

	current := fs.NewFSWalker(workingDir, path)
	against, err := c.currentProjectGitWalker.GetDirWalker(workingDir, c.breakingCheckConfig.AgainstGitRef, path)
	if err != nil {
		return nil, fmt.Errorf("GetDirWalker: %w", err)
	}
	return c.CompareBreaking(ctx, current, against, c.importRoots)
}

// CompareBreaking compares explicitly scoped inputs with revision-specific imports.
// Neither filesystem can fall back to the other revision's dependency roots.
func (c *Core) CompareBreaking(ctx context.Context, current, against DirWalker, againstImportRoots []string) ([]IssueInfo, error) {
	return c.CompareBreakingWithImports(ctx, current, against, BreakingImports{Paths: againstImportRoots})
}

// BreakingImports binds one revision's import paths to its filter and opener.
type BreakingImports struct {
	Paths       []string
	FileAllowed func(string) bool
	Open        func(string) (io.ReadCloser, error)
}

// CompareBreakingWithImports keeps source access isolated between revisions.
func (c *Core) CompareBreakingWithImports(ctx context.Context, current, against DirWalker, imports BreakingImports) ([]IssueInfo, error) {
	baseline := *c
	baseline.importRoots = imports.Paths
	baseline.importFileAllowed = imports.FileAllowed
	baseline.sourceFileOpen = imports.Open
	if len(c.breakingCheckConfig.Categories) > 0 {
		profiles, err := selectedBreakingProfiles(c.breakingCheckConfig.Categories)
		if err != nil {
			return nil, err
		}
		return c.compareBreakingProfiles(ctx, current, against, &baseline, profiles)
	}
	currentFiles, err := c.readProtoFiles(ctx, current)
	if err != nil {
		return nil, fmt.Errorf("readProtoFiles current: %w", err)
	}
	againstFiles, err := baseline.readProtoFiles(ctx, against)
	if err != nil {
		return nil, fmt.Errorf("readProtoFiles baseline: %w", err)
	}
	currentData, err := collect(currentFiles)
	if err != nil {
		return nil, fmt.Errorf("collect: %w", err)
	}
	againstData, err := collect(againstFiles)
	if err != nil {
		return nil, fmt.Errorf("collect: %w", err)
	}
	if c.breakingCheckConfig.IgnoreUnstable {
		excludeUnstablePackages(currentData)
		excludeUnstablePackages(againstData)
	}
	checker := &BreakingChecker{against: againstData, current: currentData, filesCheck: c.breakingCheckConfig.FilesCheck}
	return checker.Check()
}

func (c *Core) readProtoFiles(ctx context.Context, fsWalker DirWalker) ([]ProtoInfo, error) {
	protoFiles := make([]ProtoInfo, 0)

	err := fsWalker.WalkDir(func(path string, err error) error {
		switch {
		case err != nil:
			return err
		case ctx.Err() != nil:
			return ctx.Err()
		case filepath.Ext(path) != ".proto":
			return nil
		case path_helpers.IsIgnoredPath(path, c.breakingCheckConfig.IgnoreDirs):
			return nil
		}

		protoInfo, err := c.protoInfoRead(ctx, fsWalker, path)
		if err != nil {
			return fmt.Errorf("c.protoInfoRead: %w", err)
		}

		protoFiles = append(protoFiles, protoInfo)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("fs.WalkDir: %w", err)
	}

	return protoFiles, nil
}

func collect(protoInfos []ProtoInfo) (ProtoData, error) {
	protoData := make(ProtoData)
	collectedImports := make(map[string]bool)
	// Dependencies establish the imported context first. Explicit target paths
	// then take precedence over import aliases, regardless of traversal order.
	for _, info := range protoInfos {
		for _, importPath := range slices.Sorted(maps.Keys(info.ProtoFilesFromImport)) {
			name := string(importPath)
			if collectedImports[name] {
				continue
			}
			imported := info.ProtoFilesFromImport[importPath]
			collectProtoFileInfo(protoData, imported, GetPackageName(imported), name)
			collectedImports[name] = true
		}
	}
	collectedTargets := make(map[string]bool)
	for _, info := range protoInfos {
		if collectedTargets[info.Path] {
			continue
		}
		collectProtoFileInfo(protoData, info.Info, GetPackageName(info.Info), info.Path)
		collectedTargets[info.Path] = true
	}

	return protoData, nil
}

func collectProtoFileInfo(
	protoData ProtoData, protoFile *unordered.Proto, packageName PackageName, protoFilePath string,
) {
	collection, ok := protoData[packageName]
	if !ok {
		collection = newCollection()
	}

	readImports(collection, protoFile.ProtoBody.Imports, protoFilePath, packageName)
	readServices(collection, protoFile.ProtoBody.Services, protoFilePath, packageName)
	readMessages(collection, "", protoFile.ProtoBody.Messages, protoFilePath, packageName)
	readEnums(collection, "", protoFile.ProtoBody.Enums, protoFilePath, packageName)

	protoData[packageName] = collection
}

func readImports(collection *Collection, imports []*parser.Import, protoFilePath string, packageName PackageName) {
	for _, imp := range imports {
		collection.Imports[ImportPath(imp.Location)] = Import{
			ProtoFilePath: protoFilePath,
			PackageName:   packageName,
			Import:        imp,
		}
	}
}

func readServices(
	collection *Collection, services []*unordered.Service, protoFilePath string, packageName PackageName,
) {
	for _, service := range services {
		serviceName := service.ServiceName

		collection.Services[serviceName] = Service{
			ProtoFilePath: protoFilePath,
			PackageName:   packageName,
			Service:       service,
		}
	}
}

func readMessages(
	collection *Collection,
	messagePath string,
	messages []*unordered.Message,
	protoFilePath string,
	packageName PackageName,
) {
	for _, message := range messages {
		newMessagePath := getProtoEntityPath(messagePath, message.MessageName)

		msg := Message{
			MessagePath:   newMessagePath,
			ProtoFilePath: protoFilePath,
			PackageName:   packageName,
			Message:       message,
		}
		collection.Messages[newMessagePath] = msg

		readMessages(collection, newMessagePath, message.MessageBody.Messages, protoFilePath, packageName)
		readOneOfs(collection, newMessagePath, message.MessageBody.Oneofs, protoFilePath, packageName)
		readEnums(collection, newMessagePath, message.MessageBody.Enums, protoFilePath, packageName)
	}
}

func readOneOfs(
	collection *Collection, messagePath string, oneOfs []*parser.Oneof, protoFilePath string, packageName PackageName,
) {
	for _, oneOf := range oneOfs {
		newOneOfPath := getProtoEntityPath(messagePath, oneOf.OneofName)

		res := OneOf{
			OneOfPath:     newOneOfPath,
			ProtoFilePath: protoFilePath,
			PackageName:   packageName,
			Oneof:         oneOf,
		}
		collection.OneOfs[newOneOfPath] = res
	}
}

func readEnums(
	collection *Collection, messagePath string, enums []*unordered.Enum, protoFilePath string, packageName PackageName,
) {
	for _, enum := range enums {
		newEnumPath := getProtoEntityPath(messagePath, enum.EnumName)

		res := Enum{
			EnumPath:      newEnumPath,
			ProtoFilePath: protoFilePath,
			PackageName:   packageName,
			Enum:          enum,
		}
		collection.Enums[newEnumPath] = res
	}
}

func getProtoEntityPath(rootPath, name string) string {
	if rootPath == "" {
		return name
	}

	return fmt.Sprintf("%s.%s", rootPath, name)
}

func newCollection() *Collection {
	collection := &Collection{
		Imports:  make(map[ImportPath]Import),
		Services: make(map[string]Service),
		Messages: make(map[string]Message),
		OneOfs:   make(map[string]OneOf),
		Enums:    make(map[string]Enum),
	}
	return collection
}
