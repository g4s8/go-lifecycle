// Package lifecycle provides a simple interface for managing the application lifecycle.
//
// The main component of this package is Lifecycle, which is a service lifecycle manager.
// It provides methods to register startup and shutdown hooks, and to start and stop the service.
// Lifecycle also provides a method to subscribe to service state changes, which can be used to monitor the service state.
//
// Lifecycle package also provides a SignalHandler component,
// which can be used to handle OS signals and stop the service on receiving a signal.
//
// `Config` type contains configuration for Lifecycle, the default configuration can be obtained with `DefaultConfig`.
package lifecycle

import (
	"context"
	"sync"

	"github.com/g4s8/go-lifecycle/internal/lifecycle"
	"github.com/g4s8/go-lifecycle/pkg/types"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

// Lifecycle represents application lifecycle manager.
// It provides methods to register startup and shutdown hooks, and to start and stop the service.
type Lifecycle struct {
	config Config

	mx       sync.RWMutex
	services []*lifecycle.ServiceEntry
	configs  []types.ServiceConfig
	stateMx  sync.RWMutex
	states   []lifecycle.ServiceState
	doneCh   chan struct{}
	statePub *publisher[[]ServiceState]
}

// New creates new lifecycle manager.
func New(config Config) *Lifecycle {
	config.check()
	return &Lifecycle{
		config:   config,
		doneCh:   make(chan struct{}),
		statePub: new(publisher[[]ServiceState]),
	}
}

// RegisterStartupHook adds startup hook to lifecycle.
func (l *Lifecycle) RegisterStartupHook(name string, h types.StartupHook) {
	cfg := types.ServiceConfig{
		Name:          name,
		StartupHook:   h,
		RestartPolicy: types.DefaultRestartPolicy,
	}
	l.RegisterService(cfg)
}

// RegisterShutdownHook adds shutdown hook to lifecycle.
func (l *Lifecycle) RegisterShutdownHook(name string, h types.ShutdownHook) {
	cfg := types.ServiceConfig{
		Name:          name,
		ShutdownHook:  h,
		RestartPolicy: types.DefaultRestartPolicy,
	}
	l.RegisterService(cfg)
}

// RegisterService registers service to lifecycle with config.
func (l *Lifecycle) RegisterService(service types.ServiceConfig) {
	l.mx.Lock()
	defer l.mx.Unlock()

	stateCh := make(chan lifecycle.ServiceState)
	go l.runServiceMonitor(len(l.services), stateCh)
	l.services = append(l.services, lifecycle.NewServiceEntry(service, stateCh))
	l.configs = append(l.configs, service)
	l.states = append(l.states, lifecycle.ServiceState{Status: types.ServiceStatusInit})
}

// Statuses returns current statuses of all registered services and hooks.
func (l *Lifecycle) Statuses() []ServiceState {
	l.stateMx.RLock()
	defer l.stateMx.RUnlock()

	states := make([]ServiceState, len(l.states))
	for i, state := range l.states {
		states[i] = ServiceState{
			ID:     i,
			Name:   l.configs[i].Name,
			Status: state.Status,
			Error:  state.Error,
		}
	}
	return states
}

// SubscribeMonitor subscribes to lifecycle service state monitor.
func (l *Lifecycle) SubscribeMonitor(ch chan<- []ServiceState) Subscription {
	return l.statePub.subscribe(ch)
}

// Start starts all registered startup hooks.
func (l *Lifecycle) Start() error {
	l.mx.RLock()
	defer l.mx.RUnlock()

	baseCtx := context.Background()
	startCtx, cancel := context.WithTimeout(baseCtx, l.config.StartupTimeout)
	defer cancel()

	var errs error
	for _, svc := range l.services {
		select {
		case <-startCtx.Done():
			errs = multierr.Append(errs, errors.Wrap(startCtx.Err(), "startup timeout"))
			break
		default:
		}

		if err := svc.Start(startCtx); err != nil {
			errs = multierr.Append(errs, err)
			if l.config.StartStrategy.checkFlag(StartStrategyFailFast) {
				break
			}
		}
	}

	if errs != nil {
		if l.config.StartStrategy.checkFlag(StartStrategyRollbackOnError) {
			stopCtx, cancel := context.WithTimeout(baseCtx, l.config.ShutdownTimeout)
			defer cancel()

			if err := l.stop(stopCtx, true); err != nil {
				errs = multierr.Append(errs, errors.Wrap(err, "failed to stop lifecycle"))
			}
		}
		return errs
	}

	return nil
}

// Stop stops all registered shutdown hooks.
func (l *Lifecycle) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), l.config.ShutdownTimeout)
	defer cancel()
	return l.stop(ctx, false)
}

// Close closes lifecycle manager.
func (l *Lifecycle) Close() error {
	for _, svc := range l.services {
		svc.Close()
	}
	close(l.doneCh)
	return nil
}

func (l *Lifecycle) stop(ctx context.Context, rollback bool) error {
	l.mx.RLock()
	defer l.mx.RUnlock()

	var errs error
	for i := len(l.services) - 1; i >= 0; i-- {
		svc := l.services[i]
		if rollback && svc.State().Status == types.ServiceStatusInit {
			continue
		}
		if err := svc.Stop(ctx); err != nil {
			errs = multierr.Append(errs, err)
		}
	}
	return errs
}

func (l *Lifecycle) runServiceMonitor(id int, stateCh chan lifecycle.ServiceState) {
	for {
		select {
		case state := <-stateCh:
			l.stateMx.Lock()
			l.states[id] = state
			l.stateMx.Unlock()
			newState := make([]ServiceState, len(l.states))
			for i, state := range l.states {
				newState[i] = ServiceState{
					ID:     i,
					Name:   l.configs[i].Name,
					Status: state.Status,
					Error:  state.Error,
				}
			}
			l.statePub.publish(newState)
		case <-l.doneCh:
			close(stateCh)
			return
		}
	}
}
