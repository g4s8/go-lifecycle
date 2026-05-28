package http

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"
)

const DefaultReadTimeout = 5 * time.Second

type Option func(*http.Server)

func WithMux(mux *http.ServeMux) Option {
	return func(srv *http.Server) {
		srv.Handler = mux
	}
}

func WithReadTimeout(timeout time.Duration) Option {
	return func(srv *http.Server) {
		srv.ReadTimeout = timeout
	}
}

func NewServer(addr string, port int, opts ...Option) *http.Server {
	srv := http.Server{
		Addr:        fmt.Sprintf("%s:%d", addr, port),
		ReadTimeout: DefaultReadTimeout,
	}
	for _, opt := range opts {
		opt(&srv)
	}
	return &srv
}

const GracefulShutdownTimeout = 5 * time.Second

func RunServer(ctx context.Context, srv *http.Server) error {
	ln, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		return fmt.Errorf("listen tcp: %w", err)
	}

	doneCh := make(chan struct{})
	errCh := make(chan error, 1)

	go func(s *http.Server) {
		defer close(doneCh)
		if err := s.Serve(ln); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				return
			}
			errCh <- fmt.Errorf("serve: %w", err)
		}
	}(srv)

	var errs []error
	select {
	case <-ctx.Done():
		if err := shutdownServer(srv); err != nil {
			errs = append(errs, err)
		}
	case err := <-errCh:
		errs = append(errs, err)
		if err := ln.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close listener: %w", err))
		}
	}

	<-doneCh

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

func shutdownServer(srv *http.Server) error {
	ctx, cancel := context.WithTimeout(context.Background(), GracefulShutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown server: %w", err)
	}
	return nil
}
