package semaphore

import (
	"context"

	"golang.org/x/sync/semaphore"
)

type Semaphore struct {
	// sem is a semaphore with maximum combined weight for concurrent access
	sem *semaphore.Weighted
	// ctx is used to acquire the semaphore
	ctx context.Context
	// maxConcurrentNum is the maximum combined weight for concurrent access
	maxConcurrentNum int64
}

// New creates a new Semaphore instance
func New(maxConcurrentNum int64) *Semaphore {
	return &Semaphore{
		sem:              semaphore.NewWeighted(maxConcurrentNum),
		ctx:              context.TODO(),
		maxConcurrentNum: maxConcurrentNum,
	}
}

// Release releases a semaphore
func (s *Semaphore) Release() {
	s.sem.Release(1)
}

// Acquire acquires a semaphore, blocking until resources
// are available
func (s *Semaphore) Acquire() error {
	return s.sem.Acquire(s.ctx, 1)
}

// Wait acquires the MaxConcurrentNum semaphore, blocking until
// all resources are available
func (s *Semaphore) Wait() error {
	return s.sem.Acquire(s.ctx, s.maxConcurrentNum)
}
