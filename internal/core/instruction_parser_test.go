package core_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/core"
)

func Test_InstructionParser_Parse(t *testing.T) {
	tests := map[string]struct {
		sourcePkgName string
		source        string

		pkgName         string
		instructionName string
	}{
		"google_api_package": {
			sourcePkgName:   "google/api",
			source:          "(google.api.http)",
			pkgName:         "google.api",
			instructionName: "http",
		},
		"mine_package": {
			sourcePkgName:   "mine",
			source:          "SomeMessage",
			pkgName:         "mine",
			instructionName: "SomeMessage",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			parser := core.InstructionParser{SourcePkgName: test.sourcePkgName}

			res := parser.Parse(test.source)
			require.Equal(t, test.pkgName, res.PkgName)
			require.Equal(t, test.instructionName, res.Instruction)
		})
	}
}
