package rules

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCasingMatchesBuf(t *testing.T) {
	pascal := map[string]bool{"Foo": true, "FOO": true, "HTTPServer": true, "Foo2bar": true, "SearchAds360Service": true,
		"Foo_Bar": false, "FOO_BAR": false, "_Foo": false, "foo": false}
	for name, ok := range pascal {
		assert.Equal(t, ok, toPascalCase(name) == name, "pascal %q -> %q", name, toPascalCase(name))
	}
	lower := map[string]bool{"foo_bar": true, "foo1bar": true, "foo_": false, "foo__bar": false, "fooBar": false, "_foo": false}
	for name, ok := range lower {
		assert.Equal(t, ok, toLowerSnakeCase(name) == name, "lower %q -> %q", name, toLowerSnakeCase(name))
	}
	assert.Equal(t, "HTTP_STATUS", toUpperSnakeCase("HTTPStatus"))
	assert.Equal(t, "http_server", toLowerSnakeCase("HTTPServer"))
	assert.Equal(t, "foo_bar", toLowerSnakeCase("fooBar"))
}
