package rules_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yoheimuta/go-protoparser/v4"
	"github.com/yoheimuta/go-protoparser/v4/interpret/unordered"

	"github.com/easyp-tech/easyp/internal/lint"
)

const (
	invalidAuthProto         = `./../../../testdata/auth/service.proto`
	invalidAuthProto2        = `./../../../testdata/auth/queue.proto`
	invalidAuthProto3        = `./../../../testdata/auth/InvalidName.proto`
	invalidAuthProto4        = `./../../../testdata/invalid_pkg/queue.proto`
	invalidAuthProto5        = `./../../../testdata/invalid_options/queue.proto`
	invalidAuthProto6        = `./../../../testdata/invalid_options/session.proto`
	invalidAuthProtoEmptyPkg = `./../../../testdata/auth/empty_pkg.proto`
	invalidsSessionProto     = `./../../../testdata/invalid_pkg/session.proto`
	validAuthProto           = `./../../../testdata/api/session/v1/session.proto`
	validAuthProto2          = `./../../../testdata/api/session/v1/events.proto`
	importUsed               = "./../../../testdata/import_used/used.proto"
	importNotUsed            = "./../../../testdata/import_used/not_used.proto"
	noLintBufComment         = "./../../../testdata/no_lint/no_lint_buf_comment.proto"
	noLintEasypComment       = "./../../../testdata/no_lint/no_lint_easyp_comment.proto"
)

func start(t testing.TB) (*require.Assertions, map[string]lint.ProtoInfo) {
	t.Helper()

	assert := require.New(t)

	protos := map[string]lint.ProtoInfo{
		invalidAuthProto:         parseFile(t, assert, invalidAuthProto),
		invalidAuthProto2:        parseFile(t, assert, invalidAuthProto2),
		invalidAuthProto3:        parseFile(t, assert, invalidAuthProto3),
		invalidAuthProto4:        parseFile(t, assert, invalidAuthProto4),
		invalidAuthProto5:        parseFile(t, assert, invalidAuthProto5),
		invalidAuthProto6:        parseFile(t, assert, invalidAuthProto6),
		invalidAuthProtoEmptyPkg: parseFile(t, assert, invalidAuthProtoEmptyPkg),
		invalidsSessionProto:     parseFile(t, assert, invalidsSessionProto),
		validAuthProto:           parseFile(t, assert, validAuthProto),
		validAuthProto2:          parseFile(t, assert, validAuthProto2),
		importUsed:               parseFile(t, assert, importUsed),
		importNotUsed:            parseFile(t, assert, importNotUsed),
		noLintBufComment:         parseFile(t, assert, noLintBufComment),
		noLintEasypComment:       parseFile(t, assert, noLintEasypComment),
	}

	return assert, protos
}

func parseFile(t testing.TB, assert *require.Assertions, path string) lint.ProtoInfo {
	t.Helper()

	f, err := os.Open(path)
	assert.NoError(err)
	t.Cleanup(func() { assert.NoError(f.Close()) })

	got, err := protoparser.Parse(f)
	assert.NoError(err)

	res, err := unordered.InterpretProto(got)
	assert.NoError(err)

	protoFilesFromImport := make(map[lint.ImportPath]*unordered.Proto)

	// read imports files
	for _, imp := range res.ProtoBody.Imports {
		importPath := lint.ConvertImportPath(imp.Location)
		fullPath := filepath.Join("./../../../testdata", string(importPath))

		f, err := os.Open(fullPath)
		if err != nil {
			continue
		}

		t.Cleanup(func() { assert.NoError(f.Close()) })

		got, err := protoparser.Parse(f)
		assert.NoError(err)

		res, err := unordered.InterpretProto(got)
		assert.NoError(err)

		protoFilesFromImport[importPath] = res
	}

	return lint.ProtoInfo{
		Path:                 path,
		Info:                 res,
		ProtoFilesFromImport: protoFilesFromImport,
	}
}
