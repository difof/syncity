package syncity

import (
	"context"
)

type JobHandler func(ctx context.Context, arg any) error

type Job struct {
	handler JobHandler
	arg     any
}

type JobQueue struct {
	jobs    chan Job
	lastErr error
	closer  chan struct{}
}

func NewJobQueue(ctx context.Context, size int) *JobQueue {
	q := &JobQueue{
		jobs:   make(chan Job, size),
		closer: make(chan struct{}),
	}

	go q.runner(ctx)

	return q
}

func (q *JobQueue) runner(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			q.closer <- struct{}{}
			return
		case job := <-q.jobs:
			q.lastErr = job.handler(ctx, job.arg)
		}
	}
}

// Queue adds a new job to the queue
func (q *JobQueue) Queue(job JobHandler, arg ...any) {
	q.jobs <- Job{handler: job, arg: arg}
}
