package mpool

import (
	"runtime"
	"sync"
)

// Pool provides generic unlimited pool
type unlimitedPool[T any] struct {
	new     func() T
	release func(T)
	check   func(T) bool
	queue   chan T
	mu      sync.Mutex
}

func NewPool[T any](initial, max uint, new func() T, release func(T), check func(T) bool) (Pool[T], error) {
	if initial > max || new == nil {
		return nil, ErrorInvalidParameters
	}

	pool := &unlimitedPool[T]{
		queue:   make(chan T, max),
		new:     new,
		release: release,
		check:   check,
	}

	runtime.SetFinalizer(pool, func(v *unlimitedPool[T]) {
		v.destroy()
	})

	for ; initial > 0; initial-- {
		pool.queue <- pool.new()
	}

	return pool, nil
}

func (pool *unlimitedPool[T]) Get() (T, bool) {
	if pool.queue == nil {
		// pool aleardy destroyed, return nothing
		var zero T
		return zero, false
	}

	for {
		select {
		case item := <-pool.queue:
			if pool.check != nil && pool.check(item) == false {
				if pool.release != nil {
					pool.release(item)
				}
				item = pool.new()
			}
			return item, true
		default:
			return pool.new(), true
		}
	}
}

func (pool *unlimitedPool[T]) Put(item T) {
	select {
	case pool.queue <- item:
		return
	default:
		// pool is full or destroyed, destroy item
		if pool.release != nil {
			pool.release(item)
		}
		return
	}
}

func (pool *unlimitedPool[T]) destroy() {
	pool.mu.Lock()
	defer pool.mu.Unlock()
	if pool.queue == nil {
		// pool is aleardy destroyed
		return
	}
	close(pool.queue)
	for item := range pool.queue {
		if pool.release != nil {
			pool.release(item)
		}
	}
	pool.queue = nil
}
