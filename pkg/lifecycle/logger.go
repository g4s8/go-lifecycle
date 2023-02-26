package lifecycle

import (
	"fmt"
	"io"
)

// Logger for lifecycle messages.
type Logger interface {
	// Printf prints formatted message.
	Printf(format string, v ...interface{})
}

type nopLogger struct{}

func (nopLogger) Printf(format string, v ...interface{}) {}

// NopLogger is a no-op logger.
var NopLogger = nopLogger{}

// StdLogger is a logger that writes to the specified writer.
type StdLogger struct {
	out io.Writer
}

// NewStdLogger creates a new logger that writes to the specified writer.
func NewStdLogger(out io.Writer) *StdLogger {
	return &StdLogger{
		out: out,
	}
}

func (l *StdLogger) Printf(format string, v ...interface{}) {
	fmt.Fprintf(l.out, format, v...)
}
