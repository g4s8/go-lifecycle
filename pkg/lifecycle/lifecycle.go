package lifecycle

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"sync"
)

type Lifecycle struct {
	items []Runner
}

func (l *Lifecycle) Add(r Runner) {
	l.items = append(l.items, r)
}

func (l *Lifecycle) Start(ctx context.Context) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	errCh := make(chan error, len(l.items))

	var wg sync.WaitGroup
	for _, item := range l.items {
		wg.Add(1)
		go func(item Runner) {
			defer wg.Done()
			if err := item.Run(ctx); err != nil {
				if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
					return
				}
				cancel()
				errCh <- err
			}
		}(item)
	}

	wg.Wait()
	close(errCh)

	var errs []error
	for err := range errCh {
		errs = append(errs, err)
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}
