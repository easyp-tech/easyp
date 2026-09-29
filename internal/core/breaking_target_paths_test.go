package core

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yoheimuta/go-protoparser/v4/interpret/unordered"
)

func TestBreakingTargetPathsTakePrecedenceOverImportAliases(t *testing.T) {
	t.Parallel()
	item, err := readProtoFile(strings.NewReader("syntax = \"proto3\"; package demo; message Item {}"))
	require.NoError(t, err)
	use, err := readProtoFile(strings.NewReader("syntax = \"proto3\"; package demo; import \"item.proto\"; message Use { Item item = 1; }"))
	require.NoError(t, err)
	own := ProtoInfo{Path: "module/proto/item.proto", Info: item}
	importer := ProtoInfo{Path: "module/proto/use.proto", Info: use, ProtoFilesFromImport: map[ImportPath]*unordered.Proto{"item.proto": item}}
	for _, tt := range []struct {
		name  string
		files []ProtoInfo
	}{{"target first", []ProtoInfo{own, importer}}, {"target last", []ProtoInfo{importer, own}}} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			data, err := collect(tt.files)
			require.NoError(t, err)
			assert.Equal(t, "module/proto/item.proto", data[PackageName("demo")].Messages["Item"].ProtoFilePath)
		})
	}
}
