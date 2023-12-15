package rules_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yoheimuta/go-protoparser/v4"
	"github.com/yoheimuta/go-protoparser/v4/interpret/unordered"
)

const (
	invalidAuthProto = `./../../testdata/auth/service.proto`
)

func start(t testing.TB) (*require.Assertions, map[string]*unordered.Proto) {
	t.Helper()

	assert := require.New(t)

	protos := map[string]*unordered.Proto{
		invalidAuthProto: parseFile(t, assert, invalidAuthProto),
	}

	return assert, protos
}

func parseFile(t testing.TB, assert *require.Assertions, path string) *unordered.Proto {
	t.Helper()

	f, err := os.Open(path)
	assert.NoError(err)
	t.Cleanup(func() { assert.NoError(f.Close()) })

	got, err := protoparser.Parse(f)
	assert.NoError(err)

	res, err := unordered.InterpretProto(got)
	assert.NoError(err)

	return res
}
