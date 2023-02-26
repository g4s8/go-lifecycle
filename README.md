Simple and flexible Go lifecycle manager library.

It helps to confiigure and run multiple services within one application. It has components to
manage application lifecycle, handle system signals to implement graceful shutdown, service state monitoring
tools, health-check built-in service, and adapters for common scenarios (e.g. http server).

In some aspects it's similar to github.com/uber-go/fx library but there are main differences:
 - it's not DI (dependency injection) containers, only lifecycle manager.
 - no reflection and magic inside, just static types.
 - flexible configuration for lifecycle manager and for each service if needed.
 - built-in healthcheck service monitoring tools.

You can check `/examples` dir to see working code.

See godoc: [![GoDoc](https://godoc.org/github.com/g4s8/go-lifecycle?status.svg)](https://godoc.org/github.com/g4s8/go-lifecycle)

## Install

Install using `go get`:
```bash
go get github.com/g4s8/go-lifecycle@latest
```

## Usage

### Register startup and shutdown hooks:

In this example:
 - `lifecycle.New` - creates new lifecycle manager.
 - `RegisterStartupHook` - registers func as startup hook and calls it during startup phase.
 - `RegisterShutdownHook` - registers func as shutdown hook and calls it on shutdown.
 - `lf.Start` - starts all services and call startup hooks.

```go
lf := lifecycle.New(lifecycle.DefaultConfig)
lf.RegisterStartupHook("my-service", func(ctx context.Context, errCh chan<- error) error {
        fmt.Println("My service started")
        return nil
})
lf.RegisterShutdownHook("cleanup-hook", func(ctx context.Context) error {
        return os.Remove("/tmp/cache-file")
})
lf.Start()
```

### Shutdown on SIGTERM:
```go
sig := lifecycle.NewSignalHandler(lf, nil)
sig.Start(lifecycle.DefaultShutdownConfig)
```

The lifecycle will shut down, on SIGTERM or interrup signals.

### Configure lifecycle behavior

```go
lf := lifecycle.New(lifecycle.Config{
        StartupTimeout: 3*time.Second, // 3 seconds timout for startup, otherwise fail
        ShutdownTimeout: 4*time.Second, // 4 seconds timeout for shutdown, otherwise fail
        StartStrategy: StartStrategyFailFast | StartStrategyRollbackOnError
})
```
There are 3 startup strategies for lifecycle manager:
 - `StartStrategyFailFast` - lifecycle manager will fail on `Start` if any service fails.
 - `StartStrategyStartAll` - lifecycle manager will continue on `Start` if one or many services fails.
 - `StartStrategyRollbackOnError` - lifecycle manager will stop all started services in case of start error.

### Report service errors

The channnel `errCh` should be used to report service errors, depens on configuration the server could be restarted:
```go
lf.RegisterStartupHook("web-service", func(ctx context.Context, errCh chan<- error) error {
	ln, err := net.Listen("tcp", ":80")
	if err != nil {
		return fmt.Errorf("listen tcp: %w", err)
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
})
```

### Configure service
```go
lf.RegisterService(types.ServiceConfig{
        Name:         "my-service",
        StartupHook:  service.Start,
        ShutdownHook: service.Stop,
        RestartPolicy: types.ServiceRestartPolicy{
                RestartOnFailure: true,
                RestartCount:     3,
                RestartDelay:     time.Millisecond * 200,
        },
})
```
Here:
 - `Name` is a name of the service
 - `StartupHook` - method to call on startup
 - `ShutdownHook` - method to call on shutdown
 - `RestartPolicy` - service restart rules:
  - `RestartOnFailure` - restart service in case of runtime errors reported to `errCh`.
  - `RestartCount` - number of restart attemts until lifecycle manager gives up.
  - `RestartDelay` - min time interval between restart attempts.

### Run HTTP web service

The package `github.com/g4s8/go-lifecycle/pkg/adaptors` contains adaptors for common services, e.g. web server:
```go
web := http.NewServeMux()
web.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("OK"))
})
svc := adaptors.NewHTTPService(&http.Server{
        Addr:    ":8080",
        Handler: web,
})
svc.RegisterLifecycle("web", lf)
```

### Add healthcheck service

The package `github.com/g4s8/go-lifecycle/pkg/health` has healthcheck http service
which can monitor service statuses in lifecycle manager:
```go
hs := health.NewService(":9999", lf)
hs.RegisterLifecycle(lf)
```
