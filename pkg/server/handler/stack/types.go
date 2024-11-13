package stack

import (
	worker "kusionstack.io/kusion/pkg/infra/util/worker"
	stackmanager "kusionstack.io/kusion/pkg/server/manager/stack"
)

func NewHandler(
	stackManager *stackmanager.StackManager,
	maxAsyncConcurrent int,
	maxAsyncBuffer int,
) (*Handler, error) {
	return &Handler{
		stackManager: stackManager,
		workerPool:   worker.NewWorkerPool(maxAsyncConcurrent, maxAsyncBuffer),
	}, nil
}

type Handler struct {
	stackManager *stackmanager.StackManager
	workerPool   *worker.WorkerPool
}

// TODO: graceful shutdown of worker pool when exiting
// Capture sigterm and sigint signals to shutdown the worker pool
func (h *Handler) Shutdown() {
	h.workerPool.Wait() // wait for all workers to finish
}
