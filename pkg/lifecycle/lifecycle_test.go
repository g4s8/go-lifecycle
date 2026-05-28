package lifecycle_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"

	"github.com/g4s8/go-lifecycle/pkg/lifecycle"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestRunnerFunc(t *testing.T) {
	called := false
	var rf lifecycle.RunnerFunc = func(ctx context.Context) error {
		called = true
		return nil
	}
	err := rf.Run(context.Background())
	require.NoError(t, err)
	assert.True(t, called)
}

func TestRunnerFunc_Error(t *testing.T) {
	sentinel := errors.New("runner error")
	var rf lifecycle.RunnerFunc = func(ctx context.Context) error {
		return sentinel
	}
	err := rf.Run(context.Background())
	assert.ErrorIs(t, err, sentinel)
}

func TestLifecycle_NoRunners(t *testing.T) {
	var lf lifecycle.Lifecycle
	err := lf.Start(context.Background())
	require.NoError(t, err)
}

func TestLifecycle_SingleRunner_Success(t *testing.T) {
	var lf lifecycle.Lifecycle
	lf.Add(lifecycle.RunnerFunc(func(ctx context.Context) error {
		return nil
	}))
	err := lf.Start(context.Background())
	require.NoError(t, err)
}

func TestLifecycle_SingleRunner_Error(t *testing.T) {
	sentinel := errors.New("runner failed")
	var lf lifecycle.Lifecycle
	lf.Add(lifecycle.RunnerFunc(func(ctx context.Context) error {
		return sentinel
	}))
	err := lf.Start(context.Background())
	assert.ErrorIs(t, err, sentinel)
}

func TestLifecycle_MultipleRunners_AllSucceed(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var lf lifecycle.Lifecycle

	ready := make(chan struct{}, 3)
	for i := 0; i < 3; i++ {
		lf.Add(lifecycle.RunnerFunc(func(ctx context.Context) error {
			ready <- struct{}{}
			<-ctx.Done()
			return nil
		}))
	}

	done := make(chan error, 1)
	go func() { done <- lf.Start(ctx) }()

	// Wait for all runners to start
	for i := 0; i < 3; i++ {
		<-ready
	}
	cancel()

	err := <-done
	require.NoError(t, err)
}

func TestLifecycle_MultipleRunners_OneErrors(t *testing.T) {
	sentinel := errors.New("one runner failed")
	var lf lifecycle.Lifecycle

	lf.Add(lifecycle.RunnerFunc(func(ctx context.Context) error {
		return sentinel
	}))
	lf.Add(lifecycle.RunnerFunc(func(ctx context.Context) error {
		<-ctx.Done()
		return nil
	}))

	err := lf.Start(context.Background())
	assert.ErrorIs(t, err, sentinel)
}

func TestLifecycle_MultipleRunners_MultipleErrors(t *testing.T) {
	err1 := errors.New("first error")
	err2 := errors.New("second error")
	var lf lifecycle.Lifecycle

	lf.Add(lifecycle.RunnerFunc(func(ctx context.Context) error {
		return err1
	}))
	lf.Add(lifecycle.RunnerFunc(func(ctx context.Context) error {
		return err2
	}))

	err := lf.Start(context.Background())
	require.Error(t, err)
	assert.ErrorIs(t, err, err1)
	assert.ErrorIs(t, err, err2)
}

func TestLifecycle_ContextCanceled_IsIgnored(t *testing.T) {
	var lf lifecycle.Lifecycle
	lf.Add(lifecycle.RunnerFunc(func(ctx context.Context) error {
		return context.Canceled
	}))
	err := lf.Start(context.Background())
	require.NoError(t, err)
}

func TestLifecycle_ContextDeadlineExceeded_IsIgnored(t *testing.T) {
	var lf lifecycle.Lifecycle
	lf.Add(lifecycle.RunnerFunc(func(ctx context.Context) error {
		return context.DeadlineExceeded
	}))
	err := lf.Start(context.Background())
	require.NoError(t, err)
}
