package core

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/protoast"
)

// TestProtoastPositionsMatchLegacyContract is the golden proof of decision
// D-07b: for every first-token anchor both models expose (package, message,
// field, enum, enum value, service, rpc), protoast's Position must equal the
// legacy go-protoparser Meta.Pos on the same multi-byte UTF-8 + tab fixture.
// Current lint issue positions ARE these Meta.Pos values, so model-level
// equality proves the lint output contract carries over before any rule
// porting happens in Phase 4.
func TestProtoastPositionsMatchLegacyContract(t *testing.T) {
	t.Parallel()

	src, err := os.ReadFile("testdata/position_contract_utf8.proto")
	require.NoError(t, err)

	legacy, _, err := readProtoFileWithDirectives(bytes.NewReader(src))
	require.NoError(t, err)
	got, err := protoast.Read("testdata/position_contract_utf8.proto", bytes.NewReader(src))
	require.NoError(t, err)

	legacyBody := legacy.ProtoBody
	require.Len(t, legacyBody.Packages, 1)
	require.Len(t, legacyBody.Messages, 1)
	require.Len(t, legacyBody.Enums, 1)
	require.Len(t, legacyBody.Services, 1)

	legacyMessage := legacyBody.Messages[0]
	legacyEnum := legacyBody.Enums[0]
	legacyService := legacyBody.Services[0]
	message := got.Messages[0]
	enum := got.Enums[0]
	service := got.Services[0]

	anchors := []struct {
		name         string
		legacyLine   int
		legacyColumn int
		astPosition  protoast.Position
	}{
		{"package after a multi-byte string literal", legacyBody.Packages[0].Meta.Pos.Line, legacyBody.Packages[0].Meta.Pos.Column, got.Packages[0].Pos},
		{"message keyword", legacyMessage.Meta.Pos.Line, legacyMessage.Meta.Pos.Column, message.Pos},
		{"field after one leading tab", legacyMessage.MessageBody.Fields[0].Meta.Pos.Line, legacyMessage.MessageBody.Fields[0].Meta.Pos.Column, message.Fields[0].Pos},
		{"field after an emoji-bearing comment", legacyMessage.MessageBody.Fields[1].Meta.Pos.Line, legacyMessage.MessageBody.Fields[1].Meta.Pos.Column, message.Fields[1].Pos},
		{"pure-ASCII field", legacyMessage.MessageBody.Fields[2].Meta.Pos.Line, legacyMessage.MessageBody.Fields[2].Meta.Pos.Column, message.Fields[2].Pos},
		{"enum keyword", legacyEnum.Meta.Pos.Line, legacyEnum.Meta.Pos.Column, enum.Pos},
		{"enum value after one leading tab", legacyEnum.EnumBody.EnumFields[0].Meta.Pos.Line, legacyEnum.EnumBody.EnumFields[0].Meta.Pos.Column, enum.Values[0].Pos},
		{"enum value after a tab and an emoji-bearing comment", legacyEnum.EnumBody.EnumFields[1].Meta.Pos.Line, legacyEnum.EnumBody.EnumFields[1].Meta.Pos.Column, enum.Values[1].Pos},
		{"service keyword", legacyService.Meta.Pos.Line, legacyService.Meta.Pos.Column, service.Pos},
		{"rpc after two leading tabs", legacyService.ServiceBody.RPCs[0].Meta.Pos.Line, legacyService.ServiceBody.RPCs[0].Meta.Pos.Column, service.RPCs[0].Pos},
	}

	for _, a := range anchors {
		a := a
		t.Run(a.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, a.legacyLine, a.astPosition.Line, "line")
			assert.Equal(t, a.legacyColumn, a.astPosition.Column, "column")
		})
	}
}
