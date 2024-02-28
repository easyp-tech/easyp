package rules_test

import (
	"os"
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
	invalidAuthProtoEmptyPkg = `./../../../testdata/auth/empty_pkg.proto`
	validAuthProto           = `./../../../testdata/api/session/v1/session.proto`
	validAuthProto2          = `./../../../testdata/api/session/v1/events.proto`
)

func start(t testing.TB) (*require.Assertions, map[string]lint.ProtoInfo) {
	t.Helper()

	assert := require.New(t)

	protos := map[string]lint.ProtoInfo{
		invalidAuthProto:         parseFile(t, assert, invalidAuthProto),
		invalidAuthProto2:        parseFile(t, assert, invalidAuthProto2),
		invalidAuthProto3:        parseFile(t, assert, invalidAuthProto3),
		invalidAuthProto4:        parseFile(t, assert, invalidAuthProto4),
		invalidAuthProtoEmptyPkg: parseFile(t, assert, invalidAuthProtoEmptyPkg),
		validAuthProto:           parseFile(t, assert, validAuthProto),
		validAuthProto2:          parseFile(t, assert, validAuthProto2),
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

	return lint.ProtoInfo{
		Path: path,
		Info: res,
	}
}
