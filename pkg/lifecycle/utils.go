package lifecycle

import (
	"sync"
)

type publisher[T any] struct {
	subscriptions []*subscription[T]

	last    T
	hasLast bool
	mx      sync.Mutex
}

func (p *publisher[T]) subscribe(subscriber chan<- T) Subscription {
	p.mx.Lock()
	defer p.mx.Unlock()
	sub := newSubscription(subscriber, p)
	p.subscriptions = append(p.subscriptions, sub)
	if p.hasLast {
		sub.deliver(p.last)
	}
	return sub
}

func (p *publisher[T]) unsubscribe(sub *subscription[T]) {
	p.mx.Lock()
	defer p.mx.Unlock()
	for i, s := range p.subscriptions {
		if s == sub {
			p.subscriptions = append(p.subscriptions[:i], p.subscriptions[i+1:]...)
			return
		}
	}
}

func (p *publisher[T]) publish(item T) {
	p.mx.Lock()
	defer p.mx.Unlock()
	for _, sub := range p.subscriptions {
		sub.deliver(item)
	}
	p.last = item
	p.hasLast = true
}

// Subscription for publisher.
type Subscription interface {
	// Cancel this subscription.
	Cancel()
}

type subscription[T any] struct {
	subscriber chan<- T
	cancelCh   chan struct{}
	publisher  *publisher[T]
}

func newSubscription[T any](subscriber chan<- T, publisher *publisher[T]) *subscription[T] {
	return &subscription[T]{subscriber: subscriber, cancelCh: make(chan struct{}), publisher: publisher}
}

func (s *subscription[T]) deliver(item T) {
	select {
	case <-s.cancelCh:
		return
	default:
	}
	s.subscriber <- item
}

func (s *subscription[T]) Cancel() {
	close(s.cancelCh)
	s.publisher.unsubscribe(s)
}
