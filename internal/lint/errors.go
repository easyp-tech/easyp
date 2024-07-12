package lint

import (
	"errors"

	"github.com/yoheimuta/go-protoparser/v4/parser"
	"github.com/yoheimuta/go-protoparser/v4/parser/meta"
)

var (
	ErrInvalidRule = errors.New("invalid rule")
)

// IssueInfo contains the information of an issue and the path.
type IssueInfo struct {
	Issue
	Path string
}

// Issue contains the information of an issue.
type Issue struct {
	Position   meta.Position
	SourceName string
	Message    string
	RuleName   string
}

// AppendIssue check if lint error is ignored -> add new error to slice
// otherwise ignore appending
func AppendIssue(
	issues []Issue, lintRule Rule, pos meta.Position, sourceName string, comments []*parser.Comment,
) []Issue {
	if CheckIsIgnored(comments, GetRuleName(lintRule)) {
		return issues
	}

	return append(issues, BuildError(lintRule, pos, sourceName))
}

// BuildError creates an Issue.
func BuildError(lintRule Rule, pos meta.Position, sourceName string) Issue {
	return Issue{
		Position:   pos,
		SourceName: sourceName,
		Message:    lintRule.Message(),
		RuleName:   GetRuleName(lintRule),
	}
}
