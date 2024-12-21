package lint

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_InstructionParser_Parse(t *testing.T) {
	tests := map[string]struct {
		sourcePkgName string
		source        string

		pkgName         string
		instructionName string
	}{
		"google api package": {
			sourcePkgName:   "google/api",
			source:          "(google.api.http)",
			pkgName:         "google.api",
			instructionName: "http",
		},
		"mine package": {
			sourcePkgName:   "mine",
			source:          "SomeMessage",
			pkgName:         "mine",
			instructionName: "SomeMessage",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			parser := InstructionParser{SourcePkgName: test.sourcePkgName}

			res := parser.Parse(test.source)
			require.Equal(t, test.pkgName, res.PkgName)
			require.Equal(t, test.instructionName, res.Instruction)
		})
	}
}
