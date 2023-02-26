// The MIT License (MIT) Copyright (c) 2023 github.com/g4s8
// https://github.com/g4s8/go-lifecycle/blob/master/LICENSE

/*
Package lifecycle provides service lifecycle management.

This package provides a set of tools to manage service lifecycle, including startup and shutdown hooks, health checks, and service monitoring.

The main component of this package is `Lifecycle`, which is a service lifecycle manager. It provides methods to register startup and shutdown hooks,
and to start and stop the service. `Lifecycle` also provides a method to subscribe to service state changes, which can be used to monitor the service state.

Lifecycle package also provides a `SignalHandler` component, which can be used to handle OS signals and stop the service on receiving a signal.

Adaptors package provides a set of adaptors to integrate with other packages, such as web server adaptor.

Health package provides a health check service, which can be used to check the service health.

Package overview:
  - `github.com/g4s8/go-lifecycle/pkg/lifecycle` - components to manage lifecycle
  - `github.com/g4s8/go-lifecycle/pkg/adaptors` - some common adaptors, e.g. http server adaptor
  - `github.com/g4s8/go-lifecycle/pkg/health` - healthcheck for services
  - `github.com/g4s8/go-lifecycle/pkg/types` - public types

Example:
```go
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

```
*/
package lifecycle
