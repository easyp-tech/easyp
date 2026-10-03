package flags

import "github.com/urfave/cli/v2"

// Frozen returns a fresh flag for each command and the global scope.
func Frozen() *cli.BoolFlag {
	return &cli.BoolFlag{Name: "frozen", Usage: "require the existing manifest and lock without dependency resolution or file updates (explicit; never inferred from CI)"}
}

// IsFrozen preserves any true value in the command lineage, including a global
// flag shadowed by a command-local default or an explicit false value.
func IsFrozen(ctx *cli.Context) bool {
	for _, scope := range ctx.Lineage() {
		if scope.Bool("frozen") {
			return true
		}
	}
	return false
}
