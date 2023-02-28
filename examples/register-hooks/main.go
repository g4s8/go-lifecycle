package main

import (
	"context"
	"fmt"

	"github.com/g4s8/go-lifecycle/pkg/lifecycle"
)

func main() {
	lf := lifecycle.New(lifecycle.DefaultConfig)
	lf.RegisterStartupHook("hello", func(_ context.Context, _ chan<- error) error {
		fmt.Println("hello lifecycle")
		return nil
	})
	lf.RegisterShutdownHook("goodbye", func(_ context.Context) error {
		fmt.Println("goodbye lifecycle")
		return nil
	})
	if err := lf.Start(); err != nil {
		panic(err)
	}
	if err := lf.Stop(); err != nil {
		panic(err)
	}
	if err := lf.Close(); err != nil {
		panic(err)
	}
}
