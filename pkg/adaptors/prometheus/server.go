package prometheus

import (
	"context"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	internal_http "github.com/g4s8/go-lifecycle/internal/http"
)

type Server struct {
	srv *http.Server
}

func NewServer(r *prometheus.Registry, port int, path string) *Server {
	mux := http.NewServeMux()
	mux.Handle(path, promhttp.HandlerFor(r, promhttp.HandlerOpts{}))

	srv := internal_http.NewServer("", port, internal_http.WithMux(mux))
	return &Server{srv: srv}
}

func (s *Server) Run(ctx context.Context) error {
	return internal_http.RunServer(ctx, s.srv)
}
