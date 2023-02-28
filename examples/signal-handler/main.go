package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/g4s8/go-lifecycle/pkg/lifecycle"
	"github.com/g4s8/go-lifecycle/pkg/types"
)

type service struct {
	ticker *time.Ticker
	stopCh chan struct{}
	doneWg sync.WaitGroup
}

func (s *service) Start(ctx context.Context, errCh chan<- error) error {
	s.ticker = time.NewTicker(time.Second)
	s.stopCh = make(chan struct{})
	go func() {
		var counter int
		for {
			select {
			case <-s.ticker.C:
				fmt.Printf("tick %d\n", counter)
				counter++
			case <-s.stopCh:
				fmt.Println("stopping")
				s.doneWg.Done()
				return
			}
		}
	}()
	return nil
}

func (s *service) Stop(ctx context.Context) error {
	s.doneWg.Add(1)
	s.ticker.Stop()
	close(s.stopCh)
	s.doneWg.Wait()
	return nil
}

func main() {
	lf := lifecycle.New(lifecycle.DefaultConfig)
	var svc service
	lf.RegisterService(types.ServiceConfig{
		Name:         "service",
		StartupHook:  svc.Start,
		ShutdownHook: svc.Stop,
	})
	if err := lf.Start(); err != nil {
		panic(err)
	}
	sh := lifecycle.NewSignalHandler(lf, nil)
	sh.Start(lifecycle.DefaultShutdownConfig)
	if err := sh.Wait(); err != nil {
		panic(err)
	}
}
