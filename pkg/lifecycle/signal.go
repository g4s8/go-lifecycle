package lifecycle

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// SignalHandler is an OS signal handler that can be used to trigger a
// lifecycle shutdown.
type SignalHandler struct {
	lifecycle *Lifecycle
	logger    Logger
	signals   []os.Signal

	waitCh chan error
}

// NewSignalHandler creates a new SignalHandler that will trigger the given
// lifecycle on the given signals. If no signals are given, it will trigger on
// SIGINT and SIGTERM.
func NewSignalHandler(lifecycle *Lifecycle, logger Logger, signals ...os.Signal) *SignalHandler {
	if len(signals) == 0 {
		signals = []os.Signal{os.Interrupt, syscall.SIGTERM}
	}
	if logger == nil {
		logger = NewStdLogger(os.Stderr)
	}
	return &SignalHandler{
		lifecycle: lifecycle,
		logger:    logger,
		signals:   signals,
		waitCh:    make(chan error),
	}
}

// ShutdownConfig is the configuration for the graceful shutdown.
type ShutdownConfig struct {
	ExitOnShutdown bool
}

// DefaultShutdownConfig is the default configuration for the graceful shutdown.
var DefaultShutdownConfig = ShutdownConfig{}

// Start starts the signal handler in a new goroutine.
func (h *SignalHandler) Start(cfg ShutdownConfig) {
	go func() {
		defer close(h.waitCh)

		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		ctx, cancel := context.WithTimeout(context.Background(), h.lifecycle.config.ShutdownTimeout)
		defer cancel()
		if err := h.lifecycle.stop(ctx, false); err != nil {
			h.logger.Printf("failed to stop lifecycle: %v", err)
			h.waitCh <- err
			if cfg.ExitOnShutdown {
				os.Exit(1)
			}
		}
		if cfg.ExitOnShutdown {
			os.Exit(0)
		}
	}()
}

// Wait waits for the signal handler to stop.
func (h *SignalHandler) Wait() (err error) {
	for next := range h.waitCh {
		err = next
	}
	return
}
