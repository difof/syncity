package syncity

import (
	"context"
)

type JobHandler func(ctx context.Context, arg any) error

type Job struct {
	handler JobHandler
	arg     any
	wr      *waitResult
}

type waitResult struct {
	ch chan struct{}
}

func newWaitResult() *waitResult {
	return &waitResult{ch: make(chan struct{})}
}

func (wr *waitResult) done() {
	close(wr.ch)
}

func (wr *waitResult) wait() {
	<-wr.ch
}

type JobQueue struct {
	jobs    chan Job
	lastErr error
}

func NewJobQueue(ctx context.Context, size int) *JobQueue {
	q := &JobQueue{
		jobs: make(chan Job, size),
	}

	go q.runner(ctx)

	return q
}

func (q *JobQueue) runner(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case job := <-q.jobs:
			q.lastErr = job.handler(ctx, job.arg)

			if job.wr != nil {
				job.wr.done()
			}
		}
	}
}

// Enqueue adds a new job to the queue
func (q *JobQueue) Enqueue(job JobHandler, arg ...any) {
	q.jobs <- Job{handler: job, arg: arg}
}

// EnqueueWait adds a new job to the queue and waits for it to complete
func (q *JobQueue) EnqueueWait(job JobHandler, arg ...any) {
	wr := newWaitResult()
	q.jobs <- Job{handler: job, arg: arg, wr: wr}
	wr.wait()
}
