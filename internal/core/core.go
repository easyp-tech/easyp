// Package core contains every logic for working cli.
package core

import (
	"errors"
	"log/slog"

	"github.com/easyp-tech/easyp/internal/shared/module_reflect"
)

// Core provide to business logic of EasyP.
type Core struct {
	rules         []Rule
	ignore        []string
	deps          []string
	moduleReflect *modulereflect.ModuleReflect
	ignoreOnly    map[string][]string
	logger        *slog.Logger
	plugins       []Plugin
	inputs        Inputs
	console       Console
	storage       Storage
	moduleConfig  ModuleConfig
	lockFile      LockFile
}

var (
	ErrInvalidRule = errors.New("invalid rule")
)

func New(
	rules []Rule,
	ignore []string,
	deps []string,
	moduleReflect *modulereflect.ModuleReflect,
	ignoreOnly map[string][]string,
	logger *slog.Logger,
	plugins []Plugin,
	inputs Inputs,
	console Console,
	storage Storage,
	moduleConfig ModuleConfig,
	lockFile LockFile,
) *Core {
	return &Core{
		rules:         rules,
		ignore:        ignore,
		deps:          deps,
		moduleReflect: moduleReflect,
		ignoreOnly:    ignoreOnly,
		logger:        logger,
		plugins:       plugins,
		inputs:        inputs,
		console:       console,
		storage:       storage,
		moduleConfig:  moduleConfig,
		lockFile:      lockFile,
	}
}
