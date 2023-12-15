package flags

import (
	"flag"
	"log/slog"
)

var _ flag.Value = (*Level)(nil)

// Level for setting level by flag.
type Level struct {
	Level slog.Level
}

// String implements flag.Value.
func (l *Level) String() string {
	return l.Level.String()
}

// Set implements flag.Value.
func (l *Level) Set(s string) error {
	return l.Level.UnmarshalText([]byte(s))
}
