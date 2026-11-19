package core

// Wave-0 position probe (04-RESEARCH "Position mapping", research A1
// remainder): the locked mapping table the whole Phase-4 port relies on.
//
// Position mapping (locked — every ported rule emits Decl.Pos where main used
// Meta.Pos):
//
//	declaration kind      old model (go-protoparser)      new model (protoast)
//	package               ProtoBody.Packages[i].Meta.Pos   File.Packages[i].Pos
//	import                ProtoBody.Imports[i].Meta.Pos    File.Imports[i].Pos
//	file-level option     ProtoBody.Options[i].Meta.Pos    File.Options[i].Pos
//	message               Messages[i].Meta.Pos             Messages[i].Pos
//	field (plain)         MessageBody.Fields[i].Meta.Pos   Message.Fields[i].Pos
//	oneof                 MessageBody.Oneofs[i].Meta.Pos   Message.Oneofs[i].Pos
//	field (oneof member)  Oneofs[i].OneofFields[i].Meta.Pos  Oneof.Fields[i].Pos
//	enum                  ProtoBody.Enums[i].Meta.Pos      File.Enums[i].Pos
//	enum value            EnumBody.EnumFields[i].Meta.Pos  Enum.Values[i].Pos
//	service               ProtoBody.Services[i].Meta.Pos   File.Services[i].Pos
//	rpc                   ServiceBody.RPCs[i].Meta.Pos     Service.RPCs[i].Pos
//
// Both sides anchor the declaration's FIRST TOKEN (main's keyword position;
// enum values anchor their name, which is their first token). For parity the
// rules must use Decl.Pos — never NamePos, EnumValue.NumberPos, RPCType.Pos
// (the etalon's finer tokens) or FilePosition; the two file-level rules
// (#2 PackageDefined, #27 FileLowerSnakeCase) emit the zero Position literal
// core.Position{Filename: path}. Filename is deliberately not compared:
// protoast.Parse stamps its path argument while go-protoparser leaves
// Meta.Filename empty (main's carried literals expect the empty Filename).

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yoheimuta/go-protoparser/v4/parser/meta"

	"github.com/easyp-tech/easyp/internal/protoast"
)

func TestDeclPositionProbe(t *testing.T) {
	t.Parallel()

	src := []byte(`syntax = "proto3";

package probe.pkg;

import "google/protobuf/empty.proto";

option go_package = "example.com/probe";

message Outer {
  string plain_field = 1;
  oneof choice {
    string member_field = 2;
  }
}

enum Kind {
  KIND_UNSPECIFIED = 0;
}

service Probe {
  rpc Do(Empty) returns (Empty) {}
}
`)

	legacy, err := readProtoFile(bytes.NewReader(src))
	require.NoError(t, err)
	got, err := protoast.Parse("probe.proto", src)
	require.NoError(t, err)

	require.Len(t, legacy.ProtoBody.Packages, 1)
	require.Len(t, legacy.ProtoBody.Imports, 1)
	require.Len(t, legacy.ProtoBody.Options, 1)
	require.Len(t, legacy.ProtoBody.Messages, 1)
	require.Len(t, legacy.ProtoBody.Enums, 1)
	require.Len(t, legacy.ProtoBody.Services, 1)
	legacyMessage := legacy.ProtoBody.Messages[0]
	require.Len(t, legacyMessage.MessageBody.Fields, 1)
	require.Len(t, legacyMessage.MessageBody.Oneofs, 1)
	legacyOneof := legacyMessage.MessageBody.Oneofs[0]
	require.Len(t, legacyOneof.OneofFields, 1)
	legacyEnum := legacy.ProtoBody.Enums[0]
	require.Len(t, legacyEnum.EnumBody.EnumFields, 1)

	anchors := []struct {
		name string
		meta meta.Position
		decl protoast.Position
	}{
		{"package", legacy.ProtoBody.Packages[0].Meta.Pos, got.Packages[0].Pos},
		{"import", legacy.ProtoBody.Imports[0].Meta.Pos, got.Imports[0].Pos},
		{"file_level_option", legacy.ProtoBody.Options[0].Meta.Pos, got.Options[0].Pos},
		{"message", legacyMessage.Meta.Pos, got.Messages[0].Pos},
		{"plain_field", legacyMessage.MessageBody.Fields[0].Meta.Pos, got.Messages[0].Fields[0].Pos},
		{"oneof", legacyOneof.Meta.Pos, got.Messages[0].Oneofs[0].Pos},
		{"oneof_member_field", legacyOneof.OneofFields[0].Meta.Pos, got.Messages[0].Oneofs[0].Fields[0].Pos},
		{"enum", legacyEnum.Meta.Pos, got.Enums[0].Pos},
		{"enum_value", legacyEnum.EnumBody.EnumFields[0].Meta.Pos, got.Enums[0].Values[0].Pos},
		{"service", legacy.ProtoBody.Services[0].Meta.Pos, got.Services[0].Pos},
		{"rpc", legacy.ProtoBody.Services[0].ServiceBody.RPCs[0].Meta.Pos, got.Services[0].RPCs[0].Pos},
	}

	require.Len(t, anchors, 11, "every probed declaration kind: the mapping table above stays complete")

	for _, a := range anchors {
		a := a
		t.Run(a.name, func(t *testing.T) {
			t.Parallel()

			want := fromMetaPos(a.meta)
			require.Equal(t, want.Offset, a.decl.Offset, "offset: Meta.Pos == Decl.Pos (first token)")
			require.Equal(t, want.Line, a.decl.Line, "line: Meta.Pos == Decl.Pos (first token)")
			require.Equal(t, want.Column, a.decl.Column, "column: Meta.Pos == Decl.Pos (first token)")
		})
	}
}
