package v1

import (
	"maps"
	"slices"

	"github.com/easyp-tech/easyp/internal/rules"
)

// DefaultLinterPreset is the preset a policy file uses when its linters section
// does not select one.
const DefaultLinterPreset = "STANDARD"

// linterFields and breakingFields name the mergeable fields of a section.
const (
	lintersFieldDefault  = "default"
	lintersFieldComments = "allow_comment_ignores"

	breakingFieldBaseline = "baseline"
	breakingFieldCategory = "categories"
	breakingFieldUnstable = "ignore_unstable"
	breakingFieldIgnore   = "ignore"
)

// BaseLinterSection turns the linters section of a base policy into a decided
// section: every field carries a value, so a local section above it only has to
// state its own adjustments. An absent preset and an absent comment setting take
// the same defaults a standalone policy file would use.
func BaseLinterSection(policy LinterPolicy, presence PolicyPresence) LinterPolicy {
	base := LinterPolicy{
		Default: DefaultLinterPreset,
		Enable:  slices.Clone(policy.Enable),
		Disable: slices.Clone(policy.Disable),
	}
	if presence.Field(PolicySectionLinters, lintersFieldDefault) {
		base.Default = policy.Default
	}
	base.AllowCommentIgnores = boolPointer(policy.AllowCommentIgnores, true)
	return base
}

// MergeLinterSection applies a resolved base section under the local section.
// Without a base the local section is returned unchanged, which keeps the
// whole-section cascade identical for policies that do not extend anything.
func MergeLinterSection(base, local LinterPolicy, presence PolicyPresence, hasBase bool) LinterPolicy {
	if !hasBase {
		return local
	}
	merged := LinterPolicy{Default: base.Default}
	if presence.Field(PolicySectionLinters, lintersFieldDefault) {
		merged.Default = local.Default
	}
	selections := newLinterSelections()
	selections.addLayer(base.Enable, base.Disable)
	selections.addLayer(local.Enable, local.Disable)
	merged.Enable, merged.Disable = selections.resolve()
	if presence.Field(PolicySectionLinters, lintersFieldComments) {
		merged.AllowCommentIgnores = boolPointer(local.AllowCommentIgnores, false)
	} else {
		merged.AllowCommentIgnores = boolPointer(base.AllowCommentIgnores, true)
	}
	return merged
}

// boolPointer copies a presence-aware boolean setting, falling back to the
// value an absent setting means for that layer.
func boolPointer(value *bool, fallback bool) *bool {
	resolved := fallback
	if value != nil {
		resolved = *value
	}
	return &resolved
}

// MergeLinterSettings merges base settings under the local settings section.
// Local values win per rule and key; a present but empty local section clears
// everything the base provided.
func MergeLinterSettings(base, local map[string]map[string]string, presence PolicyPresence, hasBase bool) map[string]map[string]string {
	if !hasBase {
		return local
	}
	merged := make(map[string]map[string]string, len(base)+len(local))
	if !presence.Has(PolicySectionLinterSettings) {
		for rule, values := range base {
			merged[rule] = maps.Clone(values)
		}
		return merged
	}
	if len(local) == 0 {
		return merged
	}
	for rule, values := range base {
		merged[rule] = maps.Clone(values)
	}
	for rule, values := range local {
		if len(values) == 0 {
			merged[rule] = maps.Clone(values)
			continue
		}
		if merged[rule] == nil {
			merged[rule] = make(map[string]string)
		}
		maps.Copy(merged[rule], values)
	}
	return merged
}

// BaseBreakingSection turns the breaking section of a base policy into a
// decided section. The values are applied at the consuming policy location, so
// a baseline or ignore path taken from a base is interpreted by the consumer.
func BaseBreakingSection(policy BreakingPolicy, presence PolicyPresence) BreakingPolicy {
	return BreakingPolicy{
		Baseline:       policy.Baseline,
		Categories:     slices.Clone(policy.Categories),
		IgnoreUnstable: policy.IgnoreUnstable,
		Ignore:         slices.Clone(policy.Ignore),
	}
}

// MergeBreakingSection applies a resolved base section under the local section.
// A local value overrides the base only when the source actually wrote it, so an
// explicit false, an empty category list and an empty ignore list survive.
func MergeBreakingSection(base, local BreakingPolicy, presence PolicyPresence, hasBase bool) BreakingPolicy {
	if !hasBase {
		return local
	}
	merged := BreakingPolicy{
		Baseline:       base.Baseline,
		Categories:     slices.Clone(base.Categories),
		IgnoreUnstable: base.IgnoreUnstable,
		Ignore:         slices.Clone(base.Ignore),
	}
	if presence.Field(PolicySectionBreaking, breakingFieldBaseline) {
		merged.Baseline = local.Baseline
	}
	if presence.Field(PolicySectionBreaking, breakingFieldCategory) {
		merged.Categories = slices.Clone(local.Categories)
	}
	if presence.Field(PolicySectionBreaking, breakingFieldUnstable) {
		merged.IgnoreUnstable = local.IgnoreUnstable
	}
	if presence.Field(PolicySectionBreaking, breakingFieldIgnore) {
		merged.Ignore = slices.Clone(local.Ignore)
	}
	return merged
}

// linterSelections accumulates effective enable and disable selections while
// walking the chain from the deepest base to the local policy.
type linterSelections struct {
	enabled  map[string]bool
	disabled map[string]bool
}

func newLinterSelections() *linterSelections {
	return &linterSelections{enabled: make(map[string]bool), disabled: make(map[string]bool)}
}

// addLayer applies one policy layer. Groups expand to their rules so a local
// enable re-enables a single rule that a base group disabled, and a disable
// written in the same layer wins over an enable.
func (s *linterSelections) addLayer(enable, disable []string) {
	for _, selection := range expandLinterSelections(enable) {
		delete(s.disabled, selection)
		s.enabled[selection] = true
	}
	for _, selection := range expandLinterSelections(disable) {
		delete(s.enabled, selection)
		s.disabled[selection] = true
	}
}

// resolve returns the enabled and disabled selections in stable lexical order, so
// the engine expands the same rules once and never reports a rule twice.
func (s *linterSelections) resolve() (enabled, disabled []string) {
	for selection := range s.enabled {
		enabled = append(enabled, selection)
	}
	for selection := range s.disabled {
		disabled = append(disabled, selection)
	}
	slices.Sort(enabled)
	slices.Sort(disabled)
	return enabled, disabled
}

// expandLinterSelections replaces group keys with their rules and removes
// duplicates while preserving the written order.
func expandLinterSelections(selections []string) []string {
	groups := rules.AllGroups()
	var expanded []string
	seen := make(map[string]bool, len(selections))
	for _, selection := range selections {
		names := []string{selection}
		for _, group := range groups {
			if group.Key == selection {
				names = group.Rules
				break
			}
		}
		for _, name := range names {
			if seen[name] {
				continue
			}
			seen[name] = true
			expanded = append(expanded, name)
		}
	}
	return expanded
}
