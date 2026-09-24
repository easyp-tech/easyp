// Package core contains every logic for working cli.
package core

import (
	"errors"
	"maps"
	"slices"

	"github.com/easyp-tech/easyp/internal/adapters/console"
	"github.com/easyp-tech/easyp/internal/adapters/plugin"
	"github.com/easyp-tech/easyp/internal/logger"
)

// Core provide to business logic of EasyP.
type Core struct {
	rules         []Rule
	ignore        []string
	ignoreOnly    map[string][]string
	logger        logger.Logger
	plugins       []Plugin
	pluginWorkDir string
	inputs        Inputs
	importRoots   []string
	fileModules   map[string]string
	managedMode   ManagedModeConfig

	breakingCheckConfig     BreakingCheckConfig
	currentProjectGitWalker CurrentProjectGitWalker

	localExecutor   plugin.Executor
	remoteExecutor  plugin.Executor
	builtinExecutor plugin.Executor
	commandExecutor plugin.Executor
}

var (
	ErrInvalidRule            = errors.New("invalid rule")
	ErrRepositoryDoesNotExist = errors.New("repository does not exist")
	ErrEmptyInputFiles        = errors.New("empty input files")
)

// Options configures the lint, breaking, and generation engines.
type Options struct {
	Rules                   []Rule
	Ignore                  []string
	IgnoreOnly              map[string][]string
	Logger                  logger.Logger
	Plugins                 []Plugin
	PluginWorkDir           string
	Inputs                  Inputs
	ImportRoots             []string
	FileModules             map[string]string
	CurrentProjectGitWalker CurrentProjectGitWalker
	BreakingCheckConfig     BreakingCheckConfig
	ManagedModeConfig       ManagedModeConfig
}

// New creates a Core with the configured engines.
func New(options Options) *Core {
	terminal := console.New()
	return &Core{
		rules:                   options.Rules,
		ignore:                  options.Ignore,
		ignoreOnly:              options.IgnoreOnly,
		logger:                  options.Logger,
		plugins:                 options.Plugins,
		pluginWorkDir:           options.PluginWorkDir,
		inputs:                  options.Inputs,
		importRoots:             slices.Clone(options.ImportRoots),
		fileModules:             maps.Clone(options.FileModules),
		currentProjectGitWalker: options.CurrentProjectGitWalker,
		breakingCheckConfig:     options.BreakingCheckConfig,
		managedMode:             options.ManagedModeConfig,
		localExecutor:           plugin.NewLocalPluginExecutor(options.Logger),
		remoteExecutor:          plugin.NewRemotePluginExecutor(options.Logger),
		builtinExecutor:         plugin.NewBuiltinPluginExecutor(options.Logger),
		commandExecutor:         plugin.NewCommandPluginExecutor(terminal, options.Logger),
	}
}
