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
	rules                 []Rule
	ignore                []string
	ignoreOnly            map[string][]string
	allowCommentIgnores   bool
	knownLintRules        []string
	logger                logger.Logger
	plugins               []Plugin
	pluginWorkDir         string
	inputs                Inputs
	importRoots           []string
	importFileAllowed     func(string) bool
	fileModules           map[string]string
	managedMode           ManagedModeConfig
	goPackageOutputPrefix string

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
	Rules               []Rule
	Ignore              []string
	IgnoreOnly          map[string][]string
	AllowCommentIgnores bool
	KnownLintRules      []string
	Logger              logger.Logger
	Plugins             []Plugin
	PluginWorkDir       string
	Inputs              Inputs
	ImportRoots         []string
	// ImportFileAllowed restricts physical imports and module inputs to the
	// files selected by dependency metadata. Nil permits every file.
	ImportFileAllowed       func(string) bool
	FileModules             map[string]string
	CurrentProjectGitWalker CurrentProjectGitWalker
	BreakingCheckConfig     BreakingCheckConfig
	ManagedModeConfig       ManagedModeConfig
	// GoPackageOutputPrefix places source-relative Go output under its effective
	// go_package, relative to this import prefix. Path markers use their stable parent.
	GoPackageOutputPrefix string
}

// New creates a Core with the configured engines.
func New(options Options) *Core {
	terminal := console.New()
	return &Core{
		rules:                   options.Rules,
		ignore:                  options.Ignore,
		ignoreOnly:              options.IgnoreOnly,
		allowCommentIgnores:     options.AllowCommentIgnores,
		knownLintRules:          slices.Clone(options.KnownLintRules),
		logger:                  options.Logger,
		plugins:                 options.Plugins,
		pluginWorkDir:           options.PluginWorkDir,
		inputs:                  options.Inputs,
		importRoots:             slices.Clone(options.ImportRoots),
		importFileAllowed:       options.ImportFileAllowed,
		fileModules:             maps.Clone(options.FileModules),
		currentProjectGitWalker: options.CurrentProjectGitWalker,
		breakingCheckConfig:     options.BreakingCheckConfig,
		managedMode:             options.ManagedModeConfig,
		goPackageOutputPrefix:   options.GoPackageOutputPrefix,
		localExecutor:           plugin.NewLocalPluginExecutor(options.Logger),
		remoteExecutor:          plugin.NewRemotePluginExecutor(options.Logger),
		builtinExecutor:         plugin.NewBuiltinPluginExecutor(options.Logger),
		commandExecutor:         plugin.NewCommandPluginExecutor(terminal, options.Logger),
	}
}
