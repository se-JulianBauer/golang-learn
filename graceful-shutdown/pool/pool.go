// Package pool provides a pool of workers that can be used to execute jobs.
//
// When the passed-in context is done, the pool shut down gracefully: it will stop accepting new jobs and will wait for the workers to finish their current jobs.
package pool

import (
	"context"
	"fmt"
	"sync"
)

// Pool is a pool of workers that can be used to execute jobs.
type Pool struct {
	ctx    context.Context
	cancel context.CancelFunc

	queuedJobs      []Job
	addJob          chan Job
	deliverJob      chan<- Job
	requestJob      chan struct{}
	requestJobCount int

	workers  []*worker
	workerWg *sync.WaitGroup
}

// NewPool creates a new pool of workers. The workers and pool will be started when Start() is called.
// ctx is the context that will be used to shut down the pool. The pool will let all workers finish their current jobs and shut down when the context is done.
// workerCount is the number of workers, and therefore the number of jobs that can be executed concurrently.
func NewPool(ctx context.Context, workerCount int) Pool {
	ctx, cancel := context.WithCancel(ctx)

	deliverJob := make(chan Job, workerCount)
	requestJob := make(chan struct{}, workerCount)

	workers := make([]*worker, workerCount)
	workerWg := sync.WaitGroup{}
	for i := 0; i < workerCount; i++ {
		workers[i] = newWorker(ctx, deliverJob, requestJob, &workerWg, i)
	}
	return Pool{
		ctx:    ctx,
		cancel: cancel,

		queuedJobs:      make([]Job, 0),
		addJob:          make(chan Job),
		deliverJob:      deliverJob,
		requestJob:      requestJob,
		requestJobCount: 0,

		workers:  workers,
		workerWg: &workerWg,
	}
}

// deliverRequestedJobs delivers as many jobs as possible to the workers.
// It will stop if there are no queued jobs or if there are no more ready workers.
func (pool *Pool) deliverRequestedJobs() {
	for i := 0; i < pool.requestJobCount; i++ {
		if len(pool.queuedJobs) == 0 {
			return
		}
		job := pool.queuedJobs[len(pool.queuedJobs)-1]
		pool.queuedJobs = pool.queuedJobs[:len(pool.queuedJobs)-1]
		pool.deliverJob <- job
		pool.requestJobCount--
	}
}

// Start starts the pool and all its workers.
func (pool *Pool) Start() {
	for _, worker := range pool.workers {
		worker.start()
	}

	go func() {
		defer func() {
			// wait for all workers to finish
			fmt.Println("pool waiting for workers to finish")
			pool.workerWg.Wait()
			// close all channels
			fmt.Println("pool closing channels")
			close(pool.requestJob)
			close(pool.deliverJob)
			close(pool.addJob)
			fmt.Println("pool done")
		}()
		for {
			select {
			case <-pool.ctx.Done():
				return
			case job := <-pool.addJob:
				pool.queuedJobs = append(pool.queuedJobs, job)
				// call deliverRequestedJobs because there might be workers waiting for jobs
				pool.deliverRequestedJobs()
			case <-pool.requestJob:
				pool.requestJobCount++
				// call deliverRequestedJobs because there might be jobs waiting for workers
				pool.deliverRequestedJobs()
			}
		}
	}()
	fmt.Println("pool started")
}

// StartNewPool is a convenience function that creates a new pool and starts it. It is equivalent to pool := NewPool(...); pool.Start().
func StartNewPool(ctx context.Context, workerCount int) Pool {
	pool := NewPool(ctx, workerCount)
	pool.Start()
	return pool
}

// AddJob adds a job to the pool. If the pool is already done, it returns a PoolDoneError.
func (pool *Pool) AddJob(job Job) {
	pool.addJob <- job
}

// AddNewJob is a convenience function that creates a new job and adds it to the pool. It is equivalent to calling AddJob(NewJob(...)).
func (pool *Pool) AddNewJob(f func()) {
	pool.AddJob(NewJob(f))
}

// Stop stops the pool and all its workers, equivalent to calling the context's cancel function.
func (pool *Pool) Stop() {
	pool.cancel()
}

// Wait waits for the pool to stop, either by from a call to Stop() or when the passed-in context is done.
func (pool *Pool) Wait() {
	pool.workerWg.Wait()
}
