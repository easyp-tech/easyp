package config

const (
	SeverityError = "error"
	SeverityWarn  = "warn"
)

// ValidationIssue describes a problem with a v1 configuration file.
type ValidationIssue struct {
	Code     string `json:"code" yaml:"code"`
	Message  string `json:"message" yaml:"message"`
	Line     int    `json:"line,omitempty" yaml:"line,omitempty"`
	Column   int    `json:"column,omitempty" yaml:"column,omitempty"`
	Severity string `json:"severity,omitempty" yaml:"severity,omitempty"`
}

func HasErrors(issues []ValidationIssue) bool {
	for _, issue := range issues {
		if issue.Severity == SeverityError {
			return true
		}
	}
	return false
}
