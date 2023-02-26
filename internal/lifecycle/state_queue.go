package lifecycle

import (
	"sync"
)

type stateQueue struct {
	items []ServiceState
	mx    sync.RWMutex
}

func (q *stateQueue) check() {
	if q.items == nil {
		q.items = make([]ServiceState, 0)
	}
}

func (q *stateQueue) len() int {
	q.mx.RLock()
	defer q.mx.RUnlock()

	q.check()
	return len(q.items)
}

func (q *stateQueue) push(state ServiceState) {
	q.mx.Lock()
	defer q.mx.Unlock()

	q.check()
	q.items = append(q.items, state)
}

func (q *stateQueue) pop() (item ServiceState, ok bool) {
	q.mx.RLock()
	defer q.mx.RUnlock()

	q.check()
	if len(q.items) == 0 {
		ok = false
		return
	}
	item = q.items[0]
	q.items = q.items[1:]
	ok = true
	return
}
