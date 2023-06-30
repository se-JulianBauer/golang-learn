package pool

import (
	"context"
	"fmt"
	"sync"
)

// worker is a worker that can execute jobs.
type worker struct {
	ctx context.Context

	// deliverJob is the channel that the worker will receive jobs on.
	deliverJob <-chan Job
	// requestJob is the channel that the worker will send a signal on when it is ready to receive a job.
	requestJob chan<- struct{}

	// wg is the waitgroup that the worker will notify when it is done.
	wg *sync.WaitGroup

	// id is the id of the worker, used for debugging.
	id int
}

// newWorker creates a new worker.
// ctx is the context that will be used to shut down the worker. The worker will let its current job finish and notify the pool through the waitgroup when it is done.
func newWorker(ctx context.Context, deliverJob <-chan Job, requestJob chan<- struct{}, wg *sync.WaitGroup, id int) *worker {
	return &worker{
		ctx: ctx,

		deliverJob: deliverJob,
		requestJob: requestJob,
		wg:         wg,

		id: id,
	}
}

// start starts the worker.
// It will notify the pool through the waitgroup when it is done.
func (w *worker) start() {
	w.wg.Add(1)
	w.requestJob <- struct{}{}
	fmt.Printf("worker %d: started\n", w.id)
	go func() {
		defer func() {
			fmt.Printf("worker %d: done\n", w.id)
			w.wg.Done()
		}()
		for {
			select {
			case <-w.ctx.Done():
				return
			case job := <-w.deliverJob:
				fmt.Printf("worker %d: received job %s\n", w.id, job)
				job.run()
				// after the job is done, request a new job
				fmt.Printf("worker %d: finished with job %s\n", w.id, job)
				w.requestJob <- struct{}{}
			}
		}
	}()
}
