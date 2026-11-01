package core

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"slices"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"

	pluginexecutor "github.com/easyp-tech/easyp/internal/adapters/plugin"
	"github.com/easyp-tech/easyp/internal/version"
)

// Bound concurrent plugin processes and independent WASM memories as well as CPU work.
const maxPluginWorkers = 4

type pluginResult struct {
	source          string
	filesToGenerate []string
	response        *pluginpb.CodeGeneratorResponse
	err             error
}

// invokePlugin owns its request and response; prepared descriptors remain immutable.
func (p *GenerationPlan) invokePlugin(ctx context.Context, plugin Plugin) pluginResult {
	filesToGenerate := slices.Clone(p.files)
	if plugin.WithImports {
		for _, dependency := range p.dependencyFiles {
			if !slices.Contains(filesToGenerate, dependency) {
				filesToGenerate = append(filesToGenerate, dependency)
			}
		}
	}
	request := &pluginpb.CodeGeneratorRequest{
		FileToGenerate:  filesToGenerate,
		ProtoFile:       p.descriptors,
		CompilerVersion: version.CompilerVersion(),
	}
	request = proto.Clone(request).(*pluginpb.CodeGeneratorRequest)
	source := plugin.Source.Name
	if plugin.Source.Remote != "" {
		source = plugin.Source.Remote
	}
	if plugin.Source.Path != "" {
		source = plugin.Source.Path
	}
	result := pluginResult{source: source, filesToGenerate: filesToGenerate}
	executor := p.core.getExecutor(plugin)
	response, err := executor.Execute(ctx, pluginexecutor.Info{
		Source: source, WorkDir: p.core.pluginWorkDir,
		Command: plugin.Source.Command, Options: plugin.Options,
	}, request)
	if err != nil {
		result.err = fmt.Errorf("execute plugin %s: %w, executor: %s", source, err, executor.GetName())
		return result
	}
	if response == nil {
		result.err = fmt.Errorf("plugin %s returned no response, executor: %s", source, executor.GetName())
		return result
	}
	if response.Error != nil {
		result.err = fmt.Errorf("plugin %s error: %s, executor: %s", plugin.Source, *response.Error, executor.GetName())
		return result
	}
	result.response = response
	return result
}

// applyPluginResponse runs serially in configuration order, including insertion points.
func (p *GenerationPlan) applyPluginResponse(ctx context.Context, plugin Plugin, result pluginResult, bucket *GenerateBucket) error {
	c := p.core
	outputDir := p.root
	if plugin.Out != "" {
		outputDir = filepath.Join(p.root, plugin.Out)
	}
	var goOutputDirs map[string]string
	paths := plugin.Options["paths"]
	if c.goPackageOutputPrefix != "" && len(paths) > 0 && paths[len(paths)-1] == "source_relative" {
		goOutputDirs = goPackageOutputDirs(c.goPackageOutputPrefix, result.filesToGenerate, p.descriptors, plugin.Options)
	}
	for _, file := range result.response.File {
		originalName := file.GetName()
		if !filepath.IsLocal(filepath.FromSlash(originalName)) {
			return fmt.Errorf("plugin returned invalid output path %q", originalName)
		}
		outputName := originalName
		if file.GetInsertionPoint() == "" {
			outputName = goPackageOutputPath(file, goOutputDirs)
		}
		if !filepath.IsLocal(filepath.FromSlash(outputName)) {
			return fmt.Errorf("plugin returned invalid output path %q", outputName)
		}
		path := filepath.Join(outputDir, filepath.FromSlash(outputName))
		originalPath := filepath.Join(outputDir, filepath.FromSlash(originalName))
		if file.GetInsertionPoint() != "" {
			var err error
			path, err = bucket.insertionOutputPath(originalPath)
			if err != nil {
				return fmt.Errorf("insertionOutputPath: %w", err)
			}
		}
		c.logger.Debug(ctx, "generated file",
			slog.String("plugin", result.source),
			slog.String("file", outputName),
			slog.String("plugin_out", plugin.Out),
			slog.String("full_path", path),
		)
		if err := addFileWithInsertionPoint(ctx, path, file, bucket); err != nil {
			return fmt.Errorf("addFileWithInsertionPoint: %w", err)
		}
		if file.GetInsertionPoint() == "" {
			bucket.recordOutputPath(originalPath, path)
		}
	}
	return nil
}
