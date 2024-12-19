package breakingcheck

import (
	"bytes"
	"fmt"
	"io"

	"github.com/easyp-tech/easyp/internal/lint"
)

func printIssues(w io.Writer, issues []lint.IssueInfo) error {
	return textPrinter(w, issues)
}

// textPrinter prints the error in text format.
func textPrinter(w io.Writer, issues []lint.IssueInfo) error {
	buffer := bytes.NewBuffer(nil)
	for _, issue := range issues {
		buffer.Reset()

		_, _ = buffer.WriteString(fmt.Sprintf("%s:%d:%d:%s %s (%s)",
			issue.Path,
			issue.Position.Line,
			issue.Position.Column,
			issue.SourceName,
			issue.Message,
			issue.RuleName,
		))
		_, _ = buffer.WriteString("\n")
		if _, err := w.Write(buffer.Bytes()); err != nil {
			return fmt.Errorf("w.Write: %w", err)
		}
	}

	return nil
}
