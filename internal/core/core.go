// Package core contains every logic for working cli.
package core

import (
	"log/slog"

	modulereflect "github.com/easyp-tech/easyp/internal/api/shared/module_reflect"
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
}

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
	}
}
