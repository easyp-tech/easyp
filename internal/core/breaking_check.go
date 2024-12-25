package core

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/yoheimuta/go-protoparser/v4/interpret/unordered"
	"github.com/yoheimuta/go-protoparser/v4/parser"

	"github.com/easyp-tech/easyp/internal/fs/fs"
)

func (c *Core) BreakingCheck(ctx context.Context, workingDir, path, against string) ([]IssueInfo, error) {
	fsWalker := fs.NewFSWalker(os.DirFS(workingDir), path)

	currentProtoFiles, err := c.readProtoFiles(ctx, fsWalker)
	if err != nil {
		return nil, fmt.Errorf("c.readCurrentProtoFiles: %w", err)
	}

	againstFSWalker, err := c.currentProjectGitWalker.GetDirWalker(workingDir, against, path)
	if err != nil {
		return nil, fmt.Errorf("c.currentProjectGitWalker.GetDirWalker: %w", err)
	}
	againstProtoFiles, err := c.readProtoFiles(ctx, againstFSWalker)
	if err != nil {
		return nil, fmt.Errorf("c.readAgainstProtoFiles: %w", err)
	}

	currentProtoData, err := collect(currentProtoFiles)
	if err != nil {
		return nil, fmt.Errorf("collect(current): %w", err)
	}

	againstProtoData, err := collect(againstProtoFiles)
	if err != nil {
		return nil, fmt.Errorf("collect(against): %w", err)
	}

	breakingChecker := &BreakingChecker{
		against: againstProtoData,
		current: currentProtoData,
	}

	return breakingChecker.Check()
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
	collectedProtoFiles := make(map[string]struct{})

	for _, protoInfo := range protoInfos {
		protoFilePath := protoInfo.Path
		pkgName := GetPackageName(protoInfo.Info)

		if _, ok := collectedProtoFiles[protoFilePath]; !ok {
			collectProtoFileInfo(protoData, protoInfo.Info, pkgName, protoFilePath)
			collectedProtoFiles[protoFilePath] = struct{}{}
		}

		// collects from imports
		for importPath, protoFile := range protoInfo.ProtoFilesFromImport {
			protoFilePath := string(importPath)
			if _, ok := collectedProtoFiles[protoFilePath]; ok {
				continue
			}

			pkgName := GetPackageName(protoFile)
			collectProtoFileInfo(protoData, protoFile, pkgName, protoFilePath)
			collectedProtoFiles[protoFilePath] = struct{}{}
		}
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
