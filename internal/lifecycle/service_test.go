package lifecycle

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/g4s8/go-lifecycle/pkg/types"
	"github.com/stretchr/testify/require"
)

func TestStartFlow(t *testing.T) {
	t.Run("successful startup", func(t *testing.T) {
		ctx := newTestContext(t)
		sh := newEmptyStartupHook()
		cfg := types.ServiceConfig{
			StartupHook: sh,
		}
		svc := newTestServiceEntry(t, cfg)
		err := svc.Start(ctx)
		require.NoError(t, err)
		require.Equal(t, types.ServiceStatusRunning, svc.State().Status)
		require.NoError(t, svc.State().Error)
	})
	t.Run("startup error", func(t *testing.T) {
		ctx := newTestContext(t)
		targetErr := errors.New("test error1")
		sh := newStartupHookWithError(targetErr)
		cfg := types.ServiceConfig{
			StartupHook: sh,
		}
		svc := newTestServiceEntry(t, cfg)
		err := svc.Start(ctx)
		require.Error(t, err)
		require.ErrorIs(t, err, targetErr)
		require.Equal(t, types.ServiceStatusError, svc.State().Status)
		require.ErrorIs(t, svc.State().Error, targetErr)
	})
	t.Run("startup error timeout", func(t *testing.T) {
		ctx := newTestContext(t)
		ctx, cancel := context.WithTimeout(ctx, time.Millisecond*10)
		t.Cleanup(cancel)
		sh := newStartupHookWithTimeout(time.Millisecond * 100)
		cfg := types.ServiceConfig{
			StartupHook: sh,
		}
		svc := newTestServiceEntry(t, cfg)
		err := svc.Start(ctx)
		require.Error(t, err)
		require.ErrorIs(t, err, context.DeadlineExceeded)
		require.Equal(t, types.ServiceStatusError, svc.State().Status)
		require.ErrorIs(t, svc.State().Error, context.DeadlineExceeded)
	})
	t.Run("startup with timeout", func(t *testing.T) {
		ctx := newTestContext(t)
		ctx, cancel := context.WithTimeout(ctx, time.Millisecond*100)
		t.Cleanup(cancel)
		sh := newStartupHookWithTimeout(time.Millisecond * 200)
		cfg := types.ServiceConfig{
			StartupHook: sh,
		}
		svc := newTestServiceEntry(t, cfg)
		err := svc.Start(ctx)
		require.Error(t, err)
		require.ErrorIs(t, err, context.DeadlineExceeded)
		require.Equal(t, types.ServiceStatusError, svc.State().Status)
		require.ErrorIs(t, svc.State().Error, context.DeadlineExceeded)
	})
}

func TestRunningFlow(t *testing.T) {
	t.Run("runtime error", func(t *testing.T) {
		ctx := newTestContext(t)
		targetErr := errors.New("test runtime error")
		delay := time.Millisecond * 5
		sh := newStartupHookWithRuntimeErr(targetErr, delay)
		cfg := types.ServiceConfig{
			StartupHook: sh,
			RestartPolicy: types.ServiceRestartPolicy{
				RestartOnFailure: false,
			},
		}
		svc := newTestServiceEntry(t, cfg)
		err := svc.Start(ctx)
		require.NoError(t, err)
		select {
		case <-ctx.Done():
			t.Fatalf("context canceled: %v", ctx.Err())
		case <-time.After(delay * 2):
		}
		require.Equal(t, types.ServiceStatusError, svc.State().Status)
		require.ErrorIs(t, svc.State().Error, targetErr)
	})
	t.Run("runtime error recover simple", func(t *testing.T) {
		ctx := newTestContext(t)
		targetErr := errors.New("test runtime error 2")
		delay := time.Millisecond * 5
		sh := newStartupHookWithRuntimeErrOnce(targetErr, delay)
		cfg := types.ServiceConfig{
			StartupHook: sh,
			RestartPolicy: types.ServiceRestartPolicy{
				RestartOnFailure: true,
			},
		}
		svc := newTestServiceEntry(t, cfg)
		err := svc.Start(ctx)
		require.NoError(t, err)
		select {
		case <-ctx.Done():
			t.Fatalf("context canceled: %v", ctx.Err())
		case <-time.After(delay * 10):
		}
		require.Equal(t, types.ServiceStatusRunning, svc.State().Status)
		require.NoError(t, svc.State().Error)
	})
	t.Run("runtime error recover many", func(t *testing.T) {
		ctx := newTestContext(t)
		targetErr := errors.New("test runtime error 3")
		delay := time.Millisecond * 5
		sh := newStartupHookWithRuntimeErrCount(targetErr, delay, 4)
		cfg := types.ServiceConfig{
			StartupHook: sh,
			RestartPolicy: types.ServiceRestartPolicy{
				RestartOnFailure: true,
				RestartCount:     5,
			},
		}
		svc := newTestServiceEntry(t, cfg)
		err := svc.Start(ctx)
		require.NoError(t, err)
		select {
		case <-ctx.Done():
			t.Fatalf("context canceled: %v", ctx.Err())
		case <-time.After(delay * 10):
		}
		require.Equal(t, types.ServiceStatusRunning, svc.State().Status)
		require.NoError(t, svc.State().Error)
	})
	t.Run("runtime error recover many fail", func(t *testing.T) {
		ctx := newTestContext(t)
		targetErr := errors.New("test runtime error 4")
		delay := time.Millisecond * 4
		sh := newStartupHookWithRuntimeErrCount(targetErr, delay, 4)
		cfg := types.ServiceConfig{
			StartupHook: sh,
			RestartPolicy: types.ServiceRestartPolicy{
				RestartOnFailure: true,
				RestartCount:     3,
			},
		}
		svc := newTestServiceEntry(t, cfg)
		err := svc.Start(ctx)
		require.NoError(t, err)
		select {
		case <-ctx.Done():
			t.Fatalf("context canceled: %v", ctx.Err())
		case <-time.After(delay * 10):
		}
		require.Equal(t, types.ServiceStatusError, svc.State().Status)
		require.ErrorIs(t, svc.State().Error, targetErr)
	})
	t.Run("runtime error recover delay", func(t *testing.T) {
		ctx := newTestContext(t)
		targetErr := errors.New("test runtime error 5")
		delay := time.Millisecond * 1
		sh := newStartupHookWithRuntimeErrCount(targetErr, delay, 4)
		cfg := types.ServiceConfig{
			StartupHook: sh,
			RestartPolicy: types.ServiceRestartPolicy{
				RestartOnFailure: true,
				RestartCount:     5,
				RestartDelay:     time.Millisecond * 5,
			},
		}
		svc := newTestServiceEntry(t, cfg)
		err := svc.Start(ctx)
		require.NoError(t, err)
		select {
		case <-ctx.Done():
			t.Fatalf("context canceled: %v", ctx.Err())
		case <-time.After(delay * 20):
		}
		require.Equal(t, types.ServiceStatusRunning, svc.State().Status)
		require.NoError(t, svc.State().Error)
	})
}
