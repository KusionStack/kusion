package worker

import (
	"sync"
)

type WorkerPool struct {
	tasks               chan func() // use channel to store tasks
	wg                  sync.WaitGroup
	numAvailableWorkers int        // number of available workers
	mu                  sync.Mutex // lock read/write of numAvailableWorkers
}

func NewWorkerPool(maxConcurrentGoroutines, maxBufferGoroutines int) *WorkerPool {
	pool := &WorkerPool{
		tasks:               make(chan func(), maxBufferGoroutines),
		numAvailableWorkers: maxConcurrentGoroutines, // initialize worker count
	}

	for i := 0; i < maxConcurrentGoroutines; i++ {
		go func() {
			for task := range pool.tasks {
				pool.mu.Lock()
				pool.numAvailableWorkers-- // lower worker count
				pool.mu.Unlock()

				task() // execute the task

				pool.mu.Lock()
				pool.numAvailableWorkers++ // increase worker count
				pool.mu.Unlock()
			}
		}()
	}

	return pool
}

// Do add the task to worker pool and return whether it is added to the execution zone or buffer zone
func (p *WorkerPool) Do(task func()) bool {
	p.wg.Add(1)
	inBufferZone := true

	// check available worker
	p.mu.Lock()
	if p.numAvailableWorkers > 0 {
		inBufferZone = false
	}
	p.mu.Unlock()

	// place the task into the channel
	p.tasks <- func() {
		defer p.wg.Done()
		task()
	}

	return inBufferZone
}

// Wait for all tasks before closing the channel
func (p *WorkerPool) Wait() {
	p.wg.Wait()
	close(p.tasks) // close the channel and stop the goroutines
}
