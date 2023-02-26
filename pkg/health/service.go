package health

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/g4s8/go-lifecycle/pkg/adaptors"
	"github.com/g4s8/go-lifecycle/pkg/lifecycle"
	"github.com/g4s8/go-lifecycle/pkg/types"
	"github.com/pkg/errors"
)

// Lifecycle monitor provider
type Lifecycle interface {
	SubscribeMonitor(ch chan<- []lifecycle.ServiceState) lifecycle.Subscription
}

// Service starts HTTP server on given address.
type Service struct {
	addr   string
	lf     Lifecycle
	stopCh chan struct{}

	statesSub lifecycle.Subscription
}

// NewService creates new health service.
func NewService(addr string, lf Lifecycle) *Service {
	return &Service{
		addr:   addr,
		lf:     lf,
		stopCh: make(chan struct{}),
	}
}

// RegisterLifecycle registers service in lifecycle manager.
func (s *Service) RegisterLifecycle(lf adaptors.LifecycleRegistry) {
	lf.RegisterService(types.ServiceConfig{
		Name:         "health",
		StartupHook:  s.Start,
		ShutdownHook: s.Stop,
		RestartPolicy: types.ServiceRestartPolicy{
			RestartOnFailure: true,
			RestartDelay:     time.Millisecond * 200,
		},
	})
}

func (s *Service) Start(ctx context.Context, errCh chan<- error) error {
	h := &handler{}
	statesCh := make(chan []lifecycle.ServiceState)
	go func() {
		for {
			select {
			case next := <-statesCh:
				h.update(next)
			case <-s.stopCh:
				close(statesCh)
				return
			}
		}
	}()
	s.statesSub = s.lf.SubscribeMonitor(statesCh)

	srv := &http.Server{
		Addr:    s.addr,
		Handler: h,
	}
	addr := srv.Addr
	if addr == "" {
		addr = ":http"
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return errors.Wrap(err, "listen tcp: %w")
	}
	go func() {
		if err := srv.Serve(ln); err != nil {
			if err != http.ErrServerClosed {
				errCh <- errors.Wrap(err, "serve")
				return
			}
		}
	}()
	return nil
}

func (s *Service) Stop(ctx context.Context) error {
	s.statesSub.Cancel()
	s.statesSub = nil
	close(s.stopCh)
	return nil
}
