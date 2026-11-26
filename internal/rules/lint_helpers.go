package rules

import (
	"github.com/easyp-tech/easyp/internal/core"
)

// groupedValues collects one value per file within groups such as packages or
// directories, and reports the files of a group whose values disagree.
type groupedValues struct {
	groups map[string][]groupedValue
	order  []string
}

type groupedValue struct {
	path, value, source string
	pos                 core.Position
}

func (g *groupedValues) add(group string, value groupedValue) {
	if g.groups == nil {
		g.groups = make(map[string][]groupedValue)
	}
	if _, ok := g.groups[group]; !ok {
		g.order = append(g.order, group)
	}
	g.groups[group] = append(g.groups[group], value)
}

// firstExemptMismatches is the ported group rules' Finish with main's
// first-exempt semantics: the first value of a group (walk order; add
// preserves insertion order) is exempt, and only later files whose value
// differs from it issue, each at its own position.
func (g *groupedValues) firstExemptMismatches(rule core.Rule) []core.IssueInfo {
	var res []core.IssueInfo
	for _, group := range g.order {
		first := g.groups[group][0]
		for _, v := range g.groups[group][1:] {
			if v.value == first.value {
				continue
			}
			for _, issue := range core.AppendIssue(nil, rule, v.pos, v.source) {
				res = append(res, core.IssueInfo{Issue: issue, Path: v.path})
			}
		}
	}
	*g = groupedValues{}
	return res
}
