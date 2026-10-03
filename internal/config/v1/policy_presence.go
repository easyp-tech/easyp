package v1

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// PolicySection is a section name that participates in section-scoped cascading.
const (
	PolicySectionLinters        = "linters"
	PolicySectionLinterSettings = "linters-settings"
	PolicySectionIssues         = "issues"
	PolicySectionBreaking       = "breaking"
)

// PolicyPresence records which sections and section fields a policy file
// actually wrote. ParsePolicy fills parsed defaults such as the STANDARD preset,
// so merging can only distinguish an explicit local choice from an inserted
// default through the source text.
type PolicyPresence struct {
	Sections map[string]bool
	Fields   map[string]map[string]bool
}

// NewPolicyPresence returns an empty presence record.
func NewPolicyPresence() PolicyPresence {
	return PolicyPresence{Sections: make(map[string]bool), Fields: make(map[string]map[string]bool)}
}

// ParsePolicyPresence reports the sections and fields written by a policy file.
// It is used for local policies and for every loaded base policy, and never
// parses values, so a malformed base still fails in ParsePolicy.
func ParsePolicyPresence(raw []byte) (PolicyPresence, error) {
	presence := NewPolicyPresence()
	// Decode mappings instead of walking raw nodes: yaml.v3 applies aliases and
	// merge-key precedence, including fields supplied through <<.
	var document yaml.Node
	if err := yaml.Unmarshal(raw, &document); err != nil {
		return presence, fmt.Errorf("Unmarshal: %w", err)
	}
	if document.Kind != yaml.DocumentNode || len(document.Content) == 0 || document.Content[0].Kind != yaml.MappingNode {
		return presence, nil
	}
	var sections map[string]yaml.Node
	if err := document.Content[0].Decode(&sections); err != nil {
		return presence, fmt.Errorf("Unmarshal: %w", err)
	}
	for name, section := range sections {
		presence.Sections[name] = true
		if section.Kind != yaml.MappingNode && section.Kind != yaml.AliasNode {
			continue
		}
		var fields map[string]yaml.Node
		if err := section.Decode(&fields); err != nil {
			return presence, fmt.Errorf("policy section %s: %w", name, err)
		}
		present := make(map[string]bool, len(fields))
		for field := range fields {
			present[field] = true
		}
		presence.Fields[name] = present
	}
	return presence, nil
}

// Has reports whether the section itself is present in the source.
func (p PolicyPresence) Has(section string) bool { return p.Sections[section] }

// Field reports whether a field of a section is present in the source.
func (p PolicyPresence) Field(section, field string) bool { return p.Fields[section][field] }
