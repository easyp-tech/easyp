package api

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/urfave/cli/v2"
	"golang.org/x/mod/semver"

	"github.com/easyp-tech/easyp/internal/adapters/gitmodules"
	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/modules"
)

// Get adds a direct module requirement and resolves its dependency graph.
type Get struct{}

var _ Handler = Get{}

func (g Get) Command() *cli.Command {
	return &cli.Command{
		Name:      "get",
		Usage:     "add a Git module and its transitive dependencies",
		ArgsUsage: "<module>[@version|@commit]",
		Action:    g.Action,
	}
}

func (g Get) Action(ctx *cli.Context) error {
	if ctx.NArg() != 1 {
		return errors.New("get expects one module: easyp get <module>[@version|@commit]")
	}
	requirement, err := parseV1GetRequirement(ctx.Args().First())
	if err != nil {
		return fmt.Errorf("parseV1GetRequirement: %w", err)
	}
	root, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("Getwd: %w", err)
	}
	cache, err := moduleCache(ctx)
	if err != nil {
		return fmt.Errorf("moduleCache: %w", err)
	}
	return modules.Get(ctx.Context, root, requirement, cache)
}

func parseV1GetRequirement(spec string) (v1.Requirement, error) {
	version := ""
	source := spec
	if at := strings.LastIndex(spec, "@"); at > strings.LastIndex(spec, "/") {
		source, version = spec[:at], spec[at+1:]
		if !semver.IsValid(version) && !v1.IsCommitRef(version) {
			return v1.Requirement{}, fmt.Errorf("get %q: expected a semantic version or full Git commit after @", spec)
		}
	}
	if strings.ContainsAny(source, " \t\r\n") {
		return v1.Requirement{}, fmt.Errorf("get %q: invalid module identity", spec)
	}
	if err := gitmodules.ValidateSource(source); err != nil {
		return v1.Requirement{}, fmt.Errorf("get %q: %w", spec, err)
	}
	return v1.Requirement{Module: source, Version: version}, nil
}
