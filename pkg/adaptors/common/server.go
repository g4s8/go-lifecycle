package common

import (
	"context"
	"net/http"

	internal_http "github.com/g4s8/go-lifecycle/internal/http"
)

type HttpRunner struct {
	srv *http.Server
}

func NewHttpRunner(srv *http.Server) *HttpRunner {
	return &HttpRunner{srv: srv}
}

func (h *HttpRunner) Run(ctx context.Context) error {
	return internal_http.RunServer(ctx, h.srv)
}
