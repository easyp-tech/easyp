// Package core contains every logic for working cli.
package core

import (
	"errors"
	"log/slog"

	"github.com/easyp-tech/easyp/internal/adapters/console"
	"github.com/easyp-tech/easyp/internal/adapters/plugin"
)

// Core provide to business logic of EasyP.
type Core struct {
	rules        []Rule
	ignore       []string
	deps         []string
	ignoreOnly   map[string][]string
	logger       *slog.Logger
	plugins      []Plugin
	inputs       Inputs
	console      console.Console
	storage      Storage
	moduleConfig ModuleConfig
	lockFile     LockFile

	breakingCheckConfig     BreakingCheckConfig
	currentProjectGitWalker CurrentProjectGitWalker

	localExecutor  plugin.Executor
	remoteExecutor plugin.Executor
}

var (
	ErrInvalidRule            = errors.New("invalid rule")
	ErrRepositoryDoesNotExist = errors.New("repository does not exist")
	ErrEmptyInputFiles        = errors.New("empty input files")
)

func New(
	rules []Rule,
	ignore []string,
	deps []string,
	ignoreOnly map[string][]string,
	logger *slog.Logger,
	plugins []Plugin,
	inputs Inputs,
	console console.Console,
	storage Storage,
	moduleConfig ModuleConfig,
	lockFile LockFile,
	currentProjectGitWalker CurrentProjectGitWalker,
	breakingCheckConfig BreakingCheckConfig,
) *Core {
	return &Core{
		rules:                   rules,
		ignore:                  ignore,
		deps:                    deps,
		ignoreOnly:              ignoreOnly,
		logger:                  logger,
		plugins:                 plugins,
		inputs:                  inputs,
		console:                 console,
		storage:                 storage,
		moduleConfig:            moduleConfig,
		lockFile:                lockFile,
		currentProjectGitWalker: currentProjectGitWalker,
		breakingCheckConfig:     breakingCheckConfig,
		localExecutor:           plugin.NewLocalPluginExecutor(console, logger),
		remoteExecutor:          plugin.NewRemotePluginExecutor(logger),
	}
}
