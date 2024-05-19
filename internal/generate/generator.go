// Package generate provides the core functionality of EasyP generate.
package generate

import (
	modulereflect "github.com/easyp-tech/easyp/internal/api/shared/module_reflect"
)

type (
	// Plugin is a plugin for gRPC generator.
	Plugin struct {
		Name    string
		Out     string
		Options map[string]string
	}
	// Config is the configuration for EasyP generate.
	Config struct {
		Deps          []string
		Plugins       []Plugin
		ModuleReflect *modulereflect.ModuleReflect
	}
	// Generator is the core functionality of EasyP generate.
	Generator struct {
		deps          []string
		plugins       []Plugin
		moduleReflect *modulereflect.ModuleReflect
	}
	// Query is a query for making sh command.
	Query struct {
		Compiler string
		Dir      string
		Imports  []string
		Plugins  []Plugin
		Files    []string
	}
)

// New creates a new Lint.
func New(cfg Config) *Generator {
	return &Generator{
		deps:          cfg.Deps,
		plugins:       cfg.Plugins,
		moduleReflect: cfg.ModuleReflect,
	}
}
