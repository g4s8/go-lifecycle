package lifecycle

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/g4s8/go-lifecycle/pkg/types"
)

type testLogger struct {
	t *testing.T
}

func (l *testLogger) Msgf(format string, args ...interface{}) {
	if l == nil || l.t == nil {
		return
	}
	l.t.Helper()
	l.t.Logf(format, args...)
}

type testLoggerKey struct{}

func newTestContext(t *testing.T) context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	return context.WithValue(ctx, testLoggerKey{}, &testLogger{t})
}

func newEmptyStartupHook() types.StartupHook {
	return func(context.Context, chan<- error) error {
		return nil
	}
}

func newStartupHookWithError(err error) types.StartupHook {
	return func(context.Context, chan<- error) error {
		return err
	}
}

func newStartupHookWithTimeout(timeout time.Duration) types.StartupHook {
	return func(ctx context.Context, _ chan<- error) error {
		if timeout > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(timeout):
			}
		}
		return nil
	}
}

func newStartupHookWithRuntimeErr(err error, wait time.Duration) types.StartupHook {
	return func(ctx context.Context, errCh chan<- error) error {
		go func() {
			time.Sleep(wait)
			errCh <- err
		}()
		return nil
	}
}

func newStartupHookWithRuntimeErrOnce(err error, wait time.Duration) types.StartupHook {
	return newStartupHookWithRuntimeErrCount(err, wait, 1)
}

func newStartupHookWithRuntimeErrCount(err error, wait time.Duration, count int) types.StartupHook {
	c := int32(count)
	return func(ctx context.Context, errCh chan<- error) error {
		go func() {
			if atomic.AddInt32(&c, -1)+1 > 0 {
				time.Sleep(wait)
				errCh <- err
			}
		}()
		return nil
	}
}
