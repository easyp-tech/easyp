package rules_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yoheimuta/go-protoparser/v4"
	"github.com/yoheimuta/go-protoparser/v4/interpret/unordered"

	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/protoast"
)

const (
	invalidAuthProto         = `./../../testdata/auth/service.proto`
	invalidAuthProto2        = `./../../testdata/auth/queue.proto`
	invalidAuthProto3        = `./../../testdata/auth/InvalidName.proto`
	invalidAuthProto4        = `./../../testdata/invalid_pkg/queue.proto`
	invalidAuthProto5        = `./../../testdata/invalid_options/queue.proto`
	invalidAuthProto6        = `./../../testdata/invalid_options/session.proto`
	invalidAuthProtoEmptyPkg = `./../../testdata/auth/empty_pkg.proto`
	invalidsSessionProto     = `./../../testdata/invalid_pkg/session.proto`
	validAuthProto           = `./../../testdata/api/session/v1/session.proto`
	validAuthProto2          = `./../../testdata/api/session/v1/events.proto`
	importUsed               = "./../../testdata/import_used/used.proto"
	importNotUsed            = "./../../testdata/import_used/not_used.proto"
	noLintBufComment         = "./../../testdata/no_lint/no_lint_buf_comment.proto"
	noLintEasypComment       = "./../../testdata/no_lint/no_lint_easyp_comment.proto"
)

func start(t testing.TB) (*require.Assertions, map[string]core.ProtoInfo) {
	t.Helper()

	assert := require.New(t)

	protos := map[string]core.ProtoInfo{
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

func parseFile(t testing.TB, assert *require.Assertions, path string) core.ProtoInfo {
	t.Helper()

	f, err := os.Open(path)
	assert.NoError(err)
	t.Cleanup(func() { assert.NoError(f.Close()) })

	got, err := protoparser.Parse(f)
	assert.NoError(err)

	res, err := unordered.InterpretProto(got)
	assert.NoError(err)

	protoFilesFromImport := make(map[core.ImportPath]*unordered.Proto)

	// read imports files
	for _, imp := range res.ProtoBody.Imports {
		importPath := core.ConvertImportPath(imp.Location)
		fullPath := filepath.Join("./../../testdata", string(importPath))

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

	return core.ProtoInfo{
		Path:                 path,
		Info:                 res,
		ProtoFilesFromImport: protoFilesFromImport,
	}
}

// parseLinkedFile builds a LinkedFileInfo for a testdata file (the linked
// contract's helper, mirroring the etalon's protoast-era parseFile): the file
// itself is parsed with protoast.Parse and its direct imports are parsed from
// ./../../testdata, skipping the ones that are not stored there.
func parseLinkedFile(t testing.TB, path string) core.LinkedFileInfo {
	t.Helper()

	src, err := os.ReadFile(path)
	require.NoError(t, err)

	file, err := protoast.Parse(path, src)
	require.NoError(t, err)

	imported := make(map[core.ImportPath]*protoast.File)

	// read imports files
	for _, imp := range file.Imports {
		importPath := core.ConvertImportPath(imp.Location)
		fullPath := filepath.Join("./../../testdata", string(importPath))

		src, err := os.ReadFile(fullPath)
		if err != nil {
			continue
		}

		parsed, err := protoast.Parse(string(importPath), src)
		require.NoError(t, err)

		imported[importPath] = parsed
	}

	return core.LinkedFileInfo{
		Path:          path,
		File:          file,
		ImportedFiles: imported,
	}
}

// validate runs a linked-contract rule over infos the way lint does: module
// rules collect every file and report once, other rules report per file.
func validate(rule core.LintRule, infos ...core.LinkedFileInfo) ([]core.Issue, error) {
	moduleRule, ok := rule.(core.ModuleRule)
	if !ok {
		var issues []core.Issue
		for _, info := range infos {
			found, err := rule.Validate(info)
			if err != nil {
				return nil, err
			}
			issues = append(issues, found...)
		}
		return issues, nil
	}
	for _, info := range infos {
		if err := moduleRule.Collect(info); err != nil {
			return nil, err
		}
	}
	var issues []core.Issue
	for _, found := range moduleRule.Finish() {
		issues = append(issues, found.Issue)
	}
	return issues, nil
}

// parseProtoLinked builds a LinkedFileInfo from source text for small, focused
// cases.
func parseProtoLinked(t testing.TB, path, src string) core.LinkedFileInfo {
	t.Helper()
	file, err := protoast.Parse(path, []byte(src))
	require.NoError(t, err)
	return core.LinkedFileInfo{Path: path, File: file}
}
