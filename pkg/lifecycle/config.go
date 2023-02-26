package lifecycle

import (
	"time"
)

// StartStrategy is a strategy for startup.
type StartStrategy int

func (s StartStrategy) checkFlag(flag StartStrategy) bool {
	return s&flag == flag
}

const (
	// StartStrategyFailFast (default) returns error on first failed service.
	StartStrategyFailFast StartStrategy = 1 << iota
	// StartStrategyStartAll starts all services and returns error if any service failed.
	StartStrategyStartAll
	// StartStrategyRollbackOnError starts all services and rollback all started services on error.
	StartStrategyRollbackOnError
)

// Config is a lifecycle configuration.
type Config struct {
	// StartupTimeout is a timeout for startup.
	StartupTimeout time.Duration
	// ShutdownTimeout is a timeout for shutdown.
	ShutdownTimeout time.Duration
	// StartStrategy is a strategy for startup.
	StartStrategy StartStrategy
}

func (c *Config) check() {
	if c.StartupTimeout <= 0 {
		c.StartupTimeout = DefaultConfig.StartupTimeout
	}
	if c.ShutdownTimeout <= 0 {
		c.ShutdownTimeout = DefaultConfig.ShutdownTimeout
	}
	if c.StartStrategy == 0 {
		c.StartStrategy = DefaultConfig.StartStrategy
	}
}

// DefaultConfig is a default lifecycle configuration.
var DefaultConfig = Config{
	StartupTimeout:  5 * time.Second,
	ShutdownTimeout: 5 * time.Second,
	StartStrategy:   StartStrategyFailFast | StartStrategyRollbackOnError,
}
