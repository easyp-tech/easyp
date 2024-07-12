package lint

import (
	"errors"

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
}

// BuildError creates an Issue.
func BuildError(lintRule Rule, pos meta.Position, sourceName string) Issue {
	return Issue{
		Position:   pos,
		SourceName: sourceName,
		Message:    lintRule.Message(),
	}
}
