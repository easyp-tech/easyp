package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestV1PackageDirectoryPrefix(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		identity string
		want     string
	}{
		{identity: "example.test/rfc/user", want: "user"},
		{identity: "example.test/rfc/user/v2", want: "user"},
		{identity: "https://example.test/acme/order", want: "order"},
		{identity: "", want: ""},
	} {
		tt := tt
		t.Run(tt.identity, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, v1PackageDirectoryPrefix(tt.identity))
		})
	}
}
