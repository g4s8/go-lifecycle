package lifecycle

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/g4s8/go-lifecycle/pkg/types"
	"github.com/pkg/errors"
)

type stateTransition [2]types.ServiceStatus

type stateTransitionHandler func(ctx context.Context, service *ServiceEntry, transition stateTransition) error

var stateTransitionsV1 = map[stateTransition]stateTransitionHandler{
	{types.ServiceStatusInit, types.ServiceStatusStarting}:     onStart,
	{types.ServiceStatusStarting, types.ServiceStatusStopping}: onStop,
	{types.ServiceStatusRunning, types.ServiceStatusStopping}:  onStop,
	{types.ServiceStatusRunning, types.ServiceStatusError}:     onRuntimeError,
	{types.ServiceStatusStopped, types.ServiceStatusStarting}:  onStart,
	{types.ServiceStatusError, types.ServiceStatusStarting}:    onStart,
}

func onStart(ctx context.Context, service *ServiceEntry, transition stateTransition) error {
	// INIT -> STARTING
	if service.cfg.StartupHook != nil {
		if err := service.cfg.StartupHook(ctx, service.errCh); err != nil {
			return errors.Wrap(err, "start service")
		}
	}
	service.push(ctx, ServiceState{Status: types.ServiceStatusRunning})
	return nil
}

func onStop(ctx context.Context, service *ServiceEntry, transition stateTransition) error {
	// RUNNING -> STOPPING
	if service.cfg.ShutdownHook != nil {
		if err := service.cfg.ShutdownHook(ctx); err != nil {
			return errors.Wrap(err, "stop service")
		}
	}
	service.push(ctx, ServiceState{Status: types.ServiceStatusStopped})
	return nil
}

func onRuntimeError(ctx context.Context, service *ServiceEntry, transition stateTransition) error {
	restartPol := service.cfg.RestartPolicy
	if !restartPol.RestartOnFailure {
		return service.state.Error
	}
	if service.restartState == nil {
		service.restartState = &restartState{}
	}
	if restartPol.RestartCount > 0 && service.restartState.tryCount >= restartPol.RestartCount {
		return service.state.Error
	}
	if service.restartState.lastAttempt.IsZero() {
		service.restartState.lastAttempt = time.Now()
	} else if interval := time.Since(service.restartState.lastAttempt); restartPol.RestartDelay != 0 &&
		interval < restartPol.RestartDelay {
		t := time.NewTimer(service.cfg.RestartPolicy.RestartDelay - interval)
		select {
		case <-t.C:
		case <-ctx.Done():
			return ctx.Err()
		case <-service.closeCh:
			return errors.New("service closed")
		}
	}
	service.restartState.tryCount++
	service.restartState.lastAttempt = time.Now()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	service.cancelMx.Lock()
	service.cancelFn = cancel
	service.cancelMx.Unlock()
	service.push(ctx, ServiceState{Status: types.ServiceStatusStarting})
	return nil
}

// ServiceState represents service status and optional error.
type ServiceState struct {
	Status types.ServiceStatus
	Error  error
}

const (
	serviceLoopRunning int32 = 1
	serviceLoopStopped int32 = 0
)

type restartState struct {
	tryCount    int
	lastAttempt time.Time
}

// ServiceEntry is an internal service entry implementation.
type ServiceEntry struct {
	cfg             types.ServiceConfig
	transitionsSpec map[stateTransition]stateTransitionHandler
	stateCh         chan<- ServiceState

	errCh        chan error
	state        ServiceState
	stateMx      sync.RWMutex
	sq           stateQueue
	running      int32
	cancelFn     context.CancelFunc
	cancelMx     sync.Mutex
	closeCh      chan struct{}
	doneWg       sync.WaitGroup
	restartState *restartState
}

func newTestServiceEntry(t *testing.T, cfg types.ServiceConfig) *ServiceEntry {
	t.Helper()
	stateCh := make(chan ServiceState)
	doneCh := make(chan struct{})
	go func() {
		for {
			select {
			case <-stateCh:
			case <-doneCh:
				return
			}
		}
	}()
	svc := NewServiceEntry(cfg, stateCh)
	t.Cleanup(func() {
		svc.Close()
		close(doneCh)
		close(stateCh)
	})
	return svc
}

