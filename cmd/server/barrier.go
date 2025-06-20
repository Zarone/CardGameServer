package server

import (
	"sync"
)

type Barrier struct {
	mu            sync.Mutex
	cond          *sync.Cond
	count         int
	expectedCount int
	phase         int
}

func NewBarrier(expectedCount int) *Barrier {
	b := &Barrier{
		expectedCount: expectedCount,
	}
	b.cond = sync.NewCond(&b.mu)
	return b
}

func (b *Barrier) Wait() {
	b.mu.Lock()
	defer b.mu.Unlock()

	phase := b.phase
	b.count++
	
	if b.count == b.expectedCount {
		// Last thread to arrive
		b.count = 0
		b.phase++
		b.cond.Broadcast()
	} else {
		// Wait until all threads arrive and phase changes
		for phase == b.phase {
			b.cond.Wait()
		}
	}
}
