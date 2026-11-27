package generation

import (
	"context"
	"fmt"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/logger"
	"github.com/easyp-tech/easyp/internal/modules"
)

func generateV1ModuleWithRoots(ctx context.Context, log logger.Logger, request Request, configPath, moduleDir string, gen v1.Generate, module v1.Module, importRoots modules.SourceRoots) error {
	app, index, err := prepareV1ModuleCore(log, request, configPath, moduleDir, gen, module, importRoots, nil)
	if err != nil {
		return fmt.Errorf("prepareV1ModuleCore: %w", err)
	}
	if _, err := index.wait(); err != nil {
		return fmt.Errorf("prepareV1ModuleCore: %w", err)
	}
	if err := app.Generate(ctx, moduleDir, request.DescriptorSetOut, request.IncludeImports); err != nil {
		return fmt.Errorf("Generate: %w", err)
	}
	return nil
}

// prepareV1ModuleCore builds the module's generation core. Module ownership is
// indexed in the background unless managed mode needs it up front; callers
// wait on the index before trusting the core's results, and its error wins.
func prepareV1ModuleCore(log logger.Logger, request Request, configPath, moduleDir string, gen v1.Generate, module v1.Module, importRoots, sources modules.SourceRoots) (*core.Core, *fileModuleIndex, error) {
	if sources == nil {
		var err error
		sources, err = modules.ModuleSources(moduleDir, module)
		if err != nil {
			return nil, nil, fmt.Errorf("ModuleSources: %w", err)
		}
	}
	allSources := append(append(modules.SourceRoots(nil), sources...), importRoots...)
	// The consumer retains ownership when the same physical root is also
	// present in its dependency graph. Indexing also validates every import.
	index := startFileModuleIndex(append(append(modules.SourceRoots(nil), importRoots...), sources...))
	options, err := prepareV1GeneratorConfig(configPath, moduleDir, gen, module)
	if err != nil {
		if _, indexErr := index.wait(); indexErr != nil {
			return nil, nil, indexErr
		}
		return nil, nil, fmt.Errorf("prepareV1GeneratorConfig: %w", err)
	}
	allowed := allSources.FileAllowed()
	options.Logger = log
	options.PluginWorkDir = request.WorkDir
	options.ImportRoots = importRoots.Paths()
	options.ImportFileAllowed = allowed
	options.OpenSourceFile = allSources.SourceOpenerIn(request.session, allowed)
	options.SourceSession = request.session
	options.WalkSourceFiles = sources.WalkSnapshotFiles
	if options.ManagedModeConfig.Enabled || options.ManagedModeConfig.GoPackageOnly {
		owners, err := index.wait()
		if err != nil {
			return nil, nil, err
		}
		options.FileModules = owners
	}
	return core.New(options), index, nil
}

// fileModuleIndex runs IndexFileModules, which also rejects import
// collisions, alongside compilation.
type fileModuleIndex struct {
	done   chan struct{}
	owners map[string]string
	err    error
}

func startFileModuleIndex(roots modules.SourceRoots) *fileModuleIndex {
	index := &fileModuleIndex{done: make(chan struct{})}
	go func() {
		defer close(index.done)
		index.owners, index.err = roots.IndexFileModules()
	}()
	return index
}

// wait returns the module owner of each indexed import path.
func (index *fileModuleIndex) wait() (map[string]string, error) {
	<-index.done
	if index.err != nil {
		return nil, fmt.Errorf("IndexFileModules: %w", index.err)
	}
	return index.owners, nil
}
