// Package core contains every logic for working cli.
package core

import (
	"errors"
	"log/slog"
)

// Core provide to business logic of EasyP.
type Core struct {
	rules                   []Rule
	ignore                  []string
	deps                    []string
	ignoreOnly              map[string][]string
	logger                  *slog.Logger
	plugins                 []Plugin
	inputs                  Inputs
	console                 Console
	storage                 Storage
	moduleConfig            ModuleConfig
	lockFile                LockFile
	currentProjectGitWalker CurrentProjectGitWalker
}

var (
	ErrInvalidRule = errors.New("invalid rule")
)

func New(
	rules []Rule,
	ignore []string,
	deps []string,
	ignoreOnly map[string][]string,
	logger *slog.Logger,
	plugins []Plugin,
	inputs Inputs,
	console Console,
	storage Storage,
	moduleConfig ModuleConfig,
	lockFile LockFile,
	currentProjectGitWalker CurrentProjectGitWalker,
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
	}
}
