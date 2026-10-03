package api

import (
	"errors"
	"fmt"
	"strings"

	"github.com/urfave/cli/v2"
	"golang.org/x/mod/semver"

	"github.com/easyp-tech/easyp/internal/adapters/gitmodules"
	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/flags"
	"github.com/easyp-tech/easyp/internal/modules"
)

// Get adds a direct module requirement and resolves its dependency graph.
type Get struct{}

var _ Handler = Get{}

func (g Get) Command() *cli.Command {
	return &cli.Command{
		Name:      "get",
		Flags:     []cli.Flag{flags.Frozen()},
		Usage:     "add a Git module and its transitive dependencies",
		ArgsUsage: "<module>[@version|@tag|@commit]",
		Action:    g.Action,
	}
}

func (g Get) Action(ctx *cli.Context) error {
	if flags.IsFrozen(ctx) {
		return fmt.Errorf("get is not allowed in frozen mode")
	}
	if ctx.NArg() != 1 {
		return errors.New("get expects one module: easyp get <module>[@version|@tag|@commit]")
	}
	requirement, err := parseV1GetRequirement(ctx.Args().First())
	if err != nil {
		return fmt.Errorf("parseV1GetRequirement: %w", err)
	}
	root, err := moduleWorkingDir()
	if err != nil {
		return fmt.Errorf("moduleWorkingDir: %w", err)
	}
	cache, err := moduleCache(ctx)
	if err != nil {
		return fmt.Errorf("moduleCache: %w", err)
	}
	if requirement.Version != "" && !semver.IsValid(requirement.Version) && !v1.IsCommitRef(requirement.Version) {
		_, module, err := modules.ReadManifest(root)
		if err != nil {
			return fmt.Errorf("ReadManifest: %w", err)
		}
		if module.Name == requirement.Module {
			return fmt.Errorf("module %s cannot require itself", module.Name)
		}
		commit, err := cache.ResolveTag(ctx.Context, requirement.Module, requirement.Version)
		if err != nil {
			return fmt.Errorf("ResolveTag: %w", err)
		}
		requirement.Version = commit
	}
	return modules.Get(ctx.Context, root, requirement, cache)
}

func parseV1GetRequirement(spec string) (v1.Requirement, error) {
	version := ""
	source := spec
	// The @ in an SSH URL's authority is not a version separator.
	authorityEnd := -1
	if scheme := strings.Index(spec, "://"); scheme >= 0 {
		if slash := strings.Index(spec[scheme+3:], "/"); slash >= 0 {
			authorityEnd = scheme + 3 + slash
		} else {
			authorityEnd = len(spec)
		}
	}
	if at := strings.LastIndex(spec, "@"); at > authorityEnd {
		source, version = spec[:at], spec[at+1:]
		if version == "" || strings.ContainsAny(version, " \t\r\n") {
			return v1.Requirement{}, fmt.Errorf("get %q: expected a version, Git tag or full Git commit after @", spec)
		}
	}
	if strings.ContainsAny(source, " \t\r\n") {
		return v1.Requirement{}, fmt.Errorf("get %q: invalid module identity", spec)
	}
	if err := gitmodules.ValidateSource(source); err != nil {
		return v1.Requirement{}, fmt.Errorf("get %q: %w", spec, err)
	}
	major, err := v1.ModulePathMajor(source)
	if err != nil {
		return v1.Requirement{}, fmt.Errorf("ModulePathMajor: %w", err)
	}
	if major == "" && semver.IsValid(version) && semver.Major(version) != "v0" && semver.Major(version) != "v1" && semver.Build(version) == "" {
		// Fetch still verifies that this revision has no native protobuf.mod.
		version += "+incompatible"
	}
	return v1.Requirement{Module: source, Version: version}, nil
}
