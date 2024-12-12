// Package generate provides the core functionality of EasyP generate.
package generate

import (
	"bytes"
	"strings"

	"github.com/samber/lo"

	"github.com/easyp-tech/easyp/internal/api/shared/module_reflect"
)

type (
	// Plugin is a plugin for gRPC generator.
	Plugin struct {
		Name    string
		Out     string
		Options map[string]string
	}
	// Inputs is the source for generating code.
	Inputs struct {
		Dirs []string
	}
	// Config is the configuration for EasyP generate.
	Config struct {
		Deps          []string
		Plugins       []Plugin
		Inputs        Inputs
		ModuleReflect *modulereflect.ModuleReflect
	}
	// Generator is the core functionality of EasyP generate.
	Generator struct {
		deps          []string
		plugins       []Plugin
		inputs        Inputs
		moduleReflect *modulereflect.ModuleReflect
	}
	// Query is a query for making sh command.
	Query struct {
		Compiler string
		Imports  []string
		Plugins  []Plugin
		Files    []string
	}
)

// New creates a new Lint.
//
// example
//
//	protoc \
//	 -I . \
//	 -I /usr/local/include \
//	 --go_out=. \
//	 --go_opt=paths=source_relative \
//	 --go-grpc_out=. \
//	 --go-grpc_opt=paths=source_relative \
//	 proto/hello.proto
func New(cfg Config) *Generator {
	return &Generator{
		deps:          cfg.Deps,
		plugins:       cfg.Plugins,
		moduleReflect: cfg.ModuleReflect,
		inputs:        cfg.Inputs,
	}
}

func (q Query) build() string {
	var buf bytes.Buffer

	buf.WriteString(q.Compiler)
	buf.WriteString(" \\\n")

	for _, imp := range q.Imports {
		buf.WriteString(" -I ")
		buf.WriteString(imp)
		buf.WriteString(" \\\n")
	}

	for _, plugin := range q.Plugins {
		buf.WriteString(" --")
		buf.WriteString(plugin.Name)
		buf.WriteString("_out=")
		buf.WriteString(plugin.Out)
		buf.WriteString(" \\\n")
		buf.WriteString(" --")
		buf.WriteString(plugin.Name)
		buf.WriteString("_opt=")

		options := lo.MapToSlice(plugin.Options, func(k string, v string) string {
			return k + "=" + v
		})
		buf.WriteString(strings.Join(options, ","))
		buf.WriteString(" \\\n")
	}

	for i, file := range q.Files {
		buf.WriteString(" ")
		buf.WriteString(file)

		if i != len(q.Files)-1 {
			buf.WriteString(" \\\n")
		}
	}

	return buf.String()
}
