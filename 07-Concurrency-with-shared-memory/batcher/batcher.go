//go:build !solution

package batcher

import (
	"sync"
	"sync/atomic"

	"gitlab.com/slon/shad-go/07-Concurrency-with-shared-memory/batcher/slow"
)

type Batcher struct {
	v         *slow.Value
	value     any
	isLoading bool
	mu        sync.Mutex
	cond      *sync.Cond
	started   atomic.Uint64
	completed uint64
}

func NewBatcher(v *slow.Value) *Batcher {
	b := Batcher{v: v}
	b.cond = sync.NewCond(&b.mu)

	return &b
}

func (b *Batcher) Load() any {
	observed := b.started.Load()

	b.mu.Lock()
	defer b.mu.Unlock()

	for observed >= b.completed {
		if b.isLoading {
			b.cond.Wait()
			continue
		}

		b.isLoading = true
		generation := b.started.Add(1)
		b.mu.Unlock()

		value := b.v.Load()

		b.mu.Lock()
		b.value = value
		b.isLoading = false
		b.completed = generation
		b.cond.Broadcast()
	}

	return b.value
}
