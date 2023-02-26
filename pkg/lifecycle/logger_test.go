package lifecycle

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStdLogger(t *testing.T) {
	var buf bytes.Buffer
	l := NewStdLogger(&buf)
	l.Printf("hello %s", "world")
	require.Equal(t, "hello world", buf.String())
}
