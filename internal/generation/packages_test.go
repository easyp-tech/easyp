package generation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSourcePackageReadsOnlyDeclaration(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct{ name, source, want string }{
		{name: "simple", source: "syntax = \"proto3\"; package demo.v1; message M {}", want: "demo.v1"},
		{name: "comments", source: "/*package fake;*/ syntax='proto3'; //package wrong;\npackage demo /*comment*/ . v1 ;", want: "demo.v1"},
		{name: "string option", source: "syntax=\"proto3\"; option java_package=\"package wrong;\"; package demo.v1;", want: "demo.v1"},
		{name: "field name", source: "message M { string package = 1; }"},
		{name: "late declaration", source: "message M {} package demo.v1;", want: "demo.v1"},
		{name: "broken unselected body", source: "package other.v1; message {", want: "other.v1"},
		{name: "malformed package", source: "package demo..v1;"},
		{name: "empty", source: ""},
		{name: "byte order mark", source: "\ufeffpackage demo.v1;", want: "demo.v1"},
	} {
		t.Run(tt.name, func(t *testing.T) { t.Parallel(); assert.Equal(t, tt.want, sourcePackage([]byte(tt.source))) })
	}
}