// NewServiceEntry creates new service entry.
func NewServiceEntry(cfg types.ServiceConfig, stateCh chan<- ServiceState) *ServiceEntry {
	entry := &ServiceEntry{
		cfg:             cfg,
		transitionsSpec: stateTransitionsV1,
		errCh:           make(chan error),
		closeCh:         make(chan struct{}, 1),
		stateCh:         stateCh,
	}
	entry.doneWg.Add(1)
	go entry.errorsLoop()
	return entry
}

// Start servvice.
func (e *ServiceEntry) Start(ctx context.Context) error {
	return e.changeState(ctx, types.ServiceStatusStarting)
}

// Stop service.
func (e *ServiceEntry) Stop(ctx context.Context) error {
	return e.changeState(ctx, types.ServiceStatusStopping)
}

// State of the service.
func (e *ServiceEntry) State() ServiceState {
	e.stateMx.RLock()
	defer e.stateMx.RUnlock()
	return e.state
}

func (e *ServiceEntry) changeState(ctx context.Context, status types.ServiceStatus) error {
	ctx = e.changeContext(ctx)
	e.push(ctx, ServiceState{Status: status})
	e.stateMx.RLock()
	defer e.stateMx.RUnlock()
	if s := e.state; s.Status == types.ServiceStatusError {
		return s.Error
	}
	return nil
}

func (e *ServiceEntry) errorsLoop() {
	defer e.doneWg.Done()

	for {
		select {
		case err := <-e.errCh:
			ctx := e.changeContext(context.Background())
			e.push(ctx, ServiceState{Status: types.ServiceStatusError, Error: err})
		case <-e.closeCh:
			return
		}
	}
}

func (e *ServiceEntry) changeContext(ctx context.Context) context.Context {
	e.cancelMx.Lock()
	defer e.cancelMx.Unlock()

	if cancelFn := e.cancelFn; cancelFn != nil {
		cancelFn()
	}
	ctx, e.cancelFn = context.WithCancel(ctx)
	return ctx
}

func (e *ServiceEntry) push(ctx context.Context, state ServiceState) {
	e.sq.push(state)
	if !atomic.CompareAndSwapInt32(&e.running, serviceLoopStopped, serviceLoopRunning) {
		return
	}
	e.loop(ctx)
	atomic.StoreInt32(&e.running, serviceLoopStopped)
	e.stateMx.RLock()
	if e.state.Status == types.ServiceStatusError && isCtxErr(e.state.Error) {
		fmt.Println("DEBUG: stop on context err")
		e.stateMx.RUnlock()
		return
	}
	e.stateMx.RUnlock()
	if e.sq.len() > 0 {
		last, ok := e.sq.pop()
		if ok {
			e.push(ctx, last)
		}
	}
}

func (e *ServiceEntry) loop(ctx context.Context) {
	e.doneWg.Add(1)
	defer e.doneWg.Done()

	for {
		select {
		case <-ctx.Done():
			e.stateMx.Lock()
			e.state = ServiceState{Status: types.ServiceStatusError, Error: ctx.Err()}
			e.stateMx.Unlock()
			return
		default:
		}

		item, ok := e.sq.pop()
		if !ok {
			return
		}
		e.applyTransition(ctx, item)
		e.stateCh <- item
	}
}

func (e *ServiceEntry) applyTransition(ctx context.Context, state ServiceState) {
	transition := stateTransition{e.state.Status, state.Status}
	e.stateMx.Lock()
	e.state = state
	e.stateMx.Unlock()
	if handler, ok := e.transitionsSpec[transition]; ok {
		err := handler(ctx, e, transition)
		if err != nil {
			e.sq.push(ServiceState{Status: types.ServiceStatusError, Error: err})
			return
		}
	}
}

// Close service entry.
func (e *ServiceEntry) Close() {
	close(e.closeCh)
	e.doneWg.Wait()
}
