package common

import (
	"context"
	"encoding/json"
	"net/http"

	internal_http "github.com/g4s8/go-lifecycle/internal/http"
)

func NewHealthRunner(port int, path string, checker func(ctx context.Context) []error) *HttpRunner {
	type healthResponse struct {
		Status string   `json:"status"`
		Errors []string `json:"errors,omitempty"`
	}

	mux := http.NewServeMux()
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		var resp healthResponse
		status := http.StatusOK
		if errs := checker(r.Context()); len(errs) > 0 {
			errorMessages := make([]string, len(errs))
			for i, err := range errs {
				errorMessages[i] = err.Error()
			}
			resp.Status = "unhealthy"
			resp.Errors = errorMessages
			status = http.StatusServiceUnavailable
		} else {
			resp.Status = "healthy"
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.WriteHeader(status)

		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(resp); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
	srv := internal_http.NewServer("", port, internal_http.WithMux(mux))
	return NewHttpRunner(srv)
}
