package v1

import (
	"errors"
	"fmt"
	"io"

	"gopkg.in/yaml.v3"
)

// ErrLegacyConfiguration gives a consistent, non-writing migration entry point.
var ErrLegacyConfiguration = errors.New("legacy EasyP configuration detected; run easyp migrate --module <identity> to preview conversion to v1")

// LegacyPolicy identifies unversioned v0 documents without expanding values.
func LegacyPolicy(raw []byte) bool {
	var document yaml.Node
	err := yaml.Unmarshal(raw, &document)
	if err != nil || len(document.Content) != 1 || document.Content[0].Kind != yaml.MappingNode {
		return false
	}
	mapping := document.Content[0]
	legacy := false
	for i := 0; i+1 < len(mapping.Content); i += 2 {
		key, value := mapping.Content[i], mapping.Content[i+1]
		if key.Value == "version" && value.Value == "v1" {
			return false
		}
		switch key.Value {
		case "lint", "deps", "generate", "breaking_check":
			legacy = true
		}
	}
	return legacy
}

func requireSingleYAMLDocument(decoder *yaml.Decoder, name string) error {
	var extra any
	err := decoder.Decode(&extra)
	if errors.Is(err, io.EOF) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("Decode: %w", err)
	}
	return fmt.Errorf("%s must contain one YAML document", name)
}
