package mpool

import (
	"runtime"
	"sync"
)

// Pool provides generic limited pool
type limitedPool[T any] struct {
	new     func() T
	release func(T)
	check   func(T) bool
	queue   chan T
	max     uint
	current uint
	mu      sync.Mutex
}

func NewLimitedPool[T any](initial, max uint, new func() T, release func(T), check func(T) bool) (Pool[T], error) {
	if max == 0 || initial > max || new == nil {
		return nil, ErrorInvalidParameters
	}

	pool := &limitedPool[T]{
		queue:   make(chan T, max),
		new:     new,
		release: release,
		check:   check,
		max:     max,
		current: initial,
	}

	runtime.SetFinalizer(pool, func(v *limitedPool[T]) {
		v.destroy()
	})

	for ; initial > 0; initial-- {
		pool.queue <- pool.new()
	}

	return pool, nil
}

func (pool *limitedPool[T]) Get() (T, bool) {
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
			pool.mu.Lock()
			if pool.current < pool.max {
				pool.current++
				pool.mu.Unlock()
				return pool.new(), true
			}
			pool.mu.Unlock()
			// wait for released item
			if item, ok := <-pool.queue; ok {
				return item, true
			}
			// nothing to return
			var zero T
			return zero, false
		}
	}
}

func (pool *limitedPool[T]) Put(item T) {
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

func (pool *limitedPool[T]) destroy() {
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
	pool.max = 0
	pool.current = 0
}
