package main

import (
	"context"
	"fmt"
	"time"

	"github.com/g4s8/go-lifecycle/pkg/health"
	"github.com/g4s8/go-lifecycle/pkg/lifecycle"
	"github.com/g4s8/go-lifecycle/pkg/types"
)

func main() {
	lf := lifecycle.New(lifecycle.DefaultConfig)
	lf.RegisterService(types.ServiceConfig{
		StartupHook: func(_ context.Context, errCh chan<- error) error {
			fmt.Println("starting")
			time.Sleep(time.Second)
			go func() {
				time.Sleep(time.Second)
				errCh <- fmt.Errorf("failed")
				fmt.Println("failed")
			}()
			fmt.Println("started")
			return nil
		},
		Name: "faily",
		RestartPolicy: types.ServiceRestartPolicy{
			RestartOnFailure: true,
		},
	})
	hc := health.NewService(":9999", lf)
	hc.RegisterLifecycle(lf)

	if err := lf.Start(); err != nil {
		panic(err)
	}
	sh := lifecycle.NewSignalHandler(lf, nil)
	sh.Start(lifecycle.DefaultShutdownConfig)
	sh.Wait()
}
