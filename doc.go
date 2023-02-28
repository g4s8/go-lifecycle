// The MIT License (MIT) Copyright (c) 2023 github.com/g4s8
// https://github.com/g4s8/go-lifecycle/blob/master/LICENSE

/*
Package lifecycle is a powerful toolset for managing the lifecycle of services,
including startup and shutdown hooks, health checks, and service monitoring.

At the heart of the package is the Lifecycle component, which acts as a service lifecycle manager.
It provides a variety of methods for registering startup and shutdown hooks, as well as starting and stopping the service.
Additionally, the Lifecycle component can be used to subscribe to service state changes, which can be a valuable tool for monitoring the service state.

In addition to the Lifecycle component, the package includes a SignalHandler component,
which can be used to gracefully handle OS signals and stop the service when necessary.

To integrate the lifecycle package with other packages,
the adaptors package provides a variety of adaptors.
For example, the http server adaptor allows for seamless integration with web servers.

Finally, the health package provides a health check service that can be used to monitor the health of the service.

Package Overview:
  - github.com/g4s8/go-lifecycle/pkg/lifecycle: Components for managing service lifecycle
  - github.com/g4s8/go-lifecycle/pkg/adaptors: Common adaptors, such as the http server adaptor
  - github.com/g4s8/go-lifecycle/pkg/health: Healthcheck service for monitoring service health
  - github.com/g4s8/go-lifecycle/pkg/types: Public types used in the package.

Example:

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
*/
package lifecycle
