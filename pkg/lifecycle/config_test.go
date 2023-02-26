package lifecycle

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	var cfg Config
	cfg.check()
	require.Equal(t, DefaultConfig.StartupTimeout, cfg.StartupTimeout)
	require.Equal(t, DefaultConfig.ShutdownTimeout, cfg.ShutdownTimeout)
	require.Equal(t, DefaultConfig.StartStrategy, cfg.StartStrategy)
}
