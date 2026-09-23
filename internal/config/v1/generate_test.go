package v1

import (
	"strings"
	"testing"
)

func TestParseGenerate(t *testing.T) {
	got, err := ParseGenerate(strings.NewReader(`version: v1
generate:
  modules: [proto/user]
plugins:
  - name: go
    out: ./gen/go
    opts:
      paths: source_relative
options:
  go:
    package_prefix: github.com/acme/gen/go
`))
	if err != nil {
		t.Fatal(err)
	}
	if got.Generate.Modules[0] != "proto/user" || got.Plugins[0].Name != "go" || got.Options.Go.PackagePrefix == nil || *got.Options.Go.PackagePrefix != "github.com/acme/gen/go" {
		t.Fatalf("unexpected generator: %#v", got)
	}
}

func TestParseGenerateRejectsLegacyInputs(t *testing.T) {
	_, err := ParseGenerate(strings.NewReader("version: v1\ngenerate:\n  inputs: []\nplugins:\n  - name: python\n    out: ./gen\n"))
	if err == nil {
		t.Fatal("v1 must not accept generate.inputs")
	}
}

func TestParseGenerateRejectsInvalidManagedRule(t *testing.T) {
	_, err := ParseGenerate(strings.NewReader("version: v1\ngenerate:\n  managed:\n    enabled: true\n    override:\n      - file_option: go_package_prefix\n"))
	if err == nil {
		t.Fatal("managed override without a value was accepted")
	}
}

func TestParseGenerateListOptions(t *testing.T) {
	got, err := ParseGenerate(strings.NewReader("version: v1\nplugins:\n  - name: go\n    out: ./gen\n    opts: [paths=source_relative]\n"))
	if err != nil {
		t.Fatal(err)
	}
	if got.Plugins[0].Opts["paths"][0] != "source_relative" {
		t.Fatalf("unexpected opts: %#v", got.Plugins[0].Opts)
	}
}

func TestParseGenerateDefaultsVersion(t *testing.T) {
	got, err := ParseGenerate(strings.NewReader("plugins:\n  - name: python\n    out: ./gen\n"))
	if err != nil {
		t.Fatal(err)
	}
	if got.Version != "v1" {
		t.Fatalf("version = %q", got.Version)
	}
}

func TestRemotePluginRequiresPinnedVersion(t *testing.T) {
	if _, err := ParseGenerate(strings.NewReader("version: v1\nplugins:\n  - remote: localhost:8080/python\n    out: gen\n")); err == nil {
		t.Fatal("unpinned remote plugin was accepted")
	}
	got, err := ParseGenerate(strings.NewReader("version: v1\nplugins:\n  - remote: localhost:8080/python\n    version: v1.2.3\n    out: gen\n"))
	if err != nil || got.Plugins[0].Version != "v1.2.3" {
		t.Fatalf("pinned remote plugin rejected: %v, %#v", err, got)
	}
	if _, err := ParseGenerate(strings.NewReader("version: v1\nplugins:\n  - name: python\n    version: v1.2.3\n    out: gen\n")); err == nil {
		t.Fatal("unverifiable local plugin version was accepted")
	}
}
