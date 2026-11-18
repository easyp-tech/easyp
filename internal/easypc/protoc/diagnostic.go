package protoc

import (
	"fmt"

	"github.com/easyp-tech/easyp/internal/easypc/pkg/parser"
)

// Diagnostic represents a compilation error or warning with source coordinates.
type Diagnostic struct {
	Path    string
	Line    int
	Col     int
	Message string
	IsWarn  bool
}

func (d Diagnostic) String() string {
	if d.Line <= 0 {
		if d.IsWarn {
			return fmt.Sprintf("%s: warning: %s", d.Path, d.Message)
		}
		return fmt.Sprintf("%s: %s", d.Path, d.Message)
	}
	if d.IsWarn {
		return fmt.Sprintf("%s:%d:%d: warning: %s", d.Path, d.Line, d.Col, d.Message)
	}
	return fmt.Sprintf("%s:%d:%d: %s", d.Path, d.Line, d.Col, d.Message)
}

func fromInternalDiag(d parser.Diagnostic) Diagnostic {
	return Diagnostic{
		Path:    d.Path,
		Line:    d.Line,
		Col:     d.Col,
		Message: d.Message,
		IsWarn:  d.IsWarn,
	}
}
