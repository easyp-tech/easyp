package api

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/urfave/cli/v2"

	"github.com/easyp-tech/easyp/internal/config"
	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/fs/fs"
)

type LsFiles struct{}

var (
	flagLsFilesIncludeImports = &cli.BoolFlag{
		Name:    "include-imports",
		Usage:   "include transitive imports (like buf --include-imports)",
		Value:   true,
		Aliases: []string{"I"},
	}

	flagLsFilesFormat = &cli.GenericFlag{
		Name:       "format",
		Usage:      "output format: text or json",
		Required:   false,
		HasBeenSet: false,
		Value: &EnumValue{
			Enum:    []string{TextFormat, JSONFormat},
			Default: JSONFormat,
		},
		Aliases: []string{"f"},
	}
)

func (l LsFiles) Command() *cli.Command {
	return &cli.Command{
		Name:        "ls-files",
		Usage:       "list .proto files taking into account inputs and imports",
		UsageText:   "ls-files [--include-imports]",
		Description: "analog of 'buf ls-files --include-imports' for easyp",
		Action:      l.Action,
		Flags: []cli.Flag{
			flagLsFilesIncludeImports,
			flagLsFilesFormat,
		},
		Aliases: []string{"ls"},
	}
}

func (l LsFiles) Action(ctx *cli.Context) error {
	// Resolve config path, project root and operation root (search root).
	// Pass empty string for rootFlagName because this command doesn't declare its own --root flag.
	configPath, projectRoot, opRoot, err := resolveRoots(ctx, "")
	if err != nil {
		return err
	}

	cfg, err := config.New(ctx.Context, configPath)
	if err != nil {
		return fmt.Errorf("config.New: %w", err)
	}

	// Walker for Core (lockfile etc) - strictly based on project root
	dirWalker := fs.NewFSWalker(projectRoot, ".")
	app, err := buildCore(ctx.Context, *cfg, dirWalker)
	if err != nil {
		return fmt.Errorf("buildCore: %w", err)
	}

	opts := core.ListFilesOptions{
		IncludeImports: ctx.Bool(flagLsFilesIncludeImports.Name),
	}

	// Use operation root for listing files (allows future --root support)
	res, err := app.ListFiles(ctx.Context, opRoot, opts)
	if err != nil {
		return fmt.Errorf("app.ListFiles: %w", err)
	}

	format := ctx.String(flagLsFilesFormat.Name)
	switch format {
	case JSONFormat:
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(res); err != nil {
			return fmt.Errorf("json.Encode: %w", err)
		}
	case TextFormat:
		if err := printLsFilesText(res); err != nil {
			return fmt.Errorf("printLsFilesText: %w", err)
		}
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
	return nil
}

func printLsFilesText(res core.LsFilesResult) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)

	fmt.Fprintln(w, "ROOTS:")
	fmt.Fprintln(w, "  #\tSOURCE\tPATH")
	for i, r := range res.Roots {
		fmt.Fprintf(w, "  %d\t%s\t%s\n", i+1, r.Source, r.Path)
	}

	fmt.Fprintln(w, "\nFILES:")
	for i, f := range res.Files {
		fmt.Fprintf(w, "  #%d\n", i+1)
		fmt.Fprintf(w, "    SOURCE:\t%s\n", f.Source)
		fmt.Fprintf(w, "    IMPORT:\t%s\n", f.ImportPath)
		fmt.Fprintf(w, "    ABS:\t%s\n", f.AbsPath)
		fmt.Fprintf(w, "    ROOT:\t%s\n", f.Root)
		if i < len(res.Files)-1 {
			fmt.Fprintln(w, "  ---")
		}
	}

	if len(res.Errors) > 0 {
		fmt.Fprintln(w, "\nERRORS:")
		for i, e := range res.Errors {
			fmt.Fprintf(w, "  #%d\n", i+1)
			fmt.Fprintf(w, "    CODE:\t%s\n", e.Code)
			fmt.Fprintf(w, "    MSG:\t%s\n", e.Message)
			if i < len(res.Errors)-1 {
				fmt.Fprintln(w, "  ---")
			}
		}
	}

	return w.Flush()
}
