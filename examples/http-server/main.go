package main

import (
	"log"
	"net/http"

	"github.com/g4s8/go-lifecycle/pkg/adaptors"
	"github.com/g4s8/go-lifecycle/pkg/health"
	"github.com/g4s8/go-lifecycle/pkg/lifecycle"
)

func main() {
	web := http.NewServeMux()
	web.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	svc := adaptors.NewHTTPService(&http.Server{
		Addr:    ":8080",
		Handler: web,
	})

	lf := lifecycle.New(lifecycle.DefaultConfig)
	svc.RegisterLifecycle("web", lf)
	hs := health.NewService(":9999", lf)
	hs.RegisterLifecycle(lf)
	lf.Start()
	sig := lifecycle.NewSignalHandler(lf, nil)
	sig.Start(lifecycle.DefaultShutdownConfig)
	if err := sig.Wait(); err != nil {
		log.Fatalf("shutdown error: %v", err)
	}
	log.Print("shutdown complete")
}
